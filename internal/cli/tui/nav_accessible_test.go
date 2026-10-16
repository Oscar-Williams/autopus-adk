package tui_test

// SPEC-TUIUX-001 T3 (REQ-TUX-04, REQ-TUX-06, REQ-TUX-09; acceptance S3, S21):
// the accessible helper writes the frame as plain lines and reads a back
// token as Esc, leaving the line after it for the next form.

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/charmbracelet/huh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

func abcSelect(value *string) *huh.Form {
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title("Q").
		Options(huh.NewOption("A", "a"), huh.NewOption("B", "b"), huh.NewOption("C", "c")).Value(value)))
}

// S21: an accessible frame takes the line path without a tea.Program.
func TestRunNav_S21AccessibleFrameReadsLines(t *testing.T) {
	frame := tui.NavFrame{Header: "h", Lang: "ko", Accessible: true}
	value := "b"
	var out strings.Builder
	err := tui.RunNav(abcSelect(&value), frame, strings.NewReader("b\n"), &out)
	assert.ErrorIs(t, err, tui.ErrNavBack)
	assert.Equal(t, "b", value)

	out.Reset()
	require.NoError(t, tui.RunNav(abcSelect(&value), frame, strings.NewReader("3\n"), &out))
	assert.Equal(t, "c", value)
	assert.True(t, strings.HasPrefix(out.String(), "h\n답을 입력하고 Enter를 누르세요 · 뒤로: b\nQ \n1. A\n"), out.String())
	assert.NotContains(t, out.String(), "\x1b")
}

// The token is exactly b, back, or 뒤로 after trimming and ASCII case
// folding (a Kelvin sign is no k); any other line is an answer huh reads as
// before.
func TestRunNavAccessible_BackTokens(t *testing.T) {
	for _, line := range []string{"b", "B", "back", "BACK", " Back ", "뒤로", " 뒤로\t"} {
		value := "b"
		err := tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, strings.NewReader(line+"\n"), &strings.Builder{})
		assert.ErrorIs(t, err, tui.ErrNavBack, "%q", line)
	}
	for _, line := range []string{"1", "bb", "backs", "뒤", "bac\u212a"} {
		value := "b"
		err := tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, strings.NewReader(line+"\n1\n"), &strings.Builder{})
		assert.NoError(t, err, "%q", line)
		assert.Equal(t, "a", value, "%q is not a token: huh reads it, then 1", line)
	}
}

// The line after a token stays unread for the next form, a token on a first
// form prints the notice once and the next line answers the same question,
// and the footer of a first form has no back hint.
func TestRunNavAccessible_TokenLeavesTheNextLine(t *testing.T) {
	in := iotest.OneByteReader(strings.NewReader("b\n3\n"))
	value := "b"
	var out strings.Builder
	assert.ErrorIs(t, tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, in, &out), tui.ErrNavBack)
	require.NoError(t, tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, in, &out))
	assert.Equal(t, "c", value)

	for lang, notice := range map[string]string{
		"ko": "첫 단계입니다. 취소하려면 Ctrl+C를 누르세요.",
		"en": "This is the first step; press Ctrl+C to cancel.",
	} {
		value = "b"
		out.Reset()
		frame := tui.NavFrame{Header: "h", First: true, Lang: lang}
		require.NoError(t, tui.RunNavAccessible(abcSelect(&value), frame, strings.NewReader("b\n3\n"), &out))
		assert.Equal(t, "c", value)
		assert.Equal(t, 1, strings.Count(out.String(), notice+"\n"), out.String())
		footer := map[string]string{"ko": "답을 입력하고 Enter를 누르세요", "en": "Type an answer and press Enter"}[lang]
		assert.True(t, strings.HasPrefix(out.String(), "h\n"+footer+"\nQ \n"), out.String())
	}
}

// In a form of two selects the answer before the token is already in its
// value (probe P3 group), so the caller restores its saved state; the rest
// of the form reads end of input and keeps its value.
func TestRunNavAccessible_GroupAnswerBeforeTheTokenLands(t *testing.T) {
	a, b := "en", "en"
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("first").Options(huh.NewOption("en", "en"), huh.NewOption("ko", "ko")).Value(&a),
		huh.NewSelect[string]().Title("second").Options(huh.NewOption("en", "en"), huh.NewOption("ko", "ko")).Value(&b)))
	err := tui.RunNavAccessible(form, tui.NavFrame{}, strings.NewReader("2\nb\n2\n"), &strings.Builder{})

	assert.ErrorIs(t, err, tui.ErrNavBack)
	assert.Equal(t, "ko", a)
	assert.Equal(t, "en", b)
}

func TestNavSummary_HeadLineAndNoBlockWithoutLines(t *testing.T) {
	assert.Nil(t, tui.NavSummary("ko", nil))
	assert.Equal(t, []string{"지금까지 고른 값", "  x: 1"}, tui.NavSummary("ko", []string{"  x: 1"}))
	assert.Equal(t, []string{"Chosen so far", "  x: 1"}, tui.NavSummary("ja", []string{"  x: 1"}))
}

func TestNavAccessible_FollowsTERMDumb(t *testing.T) {
	t.Setenv("TERM", "dumb")
	assert.True(t, tui.NavAccessible())
	t.Setenv("TERM", "xterm-256color")
	assert.False(t, tui.NavAccessible())
}

// An invalid answer before the back token: the token ends the current
// prompt with its default, so huh never takes the invalid line as the answer.
// A Select would otherwise index outside its options (9 → [8], x → [-1]) and
// an Input would keep the rejected text.
func TestRunNavAccessible_InvalidAnswerThenBack(t *testing.T) {
	for _, input := range []string{"9\nb\n", "x\nb\n", "0\nback\n", "4\n뒤로\n3\n"} {
		value := "b"
		var err error
		assert.NotPanics(t, func() {
			err = tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, strings.NewReader(input), &strings.Builder{})
		}, "%q", input)
		assert.ErrorIs(t, err, tui.ErrNavBack, "%q", input)
		assert.Equal(t, "b", value, "%q keeps the select's value", input)
	}
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewConfirm().Title("C").Value(&confirmed)))
	assert.ErrorIs(t, tui.RunNavAccessible(confirm, tui.NavFrame{}, strings.NewReader("x\nb\n"), &strings.Builder{}), tui.ErrNavBack)
	assert.True(t, confirmed)

	reject := func(text string) error {
		if text == "" || text == "bad" || text == "x" {
			return errors.New("rejected")
		}
		return nil
	}
	for _, input := range []string{"bad\nb\n", "x\nback\n", "\nb\n"} {
		text := "keep"
		form := huh.NewForm(huh.NewGroup(huh.NewInput().Title("I").Validate(reject).Value(&text)))
		err := tui.RunNavAccessible(form, tui.NavFrame{}, strings.NewReader(input), &strings.Builder{})
		assert.ErrorIs(t, err, tui.ErrNavBack, "%q", input)
		assert.Equal(t, "keep", text, "%q keeps the input's value", input)
	}
	// The line after the token still answers the next form.
	in := iotest.OneByteReader(strings.NewReader("9\nb\n3\n"))
	value := "b"
	assert.ErrorIs(t, tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, in, &strings.Builder{}), tui.ErrNavBack)
	require.NoError(t, tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{}, in, &strings.Builder{}))
	assert.Equal(t, "c", value)
}

// Input that ends right after an invalid answer makes huh index a Select's
// options with that answer; the helper reads it as a cancelled form, as
// Ctrl+C, instead of crashing, on first forms and later ones alike.
func TestRunNavAccessible_InvalidAnswerThenEndIsACancel(t *testing.T) {
	for _, input := range []string{"9\n", "x\n", "9\nb\n"} {
		for _, first := range []bool{false, true} {
			if input == "9\nb\n" && !first {
				continue // a later form's token is a back, not an end
			}
			value := "b"
			var err error
			assert.NotPanics(t, func() {
				err = tui.RunNavAccessible(abcSelect(&value), tui.NavFrame{First: first}, strings.NewReader(input), &strings.Builder{})
			}, "%q first=%v", input, first)
			assert.ErrorIs(t, err, huh.ErrUserAborted, "%q first=%v", input, first)
			assert.Equal(t, "b", value)
		}
	}
}
