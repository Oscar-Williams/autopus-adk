package cli

// SPEC-REVMODEL-001 T12 (REQ-RVM-05, 08, 09, 21, 26; acceptance S19, S26,
// S27): a claude CLI provider picks a Claude tab model and effort, written
// as the values after --model and --effort only; previews show only the
// changed flag values from the raw document.

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reviewRLClaudeArgs = "            args: [--print, --model, claude-fable-5-1, --effort, max]"

// reviewAssembledArgs is the argv `auto spec review` assembles for provider
// from the loader value of dir.
func reviewAssembledArgs(t *testing.T, dir, provider string) []string {
	t.Helper()
	loaded := reviewLoaderValueKeepCwd(t, dir)
	set, err := assembleSpecReviewProviders(t.Context(), specReviewProviderRequest{Config: loaded})
	require.NoError(t, err)
	for _, assembled := range set.Providers {
		if assembled.Name == provider {
			return assembled.Args
		}
	}
	t.Fatalf("no assembled provider %s", provider)
	return nil
}

func reviewArgsHasRun(args []string, run ...string) bool {
	for i := 0; i+len(run) <= len(args); i++ {
		if slices.Equal(args[i:i+len(run)], run) {
			return true
		}
	}
	return false
}

// S19: the model select, the effort select, and the args line written;
// haiku drops the effort pair and asks no effort.
func TestReviewTab_S19ClaudeCLIModelWritesArgs(t *testing.T) {
	dir := writeReviewRL(t)
	b1 := readReviewFile(t, dir)

	// participation Enter ×3, claude opus → xhigh, judge Enter, check Enter, 적용, 유지.
	run := runReviewCLIForms(t, dir, []string{"claude"}, reviewLines(reviewOpen,
		[]string{"\n", "\n", "\n", "3\n", "4\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "claude model \n1. 그대로 — claude-fable-5-1 · max\n2. claude-fable-5-1\n3. claude-opus-5-5\n"+
		"4. claude-sonnet-5-5\n5. claude-haiku-4-5\nEnter")
	assert.Contains(t, run.out, "claude effort \n1. low\n2. medium\n3. high\n4. xhigh\n5. max\nEnter")
	assert.Equal(t, [][2]string{{reviewRLClaudeArgs, "            args: [--print, --model, claude-opus-5-5, --effort, xhigh]"}},
		reviewFileDiff(t, b1, readReviewFile(t, dir)))
	assert.True(t, reviewArgsHasRun(reviewAssembledArgs(t, dir, "claude"), "--model", "claude-opus-5-5", "--effort", "xhigh"))

	run = runReviewCLIForms(t, dir, []string{"claude"}, reviewLines(reviewOpen,
		[]string{"\n", "\n", "\n", "5\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.NotContains(t, run.out, "claude effort")
	assert.Equal(t, "args: [--print, --model, claude-haiku-4-5]", reviewArgsLine(t, readReviewFile(t, dir), "claude"))
}

// S27, T1: a placeholder in another args element stays byte for byte, and
// neither its value, its name, nor its flag reaches the screen.
func TestReviewTab_S27ClaudePreviewKeepsOtherArgsOut(t *testing.T) {
	t.Setenv("REVIEW_TOKEN", "rvm-sentinel-7f3a9c")
	dir := writeReviewRL(t)
	secret := `            args: [--print, --api-key, "${REVIEW_TOKEN}", --model, claude-fable-5-1, --effort, max]`
	replaceReviewFile(t, dir, reviewRLClaudeArgs, secret)
	before := readReviewFile(t, dir)

	var run reviewRun
	stderr := captureStderr(t, func() {
		// claude opus → Enter (effort max stays), judge Enter, check Enter, 적용, 유지.
		run = runReviewCLIForms(t, dir, []string{"claude"}, reviewLines(reviewOpen,
			[]string{"\n", "\n", "\n", "3\n", "\n", "\n", "\n", "2\n", "\n"})...)
	})

	require.NoError(t, run.err, run.out)
	for _, hidden := range []string{"rvm-sentinel-7f3a9c", "REVIEW_TOKEN", "--api-key"} {
		assert.NotContains(t, run.out, hidden)
		assert.NotContains(t, stderr, hidden)
	}
	assert.Contains(t, run.out, "바뀌는 설정:\n  orchestra.providers.claude.args --model claude-fable-5-1 → claude-opus-5-5\n바뀐 model은")
	assert.Equal(t, [][2]string{{secret, strings.Replace(secret, "claude-fable-5-1", "claude-opus-5-5", 1)}},
		reviewFileDiff(t, before, readReviewFile(t, dir)))
}

// S26, claude half: a model off the row and an effort haiku does not take
// are errors; the tab's change lines name only the changed flag values.
func TestReviewCLI_S26ClaudeValidationAndLines(t *testing.T) {
	cfg := reviewR3Config()
	reviewRLConfig(cfg)
	baseline := newReviewBaseline(cfg, nil)
	attested := newReviewAttestation()
	attested.CLI["claude"] = reviewClaudeModels()
	cases := []struct {
		value   reviewCLIValue
		wantErr string
	}{
		{reviewCLIValue{"claude-opus-9", "max"}, "claude model claude-opus-9는 이 provider의 선택지에 없습니다"},
		{reviewCLIValue{"claude-haiku-4-5", "high"}, "claude effort high는 claude-haiku-4-5이 받지 않는 단계입니다"},
	}
	for _, tc := range cases {
		working := baseline.reviewComposition.clone()
		working.CLI["claude"] = tc.value
		_, err := planReviewComposition(baseline, working, attested, composeReviewTextKO)
		require.Error(t, err, tc.value)
		assert.Equal(t, tc.wantErr, err.Error())
	}

	working := baseline.reviewComposition.clone()
	working.CLI["claude"] = reviewCLIValue{"claude-opus-5-5", "xhigh"}
	plan, err := planReviewComposition(baseline, working, attested, composeReviewTextKO)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"orchestra.providers.claude.args --model claude-fable-5-1 → claude-opus-5-5",
		"orchestra.providers.claude.args --effort max → xhigh",
	}, plan.Changes)
	assert.Equal(t, "claude claude-opus-5-5 · xhigh *", plan.Table.rows[0][2])
	assert.Equal(t, "codex gpt-6-astra · max", plan.Table.rows[1][2])
	assert.Equal(t, "agy", plan.Table.rows[2][2])

	working.CLI["claude"] = reviewCLIValue{"claude-haiku-4-5", ""}
	plan, err = planReviewComposition(baseline, working, attested, composeReviewTextKO)
	require.NoError(t, err)
	assert.Equal(t, []string{
		"orchestra.providers.claude.args --model claude-fable-5-1 → claude-haiku-4-5",
		"orchestra.providers.claude.args --effort max → 없음",
	}, plan.Changes)
}
