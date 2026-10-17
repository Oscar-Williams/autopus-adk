package cli

// The first frame of a composition select (v0.50.126 smoke): composeSelect
// runs on the default form runner, RunNav's real tea.Program reading an
// io.Pipe, at RunNav's height before any window size. Every option fits that
// frame, so it shows them all with the cursor on the default.

import (
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var composeFrameANSI = regexp.MustCompile("\x1b\\[[0-9;?]*[a-zA-Z]")

// composeFrameOutput is the terminal RunNav draws on: bubbletea writes each
// frame in one Write, so the output read once a title shows holds the whole
// frame.
type composeFrameOutput struct {
	mu  sync.Mutex
	out strings.Builder
}

func (o *composeFrameOutput) Write(p []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.out.Write(p)
}

func (o *composeFrameOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.out.String()
}

// composeSelectFirstRows runs ask on a command reading a pipe and returns
// the option rows drawn before any key: the select's bordered lines after
// its title, ANSI, border, and trailing blanks removed. Enter then answers
// the select, and ask must return nil.
func composeSelectFirstRows(t *testing.T, title string, ask func(cmd *cobra.Command) error) []string {
	t.Helper()
	t.Setenv("TERM", "xterm-256color")
	in, feed := io.Pipe()
	t.Cleanup(func() { _ = feed.Close() })
	out := &composeFrameOutput{}
	cmd := &cobra.Command{}
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(io.Discard)
	done := make(chan error, 1)
	go func() { done <- ask(cmd) }()
	deadline := time.After(3 * time.Second)
	for !strings.Contains(out.String(), title) {
		select {
		case err := <-done:
			t.Fatalf("the select returned before its first frame: %v", err)
		case <-deadline:
			t.Fatalf("no frame showed %q; output:\n%s", title, out.String())
		case <-time.After(time.Millisecond):
		}
	}
	var rows []string
	for _, line := range strings.Split(composeFrameANSI.ReplaceAllString(out.String(), ""), "\n") {
		row, ok := strings.CutPrefix(strings.Trim(line, "\r"), "┃")
		if ok && !strings.Contains(row, title) {
			rows = append(rows, strings.TrimRight(strings.TrimPrefix(row, " "), " "))
		}
	}
	_, err := feed.Write([]byte("\r"))
	require.NoError(t, err)
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("the select did not return after Enter")
	}
	return rows
}

func TestComposeSelect_FirstFrameShowsEveryOptionWithTheCursorOnTheDefault(t *testing.T) {
	t.Run("participation of a provider that takes part", func(t *testing.T) {
		// huh's zero value false is the second option: before the fix the
		// frame drew `  끄기` and a blank row.
		answer := false
		rows := composeSelectFirstRows(t, "claude 리뷰 참여", func(cmd *cobra.Command) error {
			var err error
			answer, err = composeSelect(cmd, "claude 리뷰 참여", "", []composeOption[bool]{
				{"그대로 — 켜짐", true}, {"끄기", false},
			}, true)
			return err
		})
		assert.Equal(t, []string{"> 그대로 — 켜짐", "  끄기"}, rows)
		assert.True(t, answer, "Enter keeps the default")
	})
	t.Run("a default below the first option", func(t *testing.T) {
		answer := 0
		rows := composeSelectFirstRows(t, "pick", func(cmd *cobra.Command) error {
			var err error
			answer, err = composeSelect(cmd, "pick", "", []composeOption[int]{{"a", 0}, {"b", 1}, {"c", 2}}, 2)
			return err
		})
		assert.Equal(t, []string{"  a", "  b", "> c"}, rows)
		assert.Equal(t, 2, answer)
	})
}
