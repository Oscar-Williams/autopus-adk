package cli

// SPEC-REVMODEL-001 T4 (REQ-RVM-03, 07, 12, 15; acceptance S9, S11, S12):
// the line splice writer changes only the value spans it plans, keeps every
// other byte, verifies the whole document, writes atomically with the file
// mode kept, and syncs the run's orchestra and review gate from the re-read
// file.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/config"
)

// writeReviewRH writes RH: R3's values by hand, with two-space indentation,
// a comment line right above spec, blank lines between sections, and an
// inline comment on the review provider list.
func writeReviewRH(t *testing.T) string {
	t.Helper()
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	require.NoError(t, encoder.Encode(reviewR3Config()))
	require.NoError(t, encoder.Close())
	text := buffer.String()
	for old, replacement := range map[string]string{
		"\nspec:\n":      "\n\n# review settings\nspec:\n",
		"\norchestra:\n": "\n\norchestra:\n",
		"\n    providers: [claude, codex, gemini]\n": "\n    providers: [claude, codex, gemini] # reviewers\n",
	} {
		require.Contains(t, text, old)
		text = strings.Replace(text, old, replacement, 1)
	}
	return writeBlockFixture(t, text)
}

// saveReviewEdit applies edit to the baseline of dir's loaded config through
// saveReviewComposition and returns the run's config.
func saveReviewEdit(t *testing.T, dir string, edit func(*reviewComposition)) (*config.HarnessConfig, error) {
	t.Helper()
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	baseline := newReviewBaseline(cfg, readReviewRaw(dir))
	working := baseline.reviewComposition.clone()
	edit(&working)
	return cfg, saveReviewComposition(dir, cfg, baseline, working)
}

// S1 and S11(2) at the writer: one list span, the reloaded config synced.
func TestSaveReviewComposition_ListSpanAndSync(t *testing.T) {
	dir := writeReviewR3(t)
	before := readReviewFile(t, dir)

	cfg, err := saveReviewEdit(t, dir, func(w *reviewComposition) { w.Providers = []string{"claude", "codex"} })

	require.NoError(t, err)
	after := readReviewFile(t, dir)
	assert.Equal(t, [][2]string{{"        providers: [claude, codex, gemini]", "        providers: [claude, codex]"}},
		reviewFileDiff(t, before, after))
	assert.Equal(t, []string{"claude", "codex"}, cfg.Spec.ReviewGate.Providers)
	assert.Equal(t, []string{"claude", "codex", "gemini"}, cfg.Orchestra.Commands["review"].Providers)
}

// S2 at the writer: the model value token only, as an exact selector.
func TestSaveReviewComposition_ModelValueToken(t *testing.T) {
	dir := writeReviewR3(t)
	before := readReviewFile(t, dir)

	cfg, err := saveReviewEdit(t, dir, func(w *reviewComposition) {
		w.Models["codex"] = "openai-codex/gpt-6-luna:high"
		w.Models["claude"] = "anthropic/claude-opus-5-5:max"
	})

	require.NoError(t, err)
	assert.Equal(t, [][2]string{
		{"            model: anthropic/claude-fable-5-1:max", "            model: anthropic/claude-opus-5-5:max"},
		{"            model: openai-codex/gpt-6-astra:max", "            model: openai-codex/gpt-6-luna:high"},
	}, reviewFileDiff(t, before, readReviewFile(t, dir)))
	assert.Equal(t, "openai-codex/gpt-6-luna:high", cfg.Orchestra.Providers["codex"].Model)
}

// S9(b)·(c): a missing judge key is added after the review_gate mapping's
// last line at its indentation; judge: "" changes only its value.
func TestSaveReviewComposition_JudgeInsertAndReplace(t *testing.T) {
	dir := writeReviewR3(t)
	full := readReviewFile(t, dir)
	require.Contains(t, full, "        judge: claude\n")
	rn := strings.Replace(full, "        judge: claude\n", "", 1)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(rn), 0o644))

	cfg, err := saveReviewEdit(t, dir, func(w *reviewComposition) { w.Judge = "codex" })

	require.NoError(t, err)
	after := readReviewFile(t, dir)
	gate := rn[strings.Index(rn, "    review_gate:\n"):]
	last := gate[:strings.Index(gate, "\nmethodology:")]
	last = last[strings.LastIndex(last, "\n")+1:]
	assert.Equal(t, strings.Replace(rn, last+"\n", last+"\n        judge: codex\n", 1), after)
	assert.Equal(t, "codex", cfg.Spec.ReviewGate.Judge)

	re := strings.Replace(full, "        judge: claude\n", "        judge: \"\"\n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(re), 0o644))
	_, err = saveReviewEdit(t, dir, func(w *reviewComposition) { w.Judge = "claude" })
	require.NoError(t, err)
	assert.Equal(t, [][2]string{{"        judge: \"\"", "        judge: claude"}}, reviewFileDiff(t, re, readReviewFile(t, dir)))
}

// S12, RH: two value spans change; comments, blank lines, the inline
// comment, and the two-space layout stay byte for byte.
func TestSaveReviewComposition_HandWrittenFileKeepsEveryOtherByte(t *testing.T) {
	dir := writeReviewRH(t)
	before := readReviewFile(t, dir)
	require.NoError(t, os.Chmod(filepath.Join(dir, "autopus.yaml"), 0o600))

	_, err := saveReviewEdit(t, dir, func(w *reviewComposition) {
		w.Providers = []string{"claude", "codex"}
		w.Models["codex"] = "openai-codex/gpt-6-luna:high"
	})

	require.NoError(t, err)
	assert.Equal(t, [][2]string{
		{"    providers: [claude, codex, gemini] # reviewers", "    providers: [claude, codex] # reviewers"},
		{"      model: openai-codex/gpt-6-astra:max", "      model: openai-codex/gpt-6-luna:high"},
	}, reviewFileDiff(t, before, readReviewFile(t, dir)))
	assert.Contains(t, readReviewFile(t, dir), "\n\n# review settings\nspec:\n")
	assert.Equal(t, os.FileMode(0o600), mustFileMode(t, filepath.Join(dir, "autopus.yaml")))
}

// Block lists, quoted scalars, and inline comments.
func TestSaveReviewComposition_BlockListAndQuotedScalar(t *testing.T) {
	dir := writeReviewR3(t)
	full := readReviewFile(t, dir)
	text := strings.Replace(full, "        providers: [claude, codex, gemini]\n",
		"        providers:\n            - claude\n            - codex\n            - gemini # last\n", 1)
	text = strings.Replace(text, "            model: openai-codex/gpt-6-astra:max\n",
		"            model: 'openai-codex/gpt-6-astra:max' # pinned\n", 1)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(text), 0o644))

	cfg, err := saveReviewEdit(t, dir, func(w *reviewComposition) {
		w.Providers = []string{"claude", "gemini"}
		w.Models["codex"] = "openai-codex/gpt-6-luna:high"
	})

	require.NoError(t, err)
	want := strings.Replace(text, "            - claude\n            - codex\n            - gemini # last\n",
		"            - claude\n            - gemini\n", 1)
	want = strings.Replace(want, "'openai-codex/gpt-6-astra:max' # pinned", "openai-codex/gpt-6-luna:high # pinned", 1)
	assert.Equal(t, want, readReviewFile(t, dir))
	assert.Equal(t, []string{"claude", "gemini"}, cfg.Spec.ReviewGate.Providers)
}

// SPEC-PANERM-001 REQ-10: retired keys never ride along; the result is the
// pruned file with the one span changed.
func TestSaveReviewComposition_PrunesRetiredKeysFirst(t *testing.T) {
	dir := writeReviewR3(t)
	full := readReviewFile(t, dir)
	text := strings.Replace(full, "    providers:\n        claude:\n", "    providers:\n        claude:\n            pane_args: [--x]\n", 1)
	require.NotEqual(t, full, text)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(text), 0o644))
	pruned, paths, err := pruneRetiredConfig([]byte(text))
	require.NoError(t, err)
	require.NotEmpty(t, paths)

	_, err = saveReviewEdit(t, dir, func(w *reviewComposition) { w.Providers = []string{"claude", "codex"} })

	require.NoError(t, err)
	assert.Equal(t, strings.Replace(string(pruned), "providers: [claude, codex, gemini]\n        judge",
		"providers: [claude, codex]\n        judge", 1), readReviewFile(t, dir))
}

// init's memory path sets only the planned values and validates first.
func TestApplyReviewCompositionMemory_SetsThePlannedValues(t *testing.T) {
	cfg := reviewR3Config()
	baseline := newReviewBaseline(cfg, nil)
	working := baseline.reviewComposition.clone()
	working.Providers = []string{"claude", "codex"}
	working.Models["codex"] = "openai-codex/gpt-6-luna:high"

	require.NoError(t, applyReviewCompositionMemory(cfg, baseline, working))

	assert.Equal(t, []string{"claude", "codex"}, cfg.Spec.ReviewGate.Providers)
	assert.Equal(t, "openai-codex/gpt-6-luna:high", cfg.Orchestra.Providers["codex"].Model)
	assert.Equal(t, "anthropic/claude-fable-5-1:max", cfg.Orchestra.Providers["claude"].Model)
	assert.Equal(t, []string{"claude", "codex", "gemini"}, cfg.Orchestra.Commands["review"].Providers)

	broken := working.clone()
	broken.Models["codex"] = "not a selector"
	require.Error(t, applyReviewCompositionMemory(cfg, baseline, broken))
	assert.Equal(t, "openai-codex/gpt-6-luna:high", cfg.Orchestra.Providers["codex"].Model, "a refused edit sets nothing")
}
