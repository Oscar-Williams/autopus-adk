package tui

// The RunNav test harness: a real tea.Program over an io.Pipe, synchronized
// on the views it draws instead of fixed sleeps.

import (
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/require"
)

var navTestANSI = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]")

// navRun is one RunNav run on a real program over a pipe: its error once it
// returned and every view it drew, ANSI removed. A write goes in once the
// program drew a view after the write before it, or ended: bubbletea draws a
// view after every message it handles, so no test waits on a fixed sleep.
type navRun struct {
	t       *testing.T
	program *tea.Program
	feed    *io.PipeWriter
	done    chan error
	err     error
	ended   bool
	mu      sync.Mutex
	views   []string
	// drawn is the number of views drawn when the last write went in.
	drawn int
}

// startNavTUI starts RunNav with form and frame; size, when non-zero, is
// handed to the program before any key.
func startNavTUI(t *testing.T, form *huh.Form, frame NavFrame, size tea.WindowSizeMsg) *navRun {
	t.Helper()
	run := &navRun{t: t, done: make(chan error, 1)}
	started := make(chan *tea.Program, 1)
	navProgramStarted = func(p *tea.Program) { started <- p }
	navViewed = func(view string) {
		run.mu.Lock()
		defer run.mu.Unlock()
		run.views = append(run.views, navTestANSI.ReplaceAllString(view, ""))
	}
	t.Cleanup(func() { navProgramStarted, navViewed = nil, nil })
	in, feed := io.Pipe()
	run.feed = feed
	t.Cleanup(func() { _ = feed.Close() })
	go func() { run.done <- RunNav(form, frame, in, io.Discard) }()
	run.program = <-started
	if size.Height > 0 {
		// Send returns once the event loop took the message, so it is handled
		// before the first key, which the loop reads after it.
		run.program.Send(size)
	}
	return run
}

// settle waits until the program drew a view after the last write, or ended.
func (r *navRun) settle() {
	r.t.Helper()
	deadline := time.After(3 * time.Second)
	for !r.ended && r.count() <= r.drawn {
		select {
		case r.err = <-r.done:
			r.ended = true
		case <-deadline:
			r.program.Kill()
			r.t.Fatalf("RunNav drew nothing after the last write; last view:\n%s", r.lastView())
		case <-time.After(time.Millisecond):
		}
	}
}

// waitView waits until the program drew a view that want accepts and
// returns it; the views drawn before a key is handled never match.
func (r *navRun) waitView(want func(string) bool) string {
	r.t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		if view := r.lastView(); want(view) {
			return view
		}
		select {
		case <-deadline:
			r.program.Kill()
			r.t.Fatalf("RunNav drew no expected view; last view:\n%s", r.lastView())
		case <-time.After(time.Millisecond):
		}
	}
}

// write sends one write once the write before it was drawn.
func (r *navRun) write(keys string) {
	r.t.Helper()
	r.settle()
	require.False(r.t, r.ended, "RunNav returned before the write %q", keys)
	r.drawn = r.count()
	wrote := make(chan error, 1)
	go func() {
		_, err := r.feed.Write([]byte(keys))
		wrote <- err
	}()
	select {
	case err := <-wrote:
		require.NoError(r.t, err)
	case r.err = <-r.done:
		r.ended = true
	case <-time.After(3 * time.Second):
		r.program.Kill()
		r.t.Fatalf("RunNav did not read %q", keys)
	}
}

// wait returns RunNav's error; a run that does not return is killed.
func (r *navRun) wait() error {
	r.t.Helper()
	if r.ended {
		return r.err
	}
	select {
	case r.err = <-r.done:
		r.ended = true
	case <-time.After(3 * time.Second):
		r.program.Kill()
		r.t.Fatalf("RunNav did not return; last view:\n%s", r.lastView())
	}
	return r.err
}

func (r *navRun) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.views)
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
