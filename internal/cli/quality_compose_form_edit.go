package cli

import (
	"errors"
	"maps"
	"strings"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

// composePassKind is what one form of an edit pass asks.
type composePassKind int

const (
	composePassTier composePassKind = iota
	composePassEffort
	composePassList
	composePassValue
	composePassDone
)

// composePassForm is where a form of an edit pass stands: a role group's
// tier or effort, the agent list, or an agent's value or effort. On the list
// agent is the agent to select, the one the list picked before Esc.
type composePassForm struct {
	kind  composePassKind
	group int
	agent string
}

// composePassState is what a pass saves before each form it asks: the edited
// tiers, the tool tab's effort working copy, and the form.
type composePassState struct {
	edited  map[string]string
	efforts map[string]string
	form    composePassForm
}

// composePass is one edit pass of a method (Navigation Contract, role group
// and per-agent rows): the tiers it edits and the back stack of the forms it
// asked, which also gives each form's summary of the answers before it.
type composePass struct {
	f       *composeFormRun
	labeler composeLabeler
	edited  map[string]string
	history tui.NavHistory[composePassState]
	// again marks a form shown again after Esc, already on the stack.
	again bool
}

func (f *composeFormRun) newComposePass(labeler composeLabeler, working map[string]string) *composePass {
	return &composePass{f: f, labeler: labeler, edited: maps.Clone(working)}
}

// run asks the form at pos through ask, which applies its answer and returns
// the summary line and the form after it. A form shown for the first time
// saves the state before it. Esc restores the state saved before the form
// and returns the form asked before it; on the pass's first form it returns
// tui.ErrNavBack, and composePass drops the pass.
func (p *composePass) run(pos composePassForm, ask func() (string, composePassForm, error)) (composePassForm, error) {
	if !p.again {
		p.history.Show(composePassState{edited: maps.Clone(p.edited), efforts: p.labeler.effortWorking(), form: pos})
	}
	p.again = false
	p.f.nav.summary = p.history.Summary()
	line, next, err := ask()
	if !errors.Is(err, tui.ErrNavBack) {
		if err == nil {
			p.history.Answer(line)
		}
		return next, err
	}
	saved, ok := p.history.Back()
	if !ok {
		return pos, err
	}
	p.edited = maps.Clone(saved.edited)
	p.labeler.restoreEffort(maps.Clone(saved.efforts))
	p.again = true
	current, _ := p.history.Current()
	if current.form.kind == composePassList {
		current.form.agent = saved.form.agent
	}
	return current.form, nil
}

// line is the summary line of an answer: the stored value, or keep.
func (p *composePass) line(subject, answer string) string {
	if answer == "" {
		answer = p.f.form.Keep
	}
	return "  " + subject + ": " + answer
}

// effort asks the effort form of members at pos, or moves on to next when
// the tab asks none for them.
func (p *composePass) effort(pos composePassForm, subject string, members []string, next composePassForm) (composePassForm, error) {
	form := p.f.effortForm(p.labeler, p.edited, subject, members)
	if form == nil {
		return next, nil
	}
	return p.run(pos, func() (string, composePassForm, error) {
		level, err := form.ask(p.f)
		return p.line(form.title, level), next, err
	})
}

// groups asks one select per role group in Role Group Contract order, each
// option labelled with the cells its candidate resolves (REQ-MUX2-05,
// REQ-MUX2-06), and on a tool tab the group's effort select after it
// (SPEC-MODELUX-003 REQ-MUX3-03). Keeping a group leaves its members' tiers;
// each later group's labels already include the earlier answers.
func (f *composeFormRun) groups(working map[string]string, labeler composeLabeler) (map[string]string, error) {
	pass := f.newComposePass(labeler, working)
	defer func() { f.nav.summary = nil }()
	var err error
	for pos := (composePassForm{}); pos.group < len(composeRoleGroups); {
		group := composeRoleGroups[pos.group]
		subject := group.Label(f.shared.lang)
		next := composePassForm{kind: composePassEffort, group: pos.group}
		if pos.kind == composePassEffort {
			pos, err = pass.effort(pos, subject, group.Agents, composePassForm{group: pos.group + 1})
		} else {
			pos, err = pass.run(pos, func() (string, composePassForm, error) {
				tier, err := composeSelect(f.cmd, subject+" — "+strings.Join(group.Agents, ", "), "",
					labeler.groupOptions(pass.edited, group, f.form.Keep), "")
				if err == nil && tier != "" {
					for _, agent := range group.Agents {
						pass.edited[agent] = tier
					}
				}
				return pass.line(subject, tier), next, err
			})
		}
		if err != nil {
			return nil, err
		}
	}
	return pass.edited, nil
}

// agents lists the 17 canonical agents in Role Group Contract order behind
// 완료, asks the picked agent's value, on a tool tab its effort, and lists
// again with every label refreshed, since one agent's tier can move the
// binding and with it the other agents' cells (REQ-MUX2-08). 완료 ends the
// edit; a list answer adds no summary line.
func (f *composeFormRun) agents(working map[string]string, labeler composeLabeler) (map[string]string, error) {
	pass := f.newComposePass(labeler, working)
	defer func() { f.nav.summary = nil }()
	var err error
	for pos := (composePassForm{kind: composePassList}); pos.kind != composePassDone; {
		agent := pos.agent
		switch pos.kind {
		case composePassList:
			pos, err = pass.run(pos, func() (string, composePassForm, error) {
				options := []composeOption[string]{{f.form.ListDone, ""}}
				for _, group := range composeRoleGroups {
					for _, name := range group.Agents {
						options = append(options, composeOption[string]{labeler.agentEntry(pass.edited, name), name})
					}
				}
				picked, err := composeSelect(f.cmd, f.form.List, "", options, agent)
				if picked == "" {
					return "", composePassForm{kind: composePassDone}, err
				}
				return "", composePassForm{kind: composePassValue, agent: picked}, err
			})
		case composePassValue:
			pos, err = pass.run(pos, func() (string, composePassForm, error) {
				tier, err := composeSelect(f.cmd, agent, "", labeler.agentOptions(pass.edited, agent, f.form.Keep), "")
				if err == nil && tier != "" {
					pass.edited[agent] = tier
				}
				return pass.line(agent, tier), composePassForm{kind: composePassEffort, agent: agent}, err
			})
		default:
			pos, err = pass.effort(pos, agent, []string{agent}, composePassForm{kind: composePassList})
		}
		if err != nil {
			return nil, err
		}
	}
	return pass.edited, nil
}
