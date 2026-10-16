package cli

// SPEC-REVMODEL-001 T5 (REQ-RVM-06; acceptance S4(a)·(c)): a model off the
// catalog stays selectable as 그대로, and an unreadable catalog asks no model.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
