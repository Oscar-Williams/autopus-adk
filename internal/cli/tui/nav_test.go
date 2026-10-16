package tui

// SPEC-TUIUX-001 T2 (acceptance S3, S4, S20): RunNav runs a real tea.Program
// over an io.Pipe. Every listed byte sequence goes in one write: split
// writes would read an arrow's escape byte as Esc (probe P2, Open Risk R1).

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var navTestANSI = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]")

// navRun is one RunNav run on a real program over a pipe: its error once it
// returned and every view it drew, ANSI removed.
type navRun struct {
	t       *testing.T
	program *tea.Program
	feed    *io.PipeWriter
	done    chan error
	err     error
	mu      sync.Mutex
	views   []string
}

// startNavTUI starts RunNav with form and frame; size, when non-zero, is sent
// as the window size before any key.
func startNavTUI(t *testing.T, form *huh.Form, frame NavFrame, size tea.WindowSizeMsg) *navRun {
	t.Helper()
	run := &navRun{t: t, done: make(chan error, 1)}
	started := make(chan *tea.Program, 1)
	navProgramStarted = func(p *tea.Program) {
		if size.Height > 0 {
			go p.Send(size)
		}
		started <- p
	}
	navViewed = func(view string) {
		run.mu.Lock()
		defer run.mu.Unlock()
		run.views = append(run.views, navTestANSI.ReplaceAllString(view, ""))
	}
	t.Cleanup(func() { navProgramStarted, navViewed = nil, nil })
	in, feed := io.Pipe()
	run.feed = feed
	go func() { run.done <- RunNav(form, frame, in, io.Discard) }()
	run.program = <-started
	t.Cleanup(func() { _ = feed.Close() })
	return run
}

// write sends one write after the views of the write before it settled.
func (r *navRun) write(keys string) {
	r.t.Helper()
	time.Sleep(150 * time.Millisecond)
	_, err := r.feed.Write([]byte(keys))
	require.NoError(r.t, err)
}

// running reports whether RunNav is still running a moment later.
func (r *navRun) running() bool {
	select {
	case r.err = <-r.done:
		return false
	case <-time.After(300 * time.Millisecond):
		return true
	}
}

// wait returns RunNav's error; a run that does not return is killed.
func (r *navRun) wait() error {
	r.t.Helper()
	select {
	case r.err = <-r.done:
	case <-time.After(3 * time.Second):
		r.program.Kill()
		r.t.Fatalf("RunNav did not return; last view:\n%s", r.lastView())
	}
	return r.err
}

func (r *navRun) seen() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.views...)
}

// lastView is the last non-empty view drawn so far.
func (r *navRun) lastView() string {
	views := r.seen()
	for i := len(views) - 1; i >= 0; i-- {
		if views[i] != "" {
			return views[i]
		}
	}
	return ""
}

// runNavTUI runs form with frame over writes and returns once RunNav did.
func runNavTUI(t *testing.T, form *huh.Form, frame NavFrame, size tea.WindowSizeMsg, writes ...string) *navRun {
	t.Helper()
	run := startNavTUI(t, form, frame, size)
	for _, keys := range writes {
		run.write(keys)
	}
	run.wait()
	return run
}

func navSelect(value *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().
		Options(huh.NewOption("A", "a"), huh.NewOption("B", "b"), huh.NewOption("C", "c")).Value(value)))
}

func navLines(view string) []string {
	return strings.Split(view, "\n")
}

// S4: Esc goes back, Ctrl+C aborts, Enter completes, a one-write arrow moves,
// the filter keeps its Esc, and a Confirm goes back on Esc too.
func TestRunNav_S4KeysOfTheNavigationContract(t *testing.T) {
	frame := NavFrame{Header: "h", Summary: []string{"  s: 1"}, Lang: "ko"}
	size := tea.WindowSizeMsg{Width: 80, Height: 24}
	cases := []struct {
		name   string
		writes []string
		err    error
		value  string
	}{
		{"(1) esc", []string{"\x1b"}, ErrNavBack, "b"},
		{"(2) ctrl+c", []string{"\x03"}, huh.ErrUserAborted, "b"},
		{"(3) enter", []string{"\r"}, nil, "b"},
		{"(4) down then enter", []string{"\x1b[B", "\r"}, nil, "c"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			value := "b"
			run := runNavTUI(t, navSelect(&value), frame, size, tc.writes...)

			assert.ErrorIs(t, run.err, tc.err)
			if tc.err == nil {
				assert.NoError(t, run.err)
			}
			assert.Equal(t, tc.value, value)
		})
	}
	t.Run("(5) filter keeps the first esc", func(t *testing.T) {
		value := "b"
		run := startNavTUI(t, navSelect(&value), frame, size)
		for _, keys := range []string{"/", "c", "\x1b"} {
			run.write(keys)
		}
		require.True(t, run.running(), "the esc that ends the filter input is huh's")
		run.write("\x1b")
		assert.ErrorIs(t, run.wait(), ErrNavBack)
	})
	t.Run("(6) confirm esc", func(t *testing.T) {
		confirmed := true
		form := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("ok?").Value(&confirmed)))
		run := runNavTUI(t, form, frame, size, "\x1b")
		assert.ErrorIs(t, run.err, ErrNavBack)
		assert.True(t, confirmed)
	})
	t.Run("view", func(t *testing.T) {
		value := "b"
		run := runNavTUI(t, navSelect(&value), frame, size, "\r")
		require.NoError(t, run.err)
		lines := navLines(run.seen()[0])
		require.GreaterOrEqual(t, len(lines), 3)
		assert.Equal(t, "h", lines[0])
		assert.Equal(t, "  s: 1", lines[1])
		assert.Equal(t, "방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소", lines[len(lines)-1])
		for _, view := range run.seen() {
			assert.NotContains(t, view, "enter submit")
			assert.NotContains(t, view, "filter")
		}
	})
}

// S3, TUI part: Esc on a first form keeps it open with the notice in the
// footer's place under the header, and the next Enter completes it.
func TestRunNav_S3FirstFormStaysWithTheNotice(t *testing.T) {
	value := "b"
	frame := NavFrame{Header: "init · 1/6 사용 프로필", First: true, Lang: "ko"}
	run := startNavTUI(t, navSelect(&value), frame, tea.WindowSizeMsg{})
	run.write("\x1b")
	require.True(t, run.running())
	lines := navLines(run.lastView())
	assert.Equal(t, "init · 1/6 사용 프로필", lines[0])
	assert.Equal(t, "첫 단계입니다. 취소하려면 Ctrl+C를 누르세요.", lines[len(lines)-1])
	run.write("\r")

	require.NoError(t, run.wait())
	assert.Equal(t, "b", value)
	first := navLines(run.seen()[0])
	assert.Equal(t, "방향키 이동 · Enter 선택 · Ctrl+C 취소", first[len(first)-1], "a first form has no back hint")
}

// navS20Summary is the summary of S20: the head line and answers a1 to a30.
func navS20Summary(lang string) []string {
	var answers []string
	for i := 1; i <= 30; i++ {
		answers = append(answers, fmt.Sprintf("  a%d: v", i))
	}
	return NavSummary(lang, answers)
}

func navOptions(n int) []huh.Option[int] {
	options := make([]huh.Option[int], n)
	for i := range options {
		options[i] = huh.NewOption(fmt.Sprintf("option %d", i+1), i)
	}
	return options
}

func navIntSelect(title string, n int) *huh.Form {
	value := 0
	field := huh.NewSelect[int]().Options(navOptions(n)...).Value(&value)
	if title != "" {
		field = field.Title(title)
	}
	return huh.NewForm(huh.NewGroup(field))
}

// navFormHeight measures the form RunNav draws inside height H: its twin,
// sized to H-2 the way RunNav sizes it.
func navFormHeight(form *huh.Form, height int) int {
	form.WithShowHelp(false).Init()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: height - 2})
	return lipgloss.Height(form.View())
}

// S20: the view fits the height, header first and footer last; the summary
// keeps its head line, the hidden-answers line, and the last answers, or
// goes when fewer than three lines are left. Accessible output keeps it all.
func TestRunNav_S20FitsTheTerminalHeight(t *testing.T) {
	for _, tc := range []struct{ lang, more, footer string }{
		{"ko", "  … 외 %d개", "방향키 이동 · Enter 선택 · Esc 뒤로 · Ctrl+C 취소"},
		{"en", "  … %d more", "Arrows move · Enter select · Esc back · Ctrl+C cancel"},
	} {
		t.Run(tc.lang, func(t *testing.T) {
			frame := NavFrame{Header: "h", Summary: navS20Summary(tc.lang), Lang: tc.lang}

			run := runNavTUI(t, navIntSelect("pick", 3), frame, tea.WindowSizeMsg{Width: 80, Height: 24}, "\r")
			require.NoError(t, run.err)
			lines := navLines(run.lastView())
			h := navFormHeight(navIntSelect("pick", 3), 24)
			avail := 22 - h
			assert.LessOrEqual(t, len(lines), 24)
			assert.Equal(t, "h", lines[0])
			assert.Equal(t, tc.footer, lines[len(lines)-1])
			want := append([]string{frame.Summary[0], fmt.Sprintf(tc.more, 30-(avail-2))}, frame.Summary[len(frame.Summary)-(avail-2):]...)
			assert.Equal(t, want, lines[1:1+avail])
			assert.Equal(t, "  a30: v", lines[avail])

			run = runNavTUI(t, navIntSelect("", 18), frame, tea.WindowSizeMsg{Width: 80, Height: 10}, "\r")
			require.NoError(t, run.err)
			lines = navLines(run.lastView())
			assert.LessOrEqual(t, len(lines), 10)
			assert.Equal(t, "h", lines[0])
			assert.Equal(t, tc.footer, lines[len(lines)-1])
			assert.NotContains(t, lines, frame.Summary[0])
			assert.LessOrEqual(t, len(lines)-2, 8, "the form fits in H-2")

			var out strings.Builder
			require.NoError(t, RunNavAccessible(navIntSelect("pick", 3), frame, strings.NewReader("\n"), &out))
			for _, line := range frame.Summary {
				assert.Contains(t, out.String(), line+"\n")
			}
			assert.NotContains(t, out.String(), "외")
			assert.NotContains(t, out.String(), "more")
		})
	}
}
