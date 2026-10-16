package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

// reviewPlaceholderRefused prints the CLI value refusal line and reports true
// when a value the provider's model select could change is an environment
// placeholder in the raw document: the OMP model scalar, or a model or effort
// flag value. The provider then keeps every value (REQ-RVM-26); a placeholder
// in any other args element blocks nothing.
func (f *composeFormRun) reviewPlaceholderRefused(tab *reviewTabState, name string) bool {
	raw := tab.baseline.Raw[name]
	flag, value := "model", raw.Model
	if !tab.working.OMP[name] {
		flag, value = "", ""
		row := reviewArgsRowFor(tab.baseline.Binaries[name])
		for _, found := range []reviewArgValue{findReviewArg(raw.Args, row.model), reviewEffortArg(row, raw.Args)} {
			if found.found && strings.Contains(found.value, "${") {
				flag, value = found.spelling, found.value
				break
			}
		}
	}
	if !strings.Contains(value, "${") {
		return false
	}
	f.reviewLine(fmt.Sprintf(f.reviewText().CLIValueRefused, name, flag, value))
	return true
}

// reviewEffortArg is the effort flag value row reads from args.
func reviewEffortArg(row reviewArgsRow, args []string) reviewArgValue {
	switch row.kind {
	case reviewRowClaude:
		return findReviewArg(args, row.effort)
	case reviewRowCodex:
		return findReviewCodexEffort(args)
	}
	return reviewArgValue{flag: -1}
}

// reviewCLI asks a CLI provider's model by its CLI Model Contract row
// (REQ-RVM-08, 21–23): 그대로 and the row's models, then the chosen model's
// effort. A row without a model list prints the reason line instead.
func (f *composeFormRun) reviewCLI(tab *reviewTabState, name string) error {
	text := f.reviewText()
	binary := tab.baseline.Binaries[name]
	row := reviewArgsRowFor(binary)
	models := f.reviewCLIModels(row)
	tab.attested.CLI[name] = models
	if len(models) == 0 {
		f.reviewLine(fmt.Sprintf(text.NoModelList, name, binary))
		return nil
	}
	working := tab.working.CLI[name]
	options := []composeOption[string]{{f.form.Keep + " — " + reviewCLIKeepText(binary, working), ""}}
	for _, model := range models {
		options = append(options, composeOption[string]{model.ID, model.ID})
	}
	choice, err := composeSelect(f.cmd, fmt.Sprintf(text.Model, name), "", options, "")
	if err != nil || choice == "" {
		return err
	}
	model := models[slices.IndexFunc(models, func(m reviewCLIModel) bool { return m.ID == choice })]
	effort, err := f.reviewCLIEffort(name, model, reviewCLIEffortDefault(row, model, working))
	if err != nil {
		return err
	}
	tab.working.CLI[name] = reviewCLIValue{Model: model.ID, Effort: effort}
	return nil
}

// reviewCLIModels is the model list of a CLI row.
func (f *composeFormRun) reviewCLIModels(row reviewArgsRow) []reviewCLIModel {
	if row.kind == reviewRowClaude {
		return reviewClaudeModels()
	}
	return nil
}

// reviewCLIEffort asks the effort among the levels model takes, in ladder
// order; none or one level asks nothing.
func (f *composeFormRun) reviewCLIEffort(name string, model reviewCLIModel, def string) (string, error) {
	switch len(model.Efforts) {
	case 0:
		return "", nil
	case 1:
		return model.Efforts[0], nil
	}
	levels := make([]composeOption[string], 0, len(model.Efforts))
	for _, level := range model.Efforts {
		levels = append(levels, composeOption[string]{level, level})
	}
	return composeSelect(f.cmd, fmt.Sprintf(f.reviewText().Effort, name), "", levels, def)
}

// reviewCLIEffortDefault is the effort select's default: the working effort
// when model takes it, otherwise max for claude (REQ-RVM-21) and the highest
// level for codex (REQ-RVM-22).
func reviewCLIEffortDefault(row reviewArgsRow, model reviewCLIModel, working reviewCLIValue) string {
	if len(model.Efforts) == 0 || slices.Contains(model.Efforts, working.Effort) {
		return working.Effort
	}
	if row.kind == reviewRowClaude && slices.Contains(model.Efforts, config.ClaudeOrchestraEffort) {
		return config.ClaudeOrchestraEffort
	}
	return model.Efforts[len(model.Efforts)-1]
}
