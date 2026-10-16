package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

// composeNav is where the composition step stands, for the frame of its
// next form (SPEC-TUIUX-001 Stage Indicator Contract): the command, the path
// of menu choice, tab, and method taken, and the answers of the edit pass
// before the form. runForms carries it on the command context and
// runComposeForm reads it there, so composeSelect, composeConfirm, and their
// call sites in the OMP and review tabs stay as they are.
type composeNav struct {
	command string
	lang    string
	// backOut makes the menu a form like any other: init's menu goes back to
	// its wizard. Otherwise the menu is the flow's first form.
	backOut bool
	path    []string
	// summary is the answer lines of the edit pass before the next form.
	summary []string
}

type composeNavKey struct{}

// attach carries nav on cmd's context for the forms of one run; the
// returned function puts the context cmd had back.
func (n *composeNav) attach(cmd *cobra.Command) func() {
	original := cmd.Context()
	cmd.SetContext(context.WithValue(commandContext(cmd), composeNavKey{}, n))
	return func() { cmd.SetContext(original) }
}

// composeNavOf returns the position cmd's context carries, nil outside a
// form run.
func composeNavOf(cmd *cobra.Command) *composeNav {
	nav, _ := commandContext(cmd).Value(composeNavKey{}).(*composeNav)
	return nav
}

// enter adds a stage to the path; the returned function leaves it.
func (n *composeNav) enter(stage string) func() {
	depth := len(n.path)
	n.path = append(n.path[:depth], stage)
	return func() { n.path = n.path[:depth] }
}

// header is `<command> · 모델 구성[ › <stage>]…`.
func (n *composeNav) header() string {
	title := "Model composition"
	if n.lang == "ko" {
		title = "모델 구성"
	}
	return strings.Join(append([]string{n.command + " · " + title}, n.path...), " › ")
}

// frame is the next form's frame. Every line passes through terminalSafe,
// as the form's own labels do.
func (n *composeNav) frame() tui.NavFrame {
	var summary []string
	for _, line := range tui.NavSummary(n.lang, n.summary) {
		summary = append(summary, terminalSafe(line))
	}
	return tui.NavFrame{
		Header:  terminalSafe(n.header()),
		Summary: summary,
		First:   len(n.path) == 0 && !n.backOut,
		Lang:    n.lang,
	}
}

// composeFormFrame is the frame of cmd's next composition form. A form
// outside a form run, which no entry point asks, gets a first-form frame,
// so Esc never leaves a caller that has no back to take.
func composeFormFrame(cmd *cobra.Command) tui.NavFrame {
	frame := tui.NavFrame{First: true, Lang: composeLanguage()}
	if nav := composeNavOf(cmd); nav != nil {
		frame = nav.frame()
	}
	frame.Accessible = tui.NavAccessible()
	return frame
}
