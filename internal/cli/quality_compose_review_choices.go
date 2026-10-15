package cli

import (
	"context"
	"io"
	"slices"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// reviewModelChoices is the attested list of one OMP-backed provider
// (REQ-RVM-06): every attested installed model that reports a thinking level
// and whose orchestra.ModelFamilyForSelector family is the provider's
// providerModelFamily, every attested model when that family is empty. Both
// are the functions the runtime matches provider names and models with, and
// the selector order of attested is kept.
func reviewModelChoices(name string, attested []ompCustomModelChoice) []ompCustomModelChoice {
	family := providerModelFamily(name)
	var choices []ompCustomModelChoice
	for _, choice := range attested {
		if len(choice.Thinking) > 0 && (family == "" || orchestra.ModelFamilyForSelector(choice.selector()) == family) {
			choices = append(choices, choice)
		}
	}
	return choices
}

// reviewModelOptions is the model select of one provider: 그대로 with the
// working selector, marked when the attested list lacks its model, then every
// attested model. The answer "" keeps the working model.
func reviewModelOptions(keep, working string, choices []ompCustomModelChoice, text composeReviewText) []composeOption[string] {
	label := keep + " — " + working
	if model, _, _ := orchestra.SplitModelSelector(working); !slices.ContainsFunc(choices, func(choice ompCustomModelChoice) bool {
		return choice.selector() == model
	}) {
		label += text.NotInCatalog
	}
	options := []composeOption[string]{{label, ""}}
	for _, choice := range choices {
		options = append(options, composeOption[string]{choice.selector(), choice.selector()})
	}
	return options
}

// reviewThinkingDefault is the OMP Autopus tab's thinking rule
// (quality_compose_omp_autopus.go): the working selector's level when the
// chosen model reports it, otherwise the model's first reported level.
func reviewThinkingDefault(choice ompCustomModelChoice, working string) string {
	if _, current, _ := orchestra.SplitModelSelector(working); slices.Contains(choice.Thinking, current) {
		return current
	}
	return choice.Thinking[0]
}

// reviewThinkingOptions lists the levels in the order the normalized catalog
// reports them; catalog normalization sorts them.
func reviewThinkingOptions(choice ompCustomModelChoice) []composeOption[string] {
	levels := make([]composeOption[string], 0, len(choice.Thinking))
	for _, level := range choice.Thinking {
		levels = append(levels, composeOption[string]{level, level})
	}
	return levels
}

// reviewCatalog is the one read of the installed OMP catalog a tab entry
// makes, on the run's OMP catalog runner: the first model select reads it,
// and every later select and pass of the same entry reuses the answer.
type reviewCatalog struct {
	runner  omp.OMPModelCatalogRunner
	read    bool
	choices []ompCustomModelChoice
	ok      bool
}

// load returns the attested installed models and whether the catalog could
// be read. A failed read, a catalog_invalid catalog (an explicit empty
// thinking list), and a run without a runner are all unreadable. note is the
// line a metadata-light catalog prints once.
func (c *reviewCatalog) load(ctx context.Context, out io.Writer, note string) ([]ompCustomModelChoice, bool) {
	if c.read {
		return c.choices, c.ok
	}
	c.read = true
	if c.runner == nil {
		return nil, false
	}
	choices, _, err := installedOMPAttestedModelChoices(ctx, out, c.runner, ompBuiltinDeclaredFamilies(), note)
	c.choices, c.ok = choices, err == nil
	return c.choices, c.ok
}
