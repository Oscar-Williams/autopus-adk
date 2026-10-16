package cli

// SPEC-REVMODEL-001 T12 (REQ-RVM-25, 26; Line Splice Writer Contract for
// CLI args): every args node shape takes only its target values, a missing
// or empty args takes the migrated args with the target values changed, and
// only a target value placeholder or an aliased, anchored, or block node
// refuses.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

func reviewClaudeTo(value reviewCLIValue) func(*reviewComposition) {
	return func(w *reviewComposition) { w.CLI["claude"] = value }
}

func TestSaveReviewComposition_ClaudeArgsShapes(t *testing.T) {
	opusHigh := reviewCLIValue{"claude-opus-5-5", "high"}
	block := "            args:\n                - --print\n                - --model\n                - claude-fable-5-1\n" +
		"                - --effort\n                - max\n"
	cases := []struct {
		name  string
		args  string // replaces the RL claude args line
		value reviewCLIValue
		want  string
	}{
		{"block list", block, opusHigh, strings.Replace(strings.Replace(block, "claude-fable-5-1", "claude-opus-5-5", 1), "- max", "- high", 1)},
		{"block list drops the effort pair", block, reviewCLIValue{"claude-haiku-4-5", ""},
			"            args:\n                - --print\n                - --model\n                - claude-haiku-4-5\n"},
		{"block list appends", "            args:\n                - --print # print\n", opusHigh,
			"            args:\n                - --print # print\n                - --model\n                - claude-opus-5-5\n" +
				"                - --effort\n                - high\n"},
		{"empty flow list", "            args: []\n", opusHigh, "            args: [--print, --model, claude-opus-5-5, --effort, high]\n"},
		{"null value", "            args:\n", opusHigh, "            args: [--print, --model, claude-opus-5-5, --effort, high]\n"},
		{"flow list keeps other tokens", "            args: [--print,  --verbose, --model, 'claude-fable-5-1', --effort, max] # mine\n", opusHigh,
			"            args: [--print,  --verbose, --model, claude-opus-5-5, --effort, high] # mine\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeReviewRL(t)
			replaceReviewFile(t, dir, reviewRLClaudeArgs+"\n", tc.args)
			before := readReviewFile(t, dir)

			cfg, err := saveReviewEdit(t, dir, reviewClaudeTo(tc.value))

			require.NoError(t, err)
			assert.Equal(t, strings.Replace(before, tc.args, tc.want, 1), readReviewFile(t, dir))
			assert.Equal(t, tc.value, reviewCLIValueOf(reviewClaudeRow, cfg.Orchestra.Providers["claude"].Args))
		})
	}
}

// A claude mapping without args takes the migrated default args, with the
// target values changed, at the end of the mapping.
func TestSaveReviewComposition_MissingArgsTakeTheMigratedArgs(t *testing.T) {
	dir := writeReviewRL(t)
	replaceReviewFile(t, dir, reviewRLClaudeArgs+"\n", "")
	before := readReviewFile(t, dir)

	_, err := saveReviewEdit(t, dir, reviewClaudeTo(reviewCLIValue{"claude-sonnet-5-5", "low"}))

	require.NoError(t, err)
	assert.Equal(t, strings.Replace(before, "                timeout: 480\n        codex:\n",
		"                timeout: 480\n            args: [--print, --model, claude-sonnet-5-5, --effort, low]\n        codex:\n", 1),
		readReviewFile(t, dir))
}

func TestSaveReviewComposition_ArgsRefusals(t *testing.T) {
	t.Setenv("REVIEW_MODEL", "claude-fable-5-1")
	cases := map[string]string{
		"placeholder target": "            args: [--print, --model, \"${REVIEW_MODEL}\", --effort, max]",
		"anchored args":      "            args: &a [--print, --model, claude-fable-5-1, --effort, max]",
		"two-line flow list": "            args: [--print, --model, claude-fable-5-1,\n                --effort, max]",
		"block scalar":       "            args: |-\n                --print",
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			dir := writeReviewRL(t)
			path := filepath.Join(dir, "autopus.yaml")
			cfg, err := config.Load(dir)
			require.NoError(t, err)
			replaceReviewFile(t, dir, reviewRLClaudeArgs, args)
			before := readReviewFile(t, dir)
			baseline := newReviewBaseline(cfg, readReviewRaw(dir, cfg))
			working := baseline.reviewComposition.clone()
			working.CLI["claude"] = reviewCLIValue{"claude-opus-5-5", "max"}

			err = saveReviewComposition(dir, cfg, baseline, working)

			var refusal reviewWriteRefusal
			require.True(t, errors.As(err, &refusal), "%v", err)
			assert.Equal(t, "orchestra.providers.claude.args", refusal.key)
			data, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			assert.Equal(t, before, string(data))
		})
	}
}

// A placeholder in another element blocks nothing and stays as it is.
func TestSaveReviewComposition_OtherPlaceholdersStay(t *testing.T) {
	t.Setenv("REVIEW_TOKEN", "rvm-sentinel-7f3a9c")
	dir := writeReviewRL(t)
	args := `            args: [--print, --api-key, "${REVIEW_TOKEN}", --model, claude-fable-5-1, --effort, max]`
	replaceReviewFile(t, dir, reviewRLClaudeArgs, args)
	before := readReviewFile(t, dir)

	_, err := saveReviewEdit(t, dir, reviewClaudeTo(reviewCLIValue{"claude-opus-5-5", "max"}))

	require.NoError(t, err)
	assert.Equal(t, strings.Replace(before, "claude-fable-5-1, --effort", "claude-opus-5-5, --effort", 1), readReviewFile(t, dir))
}
