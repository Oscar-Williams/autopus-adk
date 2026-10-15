package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// reviewComposition is one value of the review tab (SPEC-REVMODEL-001
// Review Tab Contract): spec.review_gate.providers, the model selector of
// every OMP-backed provider, and spec.review_gate.judge.
type reviewComposition struct {
	Providers []string
	Models    map[string]string
	Judge     string
}

func (c reviewComposition) clone() reviewComposition {
	return reviewComposition{Providers: slices.Clone(c.Providers), Models: maps.Clone(c.Models), Judge: c.Judge}
}

// takesPart reports a provider that spec.review_gate.providers names.
func (c reviewComposition) takesPart(name string) bool { return slices.Contains(c.Providers, name) }

// reviewBaseline is the tab's start, taken from the run's config when the
// tab opens (REQ-RVM-02). Change lines, the no-change check, and every write
// compare against it; the working value only seeds the selects.
type reviewBaseline struct {
	reviewComposition
	// Visible lists the orchestra.providers entries in sortedProviderKeys
	// order: only they get a participation select and count.
	Visible []string
	// OMP holds the providers whose entry has backend: omp.
	OMP map[string]bool
}

func newReviewBaseline(cfg *config.HarnessConfig) reviewBaseline {
	baseline := reviewBaseline{
		reviewComposition: reviewComposition{
			Providers: slices.Clone(cfg.Spec.ReviewGate.Providers),
			Models:    map[string]string{},
			Judge:     cfg.Spec.ReviewGate.Judge,
		},
		Visible: sortedProviderKeys(cfg.Orchestra.Providers),
		OMP:     map[string]bool{},
	}
	for name, entry := range cfg.Orchestra.Providers {
		if entry.Backend == config.ProviderBackendOMP {
			baseline.OMP[name] = true
			baseline.Models[name] = entry.Model
		}
	}
	return baseline
}

// participation is the spec.review_gate.providers list for the visible
// providers on names (REQ-RVM-03): the baseline list without the visible
// providers turned off, then the ones turned on in sortedProviderKeys order.
// A listed name without an orchestra.providers entry is neither shown nor
// changed.
func (b reviewBaseline) participation(on map[string]bool) []string {
	list := make([]string, 0, len(b.Providers)+len(b.Visible))
	for _, name := range b.Providers {
		if on[name] || !slices.Contains(b.Visible, name) {
			list = append(list, name)
		}
	}
	for _, name := range b.Visible {
		if on[name] && !slices.Contains(b.Providers, name) {
			list = append(list, name)
		}
	}
	return list
}

// participants counts the visible providers value takes part in; the count
// check counts nothing else (REQ-RVM-05).
func (b reviewBaseline) participants(value reviewComposition) int {
	count := 0
	for _, name := range b.Visible {
		if value.takesPart(name) {
			count++
		}
	}
	return count
}

// reviewAttestation is the attested model list of each OMP-backed provider
// (REQ-RVM-06). A provider without a list, such as one read from an
// unreadable catalog, attests no model.
type reviewAttestation map[string][]ompCustomModelChoice

// reviewPlanError is a validation error; its text is the Form Text line.
type reviewPlanError string

func (e reviewPlanError) Error() string { return string(e) }

// reviewPlan is a validated edit described against the baseline (REQ-RVM-09).
type reviewPlan struct {
	Changed bool
	// Changes are the change lines in Review Tab Contract order, with raw
	// names; the tab prints them through terminalSafe.
	Changes []string
	Notes   []string
	// Table is the after table, its cells terminal-safe.
	Table composeTableLayout
}

// planReviewComposition validates working against the baseline (REQ-RVM-05)
// and describes it: fewer than two participating visible providers, a judge
// that takes no part, or a changed model the provider's attested list does
// not hold, with a level that model reports, is an error. A model equal to
// the baseline is not held against the catalog.
func planReviewComposition(b reviewBaseline, working reviewComposition, attested reviewAttestation, text composeReviewText) (reviewPlan, error) {
	if n := b.participants(working); n < 2 {
		return reviewPlan{}, reviewPlanError(fmt.Sprintf(text.ErrCount, n))
	}
	if working.Judge != "" && !working.takesPart(working.Judge) {
		return reviewPlan{}, reviewPlanError(fmt.Sprintf(text.ErrJudge, working.Judge))
	}
	for _, name := range b.Visible {
		if selector := working.Models[name]; selector != b.Models[name] {
			if err := attestReviewModel(name, selector, attested[name], text); err != nil {
				return reviewPlan{}, err
			}
		}
	}
	plan := reviewPlan{Table: reviewAfterTable(b, working, text)}
	turnedOff, modelChanged := false, false
	if !slices.Equal(working.Providers, b.Providers) {
		plan.Changes = append(plan.Changes, "spec.review_gate.providers "+reviewListText(b.Providers)+" → "+reviewListText(working.Providers))
		turnedOff = slices.ContainsFunc(b.Visible, func(name string) bool { return b.takesPart(name) && !working.takesPart(name) })
	}
	for _, name := range b.Visible {
		if working.Models[name] != b.Models[name] {
			plan.Changes = append(plan.Changes, "orchestra.providers."+name+".model "+b.Models[name]+" → "+working.Models[name])
			modelChanged = true
		}
	}
	if working.Judge != b.Judge {
		plan.Changes = append(plan.Changes, "spec.review_gate.judge "+reviewJudgeText(b.Judge, text)+" → "+reviewJudgeText(working.Judge, text))
	}
	if turnedOff {
		plan.Notes = append(plan.Notes, text.CommandsNote)
	}
	if modelChanged {
		plan.Notes = append(plan.Notes, text.ModelNote)
	}
	plan.Changed = len(plan.Changes) > 0
	return plan, nil
}

// attestReviewModel holds a changed selector against the provider's
// attested list: its model must be listed and its level reported.
func attestReviewModel(name, selector string, choices []ompCustomModelChoice, text composeReviewText) error {
	model, thinking, ok := orchestra.SplitModelSelector(selector)
	index := slices.IndexFunc(choices, func(choice ompCustomModelChoice) bool { return choice.selector() == model })
	if !ok || index < 0 {
		return reviewPlanError(fmt.Sprintf(text.ErrModel, name, selector))
	}
	if !slices.Contains(choices[index].Thinking, thinking) {
		return reviewPlanError(fmt.Sprintf(text.ErrThinking, name, thinking, model))
	}
	return nil
}

// reviewAfterTable is the after table of the Review Tab Contract: one row
// per visible provider, its review state, its model (CLI backend for a CLI
// provider), and the judge mark, each changed cell ending in `*`.
func reviewAfterTable(b reviewBaseline, working reviewComposition, text composeReviewText) composeTableLayout {
	layout := composeTableLayout{header: []string{text.ColProvider, text.ColReview, text.ColModel, text.Judge}}
	for _, name := range b.Visible {
		on := working.takesPart(name)
		state := text.Off
		if on {
			state = text.On
		}
		model := text.CLIBackend
		if b.OMP[name] {
			model = working.Models[name]
		}
		judge := ""
		if working.Judge == name {
			judge = text.Judge
		}
		layout.rows = append(layout.rows, []string{
			name,
			reviewMarked(state, on != b.takesPart(name)),
			reviewMarked(model, working.Models[name] != b.Models[name]),
			reviewMarked(judge, (working.Judge == name) != (b.Judge == name)),
		})
	}
	return escapeComposeTableLayout(layout)
}

func reviewMarked(cell string, changed bool) string {
	switch {
	case !changed:
		return cell
	case cell == "":
		return "*"
	}
	return cell + " *"
}

// reviewListText writes a provider list as `[a, b]`.
func reviewListText(names []string) string {
	return "[" + strings.Join(names, ", ") + "]"
}

// reviewJudgeText names a judge, or judge 없음 for the empty one.
func reviewJudgeText(judge string, text composeReviewText) string {
	if judge == "" {
		return text.NoJudge
	}
	return judge
}
