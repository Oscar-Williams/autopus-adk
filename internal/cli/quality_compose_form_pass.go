package cli

import (
	"errors"
	"fmt"
	"maps"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

// composePass runs one edit pass of a method, its plan, and its
// confirmation. It returns the composition the next pass starts from and
// whether the operator keeps editing; Esc on the override question keeps
// editing with the pass's edit (Navigation Contract). tui.ErrNavBack means
// the pass's first form went back: the pass is dropped, its efforts too.
func (f *composeFormRun) composePass(planner composeDriverPlanner, labeler composeLabeler, working map[string]string, perAgent bool) (map[string]string, bool, error) {
	step := f.groups
	if perAgent {
		step = f.agents
	}
	efforts := labeler.effortWorking()
	edited, err := step(working, labeler)
	if err != nil {
		labeler.restoreEffort(efforts)
		return nil, false, err
	}
	plan, err := planner.PlanEdit(edited)
	switch {
	case errors.Is(err, tui.ErrNavBack):
		return edited, true, nil
	case errors.Is(err, errComposeInterrupted):
		return nil, false, err
	case err != nil:
		f.warn(err)
		return nil, false, nil
	case plan.NoChange:
		fmt.Fprintln(f.out, f.Text.NoChange)
		return nil, false, nil
	}
	again, err := f.review(plan)
	return plan.Composition, again, err
}

// review previews plan and asks the confirmation, whose default cancels. It
// reports whether the operator keeps editing, which Esc also does.
func (f *composeFormRun) review(plan composeDriverPlan) (bool, error) {
	if err := f.preview(plan); err != nil {
		return false, err
	}
	choice, err := composeSelect(f.cmd, f.form.Confirm, "", []composeOption[int]{
		{f.form.ConfirmCancel, 0}, {f.form.ConfirmApply, 1}, {f.form.ConfirmEdit, 2},
	}, 0)
	switch {
	case errors.Is(err, tui.ErrNavBack):
		return true, nil
	case err != nil:
		return false, err
	case choice == 1:
		f.apply(plan)
	case choice == 2:
		return true, nil
	default:
		fmt.Fprintln(f.out, f.Text.Cancelled)
	}
	return false, nil
}

// apply performs a confirmed write. The writer leaves the file and the
// config as it found them on refusal or failure (Write Safety Contract), so
// a warning line is all that is left to do.
func (f *composeFormRun) apply(plan composeDriverPlan) {
	if plan.Apply == nil {
		f.warn(errors.New("composition plan carries no write"))
		return
	}
	if err := plan.Apply(); err != nil {
		f.warn(err)
		return
	}
	f.applied = true
	fmt.Fprintln(f.out, f.Text.Applied)
}

// effortWorking copies a tool tab's effort working composition; 모든 도구
// 같게 has none.
func (l composeLabeler) effortWorking() map[string]string {
	if l.effort == nil {
		return nil
	}
	return maps.Clone(l.effort.working)
}

// restoreEffort puts back a copy effortWorking took.
func (l composeLabeler) restoreEffort(efforts map[string]string) {
	if l.effort != nil {
		l.effort.working = efforts
	}
}
