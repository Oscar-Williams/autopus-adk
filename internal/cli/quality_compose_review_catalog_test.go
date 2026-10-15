package cli

// SPEC-REVMODEL-001 T5 (REQ-RVM-06, 08; acceptance S4(a)·(c), S8): a model
// off the catalog stays selectable as 그대로, an unreadable catalog asks no
// model, and a CLI-backend provider gets its note instead of a model select.

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// S4(a): CATN lacks gemini's model; 그대로 is marked, and the edit of another
// provider applies with gemini unchanged.
func TestReviewTab_S4aModelOffTheCatalogStays(t *testing.T) {
	dir := writeReviewR3(t)

	run := runReviewForms(t, dir, &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATN)}, reviewLines(reviewOpen,
		[]string{"\n", "\n", "\n", "\n", "3\n", "\n", "\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Contains(t, run.out, "gemini model \n1. 그대로 — google-antigravity/gemini-3.1-pro:high (catalog에 없음)\n"+
		"2. google-antigravity/gemini-3.8-flash\nEnter")
	loaded := reviewLoaderValue(t, dir)
	assert.Equal(t, "google-antigravity/gemini-3.1-pro:high", loaded.Orchestra.Providers["gemini"].Model)
	assert.Equal(t, "openai-codex/gpt-6-luna:high", loaded.Orchestra.Providers["codex"].Model)
}

// S4(c): a failed read and a catalog_invalid catalog print the note once,
// ask no model, and still apply a participation edit.
func TestReviewTab_S4cUnreadableCatalogAsksNoModel(t *testing.T) {
	cases := map[string]*ompCLIFakeRunner{
		"CATX": {fail: true},
		"CATE": {catalog: []byte(reviewCatalogCATE)},
	}
	for name, runner := range cases {
		t.Run(name, func(t *testing.T) {
			dir := writeReviewR3(t)

			run := runReviewForms(t, dir, runner, reviewLines(reviewOpen, []string{"\n", "\n", "2\n", "\n", "\n", "2\n", "\n"})...)

			require.NoError(t, run.err, run.out)
			assert.Equal(t, 1, strings.Count(run.out, "OMP catalog를 읽지 못해 model은 그대로입니다\n"))
			assert.NotContains(t, run.out, " model \n")
			assert.Contains(t, run.out, "구성을 적용했습니다.\n")
			assert.Equal(t, []string{"claude", "codex"}, reviewLoaderValue(t, dir).Spec.ReviewGate.Providers)
		})
	}
}

// reviewRLConfig is RL: R3 with the code default CLI providers, none of them
// OMP-backed.
func reviewRLConfig(cfg *config.HarnessConfig) {
	cfg.Orchestra.Providers = config.DefaultFullConfig("p").Orchestra.Providers
}

// S8: CLI-backend providers get the reason line once per pass and no model
// or thinking select; the write leaves orchestra.providers as it was.
func TestReviewTab_S8CLIBackendProvidersKeepTheirModel(t *testing.T) {
	dir := writeReviewR3(t, reviewRLConfig)
	before, err := config.Load(dir)
	require.NoError(t, err)
	for name, entry := range before.Orchestra.Providers {
		require.Empty(t, entry.Backend, name)
	}

	run := runReviewForms(t, dir, nil, reviewLines(reviewOpen, []string{"\n", "\n", "2\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	note := "는 CLI backend라 model을 바꾸지 않습니다: 증명된 model 목록은 OMP backend에만 있고 CLI model은 args와 quality 정책이 정합니다\n"
	assert.Equal(t, 1, strings.Count(run.out, "claude"+note))
	assert.Equal(t, 1, strings.Count(run.out, "codex"+note))
	assert.NotContains(t, run.out, "gemini"+note)
	assert.NotContains(t, run.out, " model \n")
	assert.NotContains(t, run.out, " thinking (")
	assert.Regexp(t, `(?m)^claude\s+켜짐\s+CLI backend\s+judge$`, run.out)
	assert.Regexp(t, `(?m)^codex\s+켜짐\s+CLI backend$`, run.out)
	assert.Regexp(t, `(?m)^gemini\s+꺼짐 \*\s+CLI backend$`, run.out)
	assert.Empty(t, run.runner.calls, "no OMP-backed provider reads no catalog")
	after, err := config.Load(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"claude", "codex"}, after.Spec.ReviewGate.Providers)
	assert.True(t, reflect.DeepEqual(before.Orchestra.Providers, after.Orchestra.Providers))
}
