package tui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// NavFrame is what one navigation form shows around its huh form
// (SPEC-TUIUX-001 Stage Indicator Contract): a header line, the summary
// block, and the key-hint footer as the last line.
type NavFrame struct {
	// Header is the stage path line; an empty header prints no line.
	Header string
	// Summary is the block above the form, head line first; the helper
	// writes the lines as given (NavSummary builds them).
	Summary []string
	// First marks the first form of a flow: Esc and the back token show the
	// first-step notice instead of going back, and the footer has no back
	// hint.
	First bool
	// Lang picks the frame text: Korean for ko, English otherwise.
	Lang string
	// Accessible runs the form line by line, as huh does on TERM=dumb.
	Accessible bool
}

// ErrNavBack is a navigation form's answer to Esc or the accessible back
// token: the caller restores the state it saved before the form and shows
// the form the Navigation Contract names.
var ErrNavBack = errors.New("navigation: back")

// NavAccessible reports whether navigation forms run line by line. It is
// huh's own rule: NewForm turns accessible mode on when TERM is dumb.
func NavAccessible() bool {
	return os.Getenv("TERM") == "dumb"
}

// navText is the frame text of one locale.
type navText struct {
	footer, footerFirst                     string
	accessibleFooter, accessibleFooterFirst string
	notice, summaryHead, summaryMore        string // summaryMore: %d, the hidden answers
}

var (
	navTextKO = navText{
		footer:                "방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소",
		footerFirst:           "방향키 이동 · Enter 선택 · Ctrl+C 취소",
		accessibleFooter:      "답을 입력하고 Enter를 누르세요 · 뒤로: b",
		accessibleFooterFirst: "답을 입력하고 Enter를 누르세요",
		notice:                "첫 단계입니다. 취소하려면 Ctrl+C를 누르세요.",
		summaryHead:           "지금까지 고른 값",
		summaryMore:           "  … 외 %d개",
	}
	navTextEN = navText{
		footer:                "Arrows move · Enter select · Esc back · Ctrl+C cancel",
		footerFirst:           "Arrows move · Enter select · Ctrl+C cancel",
		accessibleFooter:      "Type an answer and press Enter · back: b",
		accessibleFooterFirst: "Type an answer and press Enter",
		notice:                "This is the first step; press Ctrl+C to cancel.",
		summaryHead:           "Chosen so far",
		summaryMore:           "  … %d more",
	}
)

// navTextFor returns Korean for ko and English for every other locale.
func navTextFor(lang string) navText {
	if lang == "ko" {
		return navTextKO
	}
	return navTextEN
}

// NavSummary is the summary block of answer lines: the locale's head line,
// then the lines in order. No lines make no block.
func NavSummary(lang string, lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	return append([]string{navTextFor(lang).summaryHead}, lines...)
}

// navDefaultHeight is H before the program reports a window size.
const navDefaultHeight = 24

// navDefaultWidth is the width a huh group starts with (huh NewGroup).
const navDefaultWidth = 80

// Test seams: a test sees each program RunNav starts, to send it a window
// size, and every view the program draws.
var (
	navProgramStarted func(*tea.Program)
	navViewed         func(string)
)

// RunNav runs form as one navigation form on in and out. Accessible frames
// take RunNavAccessible. Otherwise a tea model wraps the form: it draws the
// frame inside the terminal height (Height Contract), hides huh's help line,
// and maps the keys of the Navigation Contract. Esc goes back unless a
// select is taking a filter, Ctrl+C returns huh.ErrUserAborted as form.Run
// does, and a completed form returns nil.
func RunNav(form *huh.Form, frame NavFrame, in io.Reader, out io.Writer) error {
	if frame.Accessible {
		return RunNavAccessible(form, frame, in, out)
	}
	InitStyles()
	model := &navModel{form: form.WithShowHelp(false), frame: frame, text: navTextFor(frame.Lang),
		width: navDefaultWidth, height: navDefaultHeight}
	program := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out), tea.WithReportFocus())
	if navProgramStarted != nil {
		navProgramStarted(program)
	}
	_, err := program.Run()
	switch {
	case model.aborted || errors.Is(err, tea.ErrInterrupted):
		return huh.ErrUserAborted
	case err != nil:
		return fmt.Errorf("huh: %w", err)
	case model.back:
		return ErrNavBack
	}
	return nil
}

// navModel is the tea model RunNav runs: the huh form inside the frame.
type navModel struct {
	form          *huh.Form
	frame         NavFrame
	text          navText
	width, height int
	// notice shows the first-step notice in the footer's place until the
	// next key.
	notice  bool
	back    bool
	aborted bool
}

func (m *navModel) Init() tea.Cmd {
	cmd := m.form.Init()
	return tea.Batch(cmd, m.sizeForm())
}

// sizeForm hands the form the height left by the header and the footer.
func (m *navModel) sizeForm() tea.Cmd {
	model, cmd := m.form.Update(tea.WindowSizeMsg{Width: m.width, Height: max(m.height-2, 1)})
	m.form = model.(*huh.Form)
	return cmd
}

func (m *navModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, m.sizeForm()
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc && !msg.Alt && !m.filtering() {
			if m.frame.First {
				m.notice = true
				return m, nil
			}
			m.back = true
			return m, tea.Quit
		}
		m.notice = false
	}
	model, cmd := m.form.Update(msg)
	m.form = model.(*huh.Form)
	switch m.form.State {
	case huh.StateAborted:
		m.aborted = true
		return m, tea.Quit
	case huh.StateCompleted:
		return m, tea.Quit
	}
	return m, cmd
}

// filtering reports whether the focused field is a select taking a filter,
// whose Esc huh keeps: it ends the filter input.
func (m *navModel) filtering() bool {
	field, ok := m.form.GetFocusedField().(interface{ GetFiltering() bool })
	return ok && field.GetFiltering()
}

// View draws the header, the summary block cut to the height, the form, and
// the footer or the notice. A finished form draws nothing, as huh's View
// does once it quits.
func (m *navModel) View() string {
	if m.back || m.form.State != huh.StateNormal {
		return ""
	}
	form := m.form.View()
	var lines []string
	if m.frame.Header != "" {
		lines = append(lines, MutedStyle.Render(m.frame.Header))
	}
	lines = append(lines, navFitSummary(m.frame.Summary, m.height-2-lipgloss.Height(form), m.text)...)
	lines = append(lines, form, m.footer())
	view := strings.Join(lines, "\n")
	if navViewed != nil {
		navViewed(view)
	}
	return view
}

func (m *navModel) footer() string {
	switch {
	case m.notice:
		return WarningStyle.Render(m.text.notice)
	case m.frame.First:
		return MutedStyle.Render(m.text.footerFirst)
	}
	return MutedStyle.Render(m.text.footer)
}

// navFitSummary cuts the summary block, head line first, to avail lines by
// the Height Contract: all of it when it fits; with avail of three or more
// the head line, the hidden-answers line, and the last avail-2 answers;
// nothing otherwise.
func navFitSummary(summary []string, avail int, text navText) []string {
	if len(summary) <= avail {
		return summary
	}
	if avail < 3 {
		return nil
	}
	answers := summary[1:]
	shown := avail - 2
	fitted := []string{summary[0], fmt.Sprintf(text.summaryMore, len(answers)-shown)}
	return append(fitted, answers[len(answers)-shown:]...)
}
