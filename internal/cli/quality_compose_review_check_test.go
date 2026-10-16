package cli

// SPEC-REVMODEL-001 T6 (REQ-RVM-10, 11; acceptance S5, S16(b)): the live
// check runs only when the operator picks it, sends one request per
// provider through runProviderTransportSmoke with the exact smoke prompt and
// a 60-second budget, reports each result on one line, and never writes.

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// S5: four passes of one `auto update` run.
func TestReviewTab_S5LiveCheckIsOptInAndNeverWrites(t *testing.T) {
	dir := writeReviewR3(t)
	update := newReviewUpdate(t, dir, nil)
	b0 := readReviewFile(t, dir)
	open := []string{"2\n", "5\n"}
	// (1) codex luna → Enter, check, Enter at the confirmation (취소).
	first := []string{"\n", "\n", "\n", "\n", "3\n", "\n", "\n", "\n", "2\n", "\n"}
	// (2) every select Enter, check: no change, no confirmation.
	second := []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "2\n"}
	// (3) gemini flash → Enter, check, 적용.
	third := []string{"\n", "\n", "\n", "\n", "\n", "3\n", "\n", "\n", "2\n", "2\n"}
	// (4) claude opus → Enter, no check, 취소, 유지.
	fourth := []string{"\n", "\n", "\n", "3\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}

	out, err := runComposeUpdate(t, formInput(reviewLines(open, first, open, second, open, third, open, fourth)...), "--dir", dir)

	require.NoError(t, err, out)
	passes := strings.Split(out, "어느 도구를 바꿀까요?")
	require.Len(t, passes, 5, out)
	requests := update.smoke.recorded()
	require.Len(t, requests, 9)
	for i, request := range requests[:3] {
		assert.Equal(t, []string{"claude", "codex", "gemini"}[i], request.Provider)
		assert.Equal(t, "Return ONLY this exact token and no other text: AUTOPUS_PROVIDER_SMOKE_OK", request.Prompt)
		assert.Equal(t, 60*time.Second, request.Timeout)
	}
	assert.Equal(t, "openai-codex/gpt-6-luna:high", requests[1].Model)
	failure := "확인 호출이 실패한 provider가 있습니다. 적용하면 그 provider는 리뷰에서 실패합니다.\n"
	assert.Contains(t, passes[1], reviewCheckR3+"Enter a number between 1 and 2: \n"+
		"확인 호출 claude (anthropic/claude-fable-5-1:max): 통과\n"+
		"확인 호출 codex (openai-codex/gpt-6-luna:high): 통과\n"+
		"확인 호출 gemini (google-antigravity/gemini-3.1-pro:high): 실패 — HTTP 404 Not Found (Cloud Code Assist)\n"+
		failure+"적용할까요? \n1. 취소\n")
	assert.Contains(t, passes[1], "취소했습니다. 설정을 바꾸지 않았습니다.\n")
	assert.Contains(t, passes[2], "변경 없음\n"+reviewCheckR3)
	assert.Contains(t, passes[2], "확인 호출 codex (openai-codex/gpt-6-astra:max): 통과\n"+
		"확인 호출 gemini (google-antigravity/gemini-3.1-pro:high): 실패 — HTTP 404 Not Found (Cloud Code Assist)\n"+failure)
	assert.NotContains(t, passes[2], "적용할까요?")
	assert.Contains(t, passes[3], "확인 호출 gemini (google-antigravity/gemini-3.8-flash:high): 통과\n")
	assert.NotContains(t, passes[3], failure)
	assert.Contains(t, passes[4], "취소했습니다. 설정을 바꾸지 않았습니다.\n")
	assert.Equal(t, [][2]string{{"            model: google-antigravity/gemini-3.1-pro:high", "            model: google-antigravity/gemini-3.8-flash:high"}},
		reviewFileDiff(t, b0, readReviewFile(t, dir)), "only the applied pass wrote")
}

// S16(b): a provider name with terminal control characters refuses the
// live check before any request, and reaches the terminal escaped.
func TestReviewTab_S16bControlCharacterNameRefusesTheCheck(t *testing.T) {
	dir := writeReviewR3(t)
	text := strings.ReplaceAll(readReviewFile(t, dir), "gemini", `"gem\e[2Jini"`)
	text = strings.ReplaceAll(text, `google-antigravity/"gem\e[2Jini"-3.1-pro:high`, "google-antigravity/gemini-3.1-pro:high")
	dir = writeBlockFixture(t, text)

	run := runReviewForms(t, dir, nil, reviewLines(reviewOpen, []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	require.Contains(t, run.cfg.Orchestra.Providers, "gem\x1b[2Jini")
	assert.Contains(t, run.out, `확인 호출을 할 수 없습니다: provider 이름 gem\x1b[2Jini에 터미널 제어 문자가 있습니다`+"\n")
	assert.Empty(t, run.smoke.recorded())
	assert.NotContains(t, run.out, "\x1b")
}

// Result lines: a review_gate result names no model, a missing marker
// warns, and a failure detail is its first line, terminal-safe, cut to 200
// runes.
func TestReviewCheckLine_ResultShapes(t *testing.T) {
	candidate := reviewR3Config()
	candidate.Orchestra.Providers["claude"] = config.DefaultClaudeProviderEntry()
	baseline := newReviewBaseline(candidate, nil)
	shown := reviewCheckModels(baseline, baseline.reviewComposition)
	text := composeReviewTextKO

	assert.Equal(t, "확인 호출 review_gate: 실패 — no installed review providers",
		reviewCheckLine(candidate, shown, providerSmokeResult{Provider: "review_gate", Status: "fail", Detail: "no installed review providers"}, text))
	assert.Equal(t, "확인 호출 codex (openai-codex/gpt-6-astra:max): 응답에 확인 문자열이 없습니다",
		reviewCheckLine(candidate, shown, providerSmokeResult{Provider: "codex", Status: "warn", Detail: "missing"}, text))
	assert.Equal(t, "확인 호출 claude (CLI backend): 통과",
		reviewCheckLine(candidate, shown, providerSmokeResult{Provider: "claude", Status: "pass"}, text))
	long := strings.Repeat("가", 250) + "\nsecond line"
	line := reviewCheckLine(candidate, shown, providerSmokeResult{Provider: "gemini", Status: "fail", Detail: "\x1b[31m" + long}, text)
	assert.Equal(t, "확인 호출 gemini (google-antigravity/gemini-3.1-pro:high): 실패 — "+`\x1b[31m`+strings.Repeat("가", 192), line)
	assert.Equal(t, "live check codex (openai-codex/gpt-6-astra:max): passed",
		reviewCheckLine(candidate, shown, providerSmokeResult{Provider: "codex", Status: "pass"}, composeReviewTextEN))
}

// The shown model of an OMP provider: the chosen one, the raw text of an
// unchanged one, and none without raw text or for a placeholder; never the
// expanded value.
func TestReviewCheckModels_ChosenOrRaw(t *testing.T) {
	cfg := reviewR3Config()
	cfg.Orchestra.Providers["codex"] = reviewOMPEntry("openai-codex/rvm-sentinel-7f3a9c:max")
	baseline := newReviewBaseline(cfg, &reviewRawDocument{Names: true, Providers: map[string]reviewRawProvider{
		"claude": {Backend: "omp", Model: "anthropic/claude-fable-5-1:max"},
		"codex":  {Backend: "omp", Model: "${CODEX_REVIEW_MODEL}"},
	}})
	working := baseline.reviewComposition.clone()
	working.Models["claude"] = "anthropic/claude-opus-5-5:high"

	assert.Equal(t, map[string]string{"claude": "anthropic/claude-opus-5-5:high"}, reviewCheckModels(baseline, working))
	working.Models["claude"] = baseline.Models["claude"]
	assert.Equal(t, map[string]string{"claude": "anthropic/claude-fable-5-1:max"}, reviewCheckModels(baseline, working))
	assert.Equal(t, "확인 호출 gemini: 통과", reviewCheckLine(cfg, reviewCheckModels(baseline, working),
		providerSmokeResult{Provider: "gemini", Status: "pass"}, composeReviewTextKO))
}

// The result line names no model written as a placeholder: the check runs
// the expanded value, but it never reaches the screen.
func TestReviewTab_CheckLineOmitsAPlaceholderModel(t *testing.T) {
	t.Setenv("CODEX_REVIEW_MODEL", "openai-codex/rvm-sentinel-7f3a9c:max")
	dir := writeReviewR3(t)
	replaceReviewFile(t, dir, "model: openai-codex/gpt-6-astra:max", "model: ${CODEX_REVIEW_MODEL}")

	// participation ×3, claude Enter, gemini Enter, judge Enter, 확인 호출, the menu's default.
	run := runReviewForms(t, dir, nil, reviewLines(reviewOpen, []string{"\n", "\n", "\n", "\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "codex의 model 값이 환경 변수라 이 탭에서 바꾸지 않습니다: ${CODEX_REVIEW_MODEL}\n")
	assert.Contains(t, run.out, "확인 호출 codex: 통과\n")
	assert.NotContains(t, run.out, "rvm-sentinel-7f3a9c")
	requests := run.smoke.recorded()
	require.Len(t, requests, 3)
	assert.Equal(t, "openai-codex/rvm-sentinel-7f3a9c:max", requests[1].Model, "the check runs what the runtime runs")
}
