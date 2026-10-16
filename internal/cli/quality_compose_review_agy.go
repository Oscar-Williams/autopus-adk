package cli

import (
	"context"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/insajin/autopus-adk/pkg/processprobe"
)

// The Agy Catalog Contract of the review tab (SPEC-REVMODEL-001 REQ-RVM-24).
const (
	agyModelsTimeout   = 15 * time.Second
	agyModelsMaxOutput = 64 << 10
	agyModelsHeader    = "Fetching available models..."
)

// agyModelsRunner runs `<binary> models` and returns its stdout. Tests
// replace it and restore it with t.Cleanup; no test runs a real agy.
var agyModelsRunner = runAgyModels

// runAgyModels runs `<binary> models` within agyModelsTimeout, keeping at
// most agyModelsMaxOutput of stdout; stderr is bounded and dropped. A
// timeout, a non-zero exit, and an overrun are errors.
func runAgyModels(ctx context.Context, binary string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, agyModelsTimeout)
	defer cancel()
	return processprobe.OutputLimited(exec.CommandContext(ctx, binary, "models"), agyModelsMaxOutput)
}

// agyModelLine is one listed model: an id, a tab, and a display name without
// tabs or control characters.
var agyModelLine = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*\t[^\t\x00-\x1f]+$`)

// agyModel is one line of `agy models`.
type agyModel struct {
	ID, Display string
}

// parseAgyModels reads `agy models` output strictly: an exact header line is
// skipped, and every other non-empty line must be one model. Any other line,
// an oversized output, or no model at all makes the whole list unreadable.
func parseAgyModels(out []byte) ([]agyModel, bool) {
	if len(out) > agyModelsMaxOutput {
		return nil, false
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 && lines[0] == agyModelsHeader {
		lines = lines[1:]
	}
	var models []agyModel
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !agyModelLine.MatchString(line) {
			return nil, false
		}
		id, display, _ := strings.Cut(line, "\t")
		models = append(models, agyModel{ID: id, Display: display})
	}
	return models, len(models) > 0
}

// agyModelFamily is the family an agy id names by its prefix.
func agyModelFamily(id string) string {
	switch {
	case strings.HasPrefix(id, "gemini-"):
		return "google"
	case strings.HasPrefix(id, "claude-"):
		return "anthropic"
	case strings.HasPrefix(id, "gpt-"):
		return "openai"
	}
	return ""
}

// agyCLIModels is the agy row of provider name: the listed ids of its family,
// grouped under the name without a -low, -medium, or -high suffix. The
// model select shows the groups; the value written is always a listed id.
func agyCLIModels(name string, listed []agyModel) []reviewCLIModel {
	family := providerModelFamily(name)
	var models []reviewCLIModel
	for _, entry := range listed {
		if agyModelFamily(entry.ID) != family {
			continue
		}
		model := reviewCLIModel{ID: entry.ID, Base: entry.ID, Display: entry.Display}
		for _, suffix := range []string{"low", "medium", "high"} {
			if base, ok := strings.CutSuffix(entry.ID, "-"+suffix); ok {
				model.Base, model.Suffix = base, suffix
				break
			}
		}
		models = append(models, model)
	}
	slices.SortStableFunc(models, func(a, b reviewCLIModel) int {
		if c := strings.Compare(a.Base, b.Base); c != 0 {
			return c
		}
		return strings.Compare(a.Suffix, b.Suffix)
	})
	return models
}

// reviewAgyCatalog is a tab entry's reads of `agy models`, one per binary.
type reviewAgyCatalog map[string]agyCatalogRead

type agyCatalogRead struct {
	models []agyModel
	ok     bool
}

// load runs `<binary> models` the first time the entry needs it.
func (c reviewAgyCatalog) load(ctx context.Context, binary string) ([]agyModel, bool) {
	if read, done := c[binary]; done {
		return read.models, read.ok
	}
	var read agyCatalogRead
	if out, err := agyModelsRunner(ctx, binary); err == nil {
		read.models, read.ok = parseAgyModels(out)
	}
	c[binary] = read
	return read.models, read.ok
}
