package tui

// SPEC-TUIUX-001 Height Contract, v0.50.126 smoke: every frame RunNav draws
// before the first key shows a select's cursor row. RunNav runs a real
// tea.Program over an io.Pipe (nav_harness_test.go) at the sizes of the
// smoke, before any size, and in a frame too short for the form, where huh
// scrolls the select to its cursor row.

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const navSelectHeader = "update · 모델 구성 › 구성 편집 › 멀티프로바이더 리뷰"

// navSelectRows is the option rows of the focused select in view: the
// bordered lines after its title, border and trailing blanks removed.
func navSelectRows(view string) []string {
	var rows []string
	for i, line := range navLines(view) {
		if row, ok := strings.CutPrefix(line, "┃"); ok && i > 1 {
			rows = append(rows, strings.TrimRight(strings.TrimPrefix(row, " "), " "))
		}
	}
	return rows
}

// assertNavFirstFrames runs field as one form at size and checks the frames
// drawn before any key, up to the first one at size: each keeps the header,
// the footer, and the cursor row rows[cursor]; the one at size shows rows,
// or only the cursor row when short. Enter then answers the select.
func assertNavFirstFrames(t *testing.T, field huh.Field, size tea.WindowSizeMsg, short bool, rows []string, cursor int) {
	t.Helper()
	width := size.Width
	if width == 0 {
		width = navDefaultWidth
	}
	run := startNavTUI(t, huh.NewForm(huh.NewGroup(field)), NavFrame{Header: navSelectHeader, Lang: "ko"}, size)
	// The form's lines take the width of the last size it got.
	last := run.waitView(func(view string) bool {
		lines := navLines(view)
		return len(lines) > 1 && lipgloss.Width(lines[1]) == width
	})
	for _, frame := range run.seen() {
		lines := navLines(frame)
		assert.Equal(t, navSelectHeader, lines[0])
		assert.Equal(t, navTextKO.footer, lines[len(lines)-1])
		assert.Contains(t, navSelectRows(frame), rows[cursor], "a frame without the cursor row:\n%s", frame)
	}
	if short {
		rows = rows[cursor : cursor+1]
	}
	assert.Equal(t, rows, navSelectRows(last), "the frame at the size:\n%s", last)
	if size.Height > 0 {
		assert.LessOrEqual(t, len(navLines(last)), size.Height)
	}
	run.write("\r")
	require.NoError(t, run.wait())
}

func TestNavSelect_EveryFrameBeforeAKeyShowsTheCursorRow(t *testing.T) {
	for _, size := range []struct {
		name  string
		size  tea.WindowSizeMsg
		short bool
	}{
		{"120x40", tea.WindowSizeMsg{Width: 120, Height: 40}, false},
		{"120x60", tea.WindowSizeMsg{Width: 120, Height: 60}, false},
		{"before any size", tea.WindowSizeMsg{}, false},
		// The form gets H-2 = 2 rows: its title and one option row.
		{"120x4", tea.WindowSizeMsg{Width: 120, Height: 4}, true},
	} {
		t.Run(size.name+"/participation listing true first", func(t *testing.T) {
			// huh's zero value false is the second option.
			value := true
			options := []huh.Option[bool]{huh.NewOption("그대로 — 켜짐", true), huh.NewOption("끄기", false)}
			field := NavSelect(options, &value).Title("claude 리뷰 참여")
			assertNavFirstFrames(t, field, size.size, size.short, []string{"> 그대로 — 켜짐", "  끄기"}, 0)
			assert.True(t, value, "Enter keeps the default")
		})
		t.Run(size.name+"/a default below the first option", func(t *testing.T) {
			value := 2
			field := NavSelect(navOptions(3), &value).Title("pick")
			assertNavFirstFrames(t, field, size.size, size.short, []string{"  option 1", "  option 2", "> option 3"}, 2)
			assert.Equal(t, 2, value, "Enter keeps the default")
		})
	}
}
