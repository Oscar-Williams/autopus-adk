package cli

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// reviewWriteRefusal is a review edit the line splice writer cannot make
// safely (Line Splice Writer Contract). Nothing is written; key is the first
// refused path, which the write-refused line names.
type reviewWriteRefusal struct{ key string }

func (e reviewWriteRefusal) Error() string {
	return "refuse to edit " + e.key + " in autopus.yaml: the target is not a plain one-line value"
}

// composeReviewReloadError is a review write the run could not read back.
// Like composeReloadError, which it unwraps to, it ends the run: the run's
// config still holds the previous values, and a later save would write them
// back.
type composeReviewReloadError struct{ err error }

func (e composeReviewReloadError) Error() string {
	return "review settings were applied but could not be re-read: " + e.err.Error()
}

func (e composeReviewReloadError) Unwrap() error { return composeReloadError(e) }

// reviewTarget is one key an edit writes (Config Write Contract): the
// providers list, or one scalar.
type reviewTarget struct {
	path   []string
	list   []string
	value  string
	isList bool
}

func (t reviewTarget) key() string { return strings.Join(t.path, ".") }

// reviewTargets lists the keys whose working value differs from the
// baseline, in change line order: spec.review_gate.providers, each changed
// orchestra.providers.<p>.model by provider, then spec.review_gate.judge.
// Nothing equal to the baseline is written.
func reviewTargets(baseline, working reviewComposition) []reviewTarget {
	var targets []reviewTarget
	if !slices.Equal(working.Providers, baseline.Providers) {
		targets = append(targets, reviewTarget{
			path: []string{"spec", "review_gate", "providers"}, list: slices.Clone(working.Providers), isList: true,
		})
	}
	for _, name := range slices.Sorted(maps.Keys(working.Models)) {
		if working.Models[name] != baseline.Models[name] {
			targets = append(targets, reviewTarget{path: []string{"orchestra", "providers", name, "model"}, value: working.Models[name]})
		}
	}
	if working.Judge != baseline.Judge {
		targets = append(targets, reviewTarget{path: []string{"spec", "review_gate", "judge"}, value: working.Judge})
	}
	return targets
}

// saveReviewComposition writes working to dir's autopus.yaml for update and
// quality (REQ-RVM-12): the bytes after pruneRetiredConfig, with only the
// planned value spans replaced, checked by verifyReviewLineEdit and written
// in one atomic replace that keeps the file mode. A refusal or failure
// leaves the file and cfg as they were; after the write cfg takes the
// orchestra and review gate of the re-read file, so no later save of the
// run writes the previous values back.
func saveReviewComposition(dir string, cfg *config.HarnessConfig, baseline, working reviewComposition) error {
	targets := reviewTargets(baseline, working)
	if len(targets) == 0 {
		return nil
	}
	for _, target := range targets {
		if target.path[0] == "orchestra" {
			if _, _, ok := orchestra.SplitModelSelector(target.value); !ok {
				return reviewWriteRefusal{key: target.key()}
			}
		}
	}
	path := filepath.Join(dir, "autopus.yaml")
	original, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	source, _, err := pruneRetiredConfig(original)
	if err != nil {
		return err
	}
	edited, err := spliceReviewTargets(source, targets)
	if err != nil {
		return err
	}
	if err := verifyReviewLineEdit(source, edited, targets); err != nil {
		return err
	}
	if err := atomicWriteQualityConfig(path, edited); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	reloaded, err := composeReloadConfig(dir)
	if err != nil {
		return composeReviewReloadError{err: err}
	}
	cfg.Orchestra = reloaded.Orchestra
	cfg.Spec.ReviewGate = reloaded.Spec.ReviewGate
	return nil
}

// applyReviewCompositionMemory applies working to init's in-memory config,
// which init's save writes (REQ-RVM-13). The candidate is validated before
// anything is set, so a refused edit changes nothing.
func applyReviewCompositionMemory(cfg *config.HarnessConfig, baseline, working reviewComposition) error {
	candidate := *cfg
	candidate.Orchestra.Providers = maps.Clone(cfg.Orchestra.Providers)
	candidate.Spec.ReviewGate.Providers = slices.Clone(working.Providers)
	candidate.Spec.ReviewGate.Judge = working.Judge
	for name, selector := range working.Models {
		if selector == baseline.Models[name] {
			continue
		}
		entry := candidate.Orchestra.Providers[name]
		entry.Model = selector
		candidate.Orchestra.Providers[name] = entry
	}
	if err := candidate.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	cfg.Orchestra.Providers = candidate.Orchestra.Providers
	cfg.Spec.ReviewGate = candidate.Spec.ReviewGate
	return nil
}
