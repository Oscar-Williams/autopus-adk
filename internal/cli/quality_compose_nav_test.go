package cli

// SPEC-TUIUX-001 T6 (REQ-TUX-01, 06, 08, 16, 17; acceptance S9, S21): the
// composition step's position rides on the command context from runForms to
// runComposeForm, which builds every form's frame from it.

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

func TestComposeNav_HeaderPathFirstAndSummary(t *testing.T) {
	nav := &composeNav{command: "quality", lang: "ko"}
	frame := nav.frame()
	assert.Equal(t, "quality · 모델 구성", frame.Header)
	assert.True(t, frame.First, "the menu of update and quality is a first form")
	assert.Nil(t, frame.Summary)

	leaveEdit := nav.enter("구성 편집")
	leaveTab := nav.enter("Codex")
	nav.enter("역할 그룹")
	nav.summary = []string{"  설계·판단: opus"}
	frame = nav.frame()
	assert.Equal(t, "quality · 모델 구성 › 구성 편집 › Codex › 역할 그룹", frame.Header)
	assert.False(t, frame.First)
	assert.Equal(t, []string{"지금까지 고른 값", "  설계·판단: opus"}, frame.Summary)

	leaveTab()
	assert.Equal(t, "quality · 모델 구성 › 구성 편집", nav.header(), "leaving a stage drops every stage after it")
	leaveEdit()
	assert.Equal(t, "quality · 모델 구성", nav.header())

	en := &composeNav{command: "update", lang: "en", backOut: true, path: []string{"edit the composition", "Codex", "role groups"}}
	assert.Equal(t, "update · Model composition › edit the composition › Codex › role groups", en.header())
	assert.False(t, (&composeNav{command: "init", lang: "ko", backOut: true}).frame().First, "init's menu goes back to its wizard")
	assert.Equal(t, "q · 모델 구성 › a\\x1b[31m", (&composeNav{command: "q", lang: "ko", path: []string{"a\x1b[31m"}}).frame().Header)
}

func TestComposeNav_RunFormsCarriesThePositionAndRestoresTheContext(t *testing.T) {
	type key struct{}
	cmd := &cobra.Command{Use: "quality"}
	ctx := context.WithValue(context.Background(), key{}, "kept")
	cmd.SetContext(ctx)
	nav := &composeNav{command: "quality", lang: "ko"}

	leave := nav.attach(cmd)
	assert.Same(t, nav, composeNavOf(cmd))
	assert.Equal(t, "kept", cmd.Context().Value(key{}), "the command's own values stay reachable")
	leave()
	assert.Equal(t, ctx, cmd.Context())
	assert.Nil(t, composeNavOf(cmd))

	bare := &cobra.Command{}
	leave = nav.attach(bare)
	assert.Same(t, nav, composeNavOf(bare))
	leave()
	assert.Nil(t, bare.Context())
	frame := composeFormFrame(bare)
	assert.True(t, frame.First, "a form outside a run has no back to take")
}

// S21, composition part: TERM=dumb runs every frame line by line.
func TestComposeNav_S21AccessibleFollowsTERMDumb(t *testing.T) {
	for term, accessible := range map[string]bool{"dumb": true, "xterm-256color": false} {
		t.Run(term, func(t *testing.T) {
			newTUIUXEnv(t, true, "")
			t.Setenv("TERM", term)
			dir := tuiuxFixtureD(t)
			stubTUIUXFormsProbe(t, true)
			original := composeRunForm
			t.Cleanup(func() { composeRunForm = original })
			var frames []tui.NavFrame
			composeRunForm = func(cmd *cobra.Command, form *huh.Form, frame tui.NavFrame) error {
				frames = append(frames, frame)
				return tui.RunNavAccessible(form, frame, cmd.InOrStdin(), cmd.OutOrStdout())
			}

			out, err := runQualityCompose(t, dir, formInput("\n"))

			require.NoError(t, err, out)
			require.NotEmpty(t, frames)
			for _, frame := range frames {
				assert.Equal(t, accessible, frame.Accessible)
			}
		})
	}
}

// S9, menu part: the update and quality menus carry their command's header
// and the first-form footer.
func TestComposeNav_S9MenuHeaders(t *testing.T) {
	for _, command := range []string{"quality", "update"} {
		t.Run(command, func(t *testing.T) {
			newTUIUXEnv(t, true, "")
			dir := tuiuxFixtureD(t)
			useNavComposeForms(t)
			var out string
			var err error
			if command == "quality" {
				out, err = runQualityCompose(t, dir, formInput("\n"))
			} else {
				out, err = runComposeUpdate(t, formInput("\n"), "--dir", dir)
			}
			require.NoError(t, err, out)
			assert.Contains(t, out, "\n"+command+" · 모델 구성\n답을 입력하고 Enter를 누르세요\n구성을 바꿀까요? \n")
			assert.Equal(t, 1, strings.Count(out, "답을 입력하고 Enter를 누르세요\n"))
		})
	}
}

// S15 (f): quality's form gate does not read CI, so a truthy CI still asks
// the menu with its header. The only difference is the review tab's own CI
// rule (SPEC-REVMODEL-001 REQ-RVM-14), which drops its summary line.
func TestComposeNav_S15fQualityIgnoresCI(t *testing.T) {
	run := func(ci string) string {
		newTUIUXEnv(t, true, ci)
		dir := tuiuxFixtureD(t)
		useNavComposeForms(t)
		out, err := runQualityCompose(t, dir, formInput("\n"))
		require.NoError(t, err, out)
		return strings.ReplaceAll(out, dir, "<DIR>")
	}
	withCI, without := run("true"), run("")

	assert.Contains(t, withCI, "\nquality · 모델 구성\n")
	review := regexp.MustCompile(`(?m)^멀티프로바이더 리뷰: .*\n`)
	assert.Equal(t, review.ReplaceAllString(without, ""), withCI)
}
