package cli

// SPEC-REVMODEL-001 T13 (REQ-RVM-20, 22, 25, 26; acceptance S16, S20, S24,
// S25, S27): a codex CLI provider picks a model of the run's Codex catalog
// and one of its efforts, written as model_policy: pinned and the values of
// -m and model_reasoning_effort only, which no profile overrides afterwards.

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

const reviewRLCodexArgs = `            args: [exec, --json, --sandbox, workspace-write, -m, gpt-6-astra, -c, model_reasoning_effort="max"]`

// reviewRuntimeArgs is the codex argv the runtime assembles from dir's loader
// value after a global --effort low.
func reviewRuntimeArgs(t *testing.T, dir string) []string {
	t.Helper()
	effective, err := loadEffectiveHarnessConfigForDir(dir, globalFlags{})
	require.NoError(t, err)
	effective = applyRuntimeHarnessOverrides(effective, globalFlags{Effort: "low"})
	set, err := assembleSpecReviewProviders(t.Context(), specReviewProviderRequest{Config: effective.Config})
	require.NoError(t, err)
	for _, provider := range set.Providers {
		if provider.Name == "codex" {
			return provider.Args
		}
	}
	t.Fatal("no assembled codex")
	return nil
}

// S20: the catalog's models by name, the model's levels only, pinned, and an
// argv no quality profile rewrites.
func TestReviewTab_S20CodexCLIModelIsPinned(t *testing.T) {
	dir, b1 := writeReviewRL(t), writeReviewRL(t)
	before := readReviewFile(t, dir)

	// participation Enter ×3, claude Enter, codex gpt-6.1-sol → high, judge Enter, check Enter, 적용, 유지.
	run := runReviewCLIForms(t, dir, []string{"claude", "codex"}, reviewLines(reviewOpen,
		[]string{"\n", "\n", "\n", "\n", "4\n", "3\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "codex model \n1. 그대로 — gpt-6-astra · max\n2. gpt-6-astra\n3. gpt-6-luna\n4. gpt-6.1-sol\nEnter")
	assert.Contains(t, run.out, "codex effort \n1. low\n2. medium\n3. high\n4. xhigh\nEnter")
	assert.Equal(t, [][2]string{
		{reviewRLCodexArgs, `            args: [exec, --json, --sandbox, workspace-write, -m, gpt-6.1-sol, -c, model_reasoning_effort="high"]`},
		{"            model_policy: quality", "            model_policy: pinned"},
	}, reviewFileDiff(t, before, readReviewFile(t, dir)))
	pinned := reviewRuntimeArgs(t, dir)
	assert.True(t, reviewArgsHasRun(pinned, "-m", "gpt-6.1-sol"), "%v", pinned)
	assert.Contains(t, pinned, `model_reasoning_effort="high"`)
	quality := reviewRuntimeArgs(t, b1)
	assert.True(t, reviewArgsHasRun(quality, "-m", "gpt-6-astra"), "%v", quality)
	assert.Contains(t, quality, `model_reasoning_effort="low"`)
}

// S20: the effort default is the working effort when the model takes it,
// otherwise its highest level.
func TestReviewCLIEffortDefault_CodexHighestReportedLevel(t *testing.T) {
	models, ok := reviewCodexModels(driftGenerationSnapshot{codexCatalog: []byte(mustReadComposeCatalog(t, "CAT-E")), codexCatalogFixed: true})
	require.True(t, ok)
	sol := models[2]
	require.Equal(t, "gpt-6.1-sol", sol.ID)
	assert.Equal(t, "xhigh", reviewCLIEffortDefault(reviewCodexRow, sol, reviewCLIValue{"gpt-6-astra", "max"}))
	assert.Equal(t, "low", reviewCLIEffortDefault(reviewCodexRow, sol, reviewCLIValue{"gpt-6-astra", "low"}))
	_, ok = reviewCodexModels(driftGenerationSnapshot{})
	assert.False(t, ok, "no snapshot is an unreadable catalog")
}

// S24: auto quality decides on the migrated args and writes the raw node:
// a codex entry without args gets model_policy and args lines, a pinned one
// with extra flags changes two tokens.
func TestQualityComposeReview_S24MigratedBaselineAndRawArgs(t *testing.T) {
	newComposeUpdateRun(t, true, "")
	useAccessibleComposeForms(t)
	composeSessionCatalog(t, "CAT-E")
	useReviewPath(t, "claude", "codex")
	deps := ompPlatformDependencies{}
	codexBlock := "        codex:\n            binary: codex\n" + reviewRLCodexArgs + "\n            model_policy: quality\n" +
		"            subprocess:\n                schema_flag: --output-schema\n                timeout: 420\n"

	rq := writeReviewRL(t)
	replaceReviewFile(t, rq, codexBlock, "        codex:\n            binary: codex\n")
	replaceReviewFile(t, rq, reviewRLClaudeArgs, "            args: [--print, --verbose, --model, claude-fable-5-1, --effort, max]")
	before := readReviewFile(t, rq)
	// 구성 편집, 리뷰, participation ×3, claude opus → high, codex luna → high, judge, check, 적용, 유지.
	out, err := runQualityForms(t, rq, &deps, formInput(reviewLines([]string{"2\n", "5\n"},
		[]string{"\n", "\n", "\n", "3\n", "3\n", "3\n", "3\n", "\n", "\n", "2\n", "\n"})...))

	require.NoError(t, err, out)
	assert.Contains(t, out, "codex model \n1. 그대로 — gpt-6-astra · max\n")
	want := strings.Replace(before, "        codex:\n            binary: codex\n", "        codex:\n            binary: codex\n"+
		"            model_policy: pinned\n"+
		`            args: [exec, --json, --sandbox, workspace-write, -m, gpt-6-luna, -c, model_reasoning_effort="high"]`+"\n", 1)
	want = strings.Replace(want, "--verbose, --model, claude-fable-5-1, --effort, max", "--verbose, --model, claude-opus-5-5, --effort, high", 1)
	assert.Equal(t, want, readReviewFile(t, rq))
	assert.True(t, reviewArgsHasRun(reviewAssembledArgs(t, rq, "codex"), "-m", "gpt-6-luna"))

	rp := writeReviewRL(t)
	rpArgs := `            args: [exec, --json, --json, --sandbox, workspace-write, -m, gpt-6-astra, -c, model_reasoning_effort="max", --skip-git-repo-check]`
	replaceReviewFile(t, rp, reviewRLCodexArgs+"\n            model_policy: quality\n", rpArgs+"\n            model_policy: pinned\n")
	before = readReviewFile(t, rp)
	out, err = runQualityForms(t, rp, &deps, formInput(reviewLines([]string{"2\n", "5\n"},
		[]string{"\n", "\n", "\n", "\n", "3\n", "3\n", "\n", "\n", "2\n", "\n"})...))

	require.NoError(t, err, out)
	assert.Equal(t, [][2]string{{rpArgs, strings.Replace(strings.Replace(rpArgs, "gpt-6-astra", "gpt-6-luna", 1),
		`effort="max"`, `effort="high"`, 1)}}, reviewFileDiff(t, before, readReviewFile(t, rp)))
}

// S25: a model with no level in low–max is not offered, and naming it is an
// error.
func TestReviewTab_S25CodexModelWithoutLadderLevelsIsHidden(t *testing.T) {
	catalog := strings.Replace(mustReadComposeCatalog(t, "CAT-E"), `{"models":[`,
		`{"models":[{"slug":"gpt-6-mini","supported_reasoning_levels":[{"effort":"minimal"}]},`, 1)
	dir := writeReviewRL(t)

	run := runReviewCLIFormsWith(t, dir, catalog, nil, reviewLines(reviewOpen, []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "codex model \n1. 그대로 — gpt-6-astra · max\n2. gpt-6-astra\n3. gpt-6-luna\n4. gpt-6.1-sol\nEnter")
	assert.NotContains(t, run.out, "gpt-6-mini")

	cfg := reviewR3Config()
	reviewRLConfig(cfg)
	baseline := newReviewBaseline(cfg, nil)
	attested := newReviewAttestation()
	attested.CLI["codex"], _ = reviewCodexModels(driftGenerationSnapshot{codexCatalog: []byte(catalog), codexCatalogFixed: true})
	working := baseline.reviewComposition.clone()
	working.CLI["codex"] = reviewCLIValue{"gpt-6-mini", "minimal"}
	_, err := planReviewComposition(baseline, working, attested, composeReviewTextKO)
	require.Error(t, err)
	assert.Equal(t, "codex model gpt-6-mini는 이 provider의 선택지에 없습니다", err.Error())
}

// S27, T2: a placeholder -m value gets the refusal line instead of the codex
// model select, and its value never reaches the screen.
func TestReviewTab_S27CodexPlaceholderModelIsRefused(t *testing.T) {
	t.Setenv("CODEX_MODEL", "rvm-model-sentinel-41")
	dir := writeReviewRL(t)
	replaceReviewFile(t, dir, "-m, gpt-6-astra,", `-m, "${CODEX_MODEL}",`)
	sum := reviewFileSum(t, dir)

	var run reviewRun
	stderr := captureStderr(t, func() {
		run = runReviewCLIForms(t, dir, nil, reviewLines(reviewOpen, []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n"})...)
	})

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "codex의 -m 값이 환경 변수라 이 탭에서 바꾸지 않습니다: ${CODEX_MODEL}\n")
	assert.NotContains(t, run.out, "codex model \n")
	assert.NotContains(t, run.out+stderr, "rvm-model-sentinel-41")
	assert.Equal(t, sum, reviewFileSum(t, dir))
}

// S16: the codex fallback line passes config values through terminalSafe.
func TestReportCodexRuntimeFallback_ValuesAreTerminalSafe(t *testing.T) {
	var out bytes.Buffer
	provider := orchestra.ProviderConfig{Name: "codex", Binary: "codex", ModelPolicy: config.ProviderModelPolicyQuality,
		Args: []string{"exec", "-m", "gpt\x1b[2J", "-c", "model_reasoning_effort=\"hi\x1b[0mgh\""}}
	probe := func(context.Context, string) ([]byte, error) { return nil, nil }

	resolveCodexProviderCapabilitiesWith(context.Background(), []orchestra.ProviderConfig{provider}, probe, &out)

	assert.NotContains(t, out.String(), "\x1b")
	assert.Contains(t, out.String(), `requested=gpt\x1b[2J/hi\x1b[0mgh`)
}

// L1: a Codex catalog slug is offered only when it reads as a model name; a
// flag-like or control-character slug never reaches the select or argv.
func TestReviewCodexModels_OnlySlugsThatReadAsNames(t *testing.T) {
	catalog := strings.Replace(mustReadComposeCatalog(t, "CAT-E"), `{"models":[`, `{"models":[`+
		`{"slug":"--dangerously-bypass-approvals-and-sandbox","supported_reasoning_levels":[{"effort":"low"}]},`+
		`{"slug":"\u0000omp","supported_reasoning_levels":[{"effort":"low"}]},`+
		`{"slug":"gpt 6","supported_reasoning_levels":[{"effort":"low"}]},`+
		`{"slug":"openai/gpt-6.2:preview","supported_reasoning_levels":[{"effort":"low"}]},`, 1)

	models, ok := reviewCodexModels(driftGenerationSnapshot{codexCatalog: []byte(catalog), codexCatalogFixed: true})

	require.True(t, ok)
	assert.Equal(t, []string{"gpt-6-astra", "gpt-6-luna", "gpt-6.1-sol", "openai/gpt-6.2:preview"}, reviewSelectorsOfCLI(models))

	dir := writeReviewRL(t)
	run := runReviewCLIFormsWith(t, dir, catalog, nil, reviewLines(reviewOpen, []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"})...)
	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "codex model \n1. 그대로 — gpt-6-astra · max\n2. gpt-6-astra\n3. gpt-6-luna\n4. gpt-6.1-sol\n5. openai/gpt-6.2:preview\nEnter")
	assert.NotContains(t, run.out, "dangerously")
}

// L1: a wanted value that starts with `-` would read as a flag of its own,
// so the planner refuses it and both writers write nothing.
func TestPlanReviewArgsEdit_RefusesAFlagLikeValue(t *testing.T) {
	args := []string{"exec", "-m", "gpt-6-astra", "-c", `model_reasoning_effort="max"`}
	for _, want := range []reviewCLIValue{{Model: "--dangerously-bypass-approvals-and-sandbox"}, {Model: "gpt-6-luna", Effort: "-x"}} {
		_, ok := planReviewArgsEdit(args, reviewCodexRow, want)
		assert.False(t, ok, "%v", want)
	}
	edit, ok := planReviewArgsEdit(args, reviewCodexRow, reviewCLIValue{"gpt-6-luna", "high"})
	require.True(t, ok)
	assert.Equal(t, []string{"exec", "-m", "gpt-6-luna", "-c", `model_reasoning_effort="high"`}, edit.apply(args))

	dir := writeReviewRL(t)
	sum := reviewFileSum(t, dir)
	_, err := saveReviewEdit(t, dir, func(w *reviewComposition) {
		w.CLI["codex"] = reviewCLIValue{Model: "--dangerously-bypass-approvals-and-sandbox"}
	})
	var refusal reviewWriteRefusal
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, "orchestra.providers.codex.args", refusal.key)
	assert.Equal(t, sum, reviewFileSum(t, dir))

	cfg := reviewR3Config()
	reviewRLConfig(cfg)
	baseline := newReviewBaseline(cfg, nil)
	working := baseline.reviewComposition.clone()
	working.CLI["codex"] = reviewCLIValue{Model: "--dangerously-bypass-approvals-and-sandbox"}
	before := slices.Clone(cfg.Orchestra.Providers["codex"].Args)
	require.ErrorAs(t, applyReviewCompositionMemory(cfg, baseline, working), &refusal)
	assert.Equal(t, before, cfg.Orchestra.Providers["codex"].Args)
}

// L1: the switch answer never indexes an empty choice list.
func TestReviewSwitch_NoChoicesSwitchesNothing(t *testing.T) {
	var f composeFormRun
	tab := &reviewTabState{working: reviewComposition{OMP: map[string]bool{}, Models: map[string]string{}}}

	assert.NotPanics(t, func() { assert.NoError(t, f.reviewSwitch(tab, "codex", nil)) })
	assert.False(t, tab.working.OMP["codex"])
}
