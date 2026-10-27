package cli

import "fmt"

// composeEffortTitle titles an effort select with its group or agent; the
// form text spells it the same in every locale (SPEC-MODELUX-003 Form Text).
const composeEffortTitle = "%s effort"

// composeEffortForm is one effort select: its options and the efforts each
// level leaves in the tab's working composition.
type composeEffortForm struct {
	tab     *composeEffortTab
	title   string
	options []composeOption[string]
	chosen  map[string]map[string]string
}

// effortForm builds the effort select of members after their tier select on
// a Claude Code or Codex tab (SPEC-MODELUX-003 REQ-MUX3-03). 그대로 comes
// first and is the default; each allowed level follows, labelled with the
// cells it leaves and `(자동)` on the automatic one. It is nil, and nothing
// is asked, on 모든 도구 같게, when no member can take an effort, or when no
// level is allowed; an unreadable Codex catalog prints its notice, once per
// tab, instead of a select.
func (f *composeFormRun) effortForm(labeler composeLabeler, tiers map[string]string, subject string, members []string) *composeEffortForm {
	tab := labeler.effort
	if tab == nil {
		return nil
	}
	agents := tab.affected(tiers, members)
	if len(agents) == 0 {
		return nil
	}
	levels, readable := tab.levels(tiers, agents)
	if !readable {
		if !tab.unreadNoted {
			fmt.Fprintln(f.out, f.form.EffortUnread)
			tab.unreadNoted = true
		}
		return nil
	}
	if len(levels) == 0 {
		return nil
	}
	form := &composeEffortForm{tab: tab, title: fmt.Sprintf(composeEffortTitle, subject), chosen: make(map[string]map[string]string, len(levels))}
	form.options = []composeOption[string]{{f.form.Keep + " — " + labeler.cells(tab.candidate(tiers, tab.working), agents), ""}}
	for _, level := range levels {
		efforts, automatic := tab.option(tiers, agents, level)
		label := labeler.cells(tab.candidate(tiers, efforts), agents)
		if automatic {
			label += " " + f.form.EffortAuto
		}
		form.options = append(form.options, composeOption[string]{label, level})
		form.chosen[level] = efforts
	}
	return form
}

// ask asks the select and keeps the chosen level's efforts.
func (e *composeEffortForm) ask(f *composeFormRun) (string, error) {
	level, err := composeSelect(f.cmd, e.title, "", e.options, "")
	if err == nil && level != "" {
		e.tab.working = e.chosen[level]
	}
	return level, err
}
