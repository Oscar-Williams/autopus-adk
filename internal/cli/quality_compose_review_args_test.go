package cli

// SPEC-REVMODEL-001 T12 (REQ-RVM-21, 25; CLI Model Contract editCLIArgs):
// the args model reads the last flag value, edits only the values its row
// names, appends a missing pair, removes claude's effort for haiku, and keeps
// every other element, its order, and its duplicates.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindReviewArg_PairsInlineAndSeparator(t *testing.T) {
	args := []string{"--print", "--model", "a", "--verbose", "--model=b", "--", "--model", "c"}

	found := findReviewArg(args, []string{"--model"})

	assert.Equal(t, reviewArgValue{found: true, index: 4, flag: -1, value: "b", prefix: "--model=", spelling: "--model"}, found)
	assert.Equal(t, "a", findReviewArg(args[:4], []string{"--model"}).value)
	assert.False(t, findReviewArg([]string{"--print", "--model"}, []string{"--model"}).found, "a flag without a value")
	assert.Equal(t, "-m", findReviewArg([]string{"exec", "-m", "x"}, reviewCodexRow.model).spelling)
}

func TestFindReviewCodexEffort_EveryForm(t *testing.T) {
	cases := map[string]struct {
		args  []string
		index int
		value string
	}{
		"pair":     {[]string{"exec", "-c", `model_reasoning_effort="max"`}, 2, `"max"`},
		"long":     {[]string{"exec", `--config=model_reasoning_effort="high"`}, 1, `"high"`},
		"bare":     {[]string{"exec", `model_reasoning_effort='low'`}, 1, `'low'`},
		"last one": {[]string{"-c", `model_reasoning_effort="low"`, "-c", `model_reasoning_effort="max"`}, 3, `"max"`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			found := findReviewCodexEffort(tc.args)
			assert.True(t, found.found)
			assert.Equal(t, tc.index, found.index)
			assert.Equal(t, tc.value, found.value)
			assert.Equal(t, reviewCodexEffortKey, found.spelling)
		})
	}
	assert.Equal(t, "max", reviewCodexEffortText(`"max"`))
	assert.False(t, findReviewCodexEffort([]string{"-c", "sandbox=read-only"}).found)
}

func TestPlanReviewArgsEdit_Claude(t *testing.T) {
	b1 := []string{"--print", "--model", "claude-fable-5-1", "--effort", "max"}
	cases := []struct {
		name string
		args []string
		want reviewCLIValue
		out  []string
	}{
		{"model and effort", b1, reviewCLIValue{"claude-opus-5-5", "xhigh"}, []string{"--print", "--model", "claude-opus-5-5", "--effort", "xhigh"}},
		{"haiku drops the effort", b1, reviewCLIValue{"claude-haiku-4-5", ""}, []string{"--print", "--model", "claude-haiku-4-5"}},
		{"missing flags are appended", []string{"--print"}, reviewCLIValue{"claude-opus-5-5", "high"},
			[]string{"--print", "--model", "claude-opus-5-5", "--effort", "high"}},
		{"inline values stay inline", []string{"--print", "--model=claude-fable-5-1", "--effort=max"}, reviewCLIValue{"claude-opus-5-5", "low"},
			[]string{"--print", "--model=claude-opus-5-5", "--effort=low"}},
		{"other elements stay", []string{"--print", "--verbose", "--model", "claude-fable-5-1", "--effort", "max", "--verbose"},
			reviewCLIValue{"claude-opus-5-5", "high"}, []string{"--print", "--verbose", "--model", "claude-opus-5-5", "--effort", "high", "--verbose"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.out, planReviewArgsEdit(tc.args, reviewClaudeRow, tc.want).apply(tc.args))
		})
	}
	unchanged := planReviewArgsEdit(b1, reviewClaudeRow, reviewCLIValue{"claude-fable-5-1", "max"})
	assert.True(t, unchanged.empty(), "values already held write nothing")
	assert.Equal(t, reviewCLIValue{"claude-fable-5-1", "max"}, reviewCLIValueOf(reviewClaudeRow, b1))
}

func TestReviewArgsRowFor_MatchesTheBinaryName(t *testing.T) {
	assert.Equal(t, reviewRowClaude, reviewArgsRowFor("/usr/local/bin/claude").kind)
	assert.Equal(t, reviewRowCodex, reviewArgsRowFor("codex.exe").kind)
	assert.Equal(t, reviewRowAgy, reviewArgsRowFor("agy").kind)
	assert.Equal(t, reviewRowOther, reviewArgsRowFor("gemini").kind)
	assert.Equal(t, "claude-fable-5-1 · max", reviewCLIKeepText("claude", reviewCLIValue{"claude-fable-5-1", "max"}))
	assert.Equal(t, "agy", reviewCLIKeepText("agy", reviewCLIValue{}))
	assert.Equal(t, "codex gpt-6-astra · max", reviewCLICellText("codex", reviewCLIValue{"gpt-6-astra", "max"}))
}
