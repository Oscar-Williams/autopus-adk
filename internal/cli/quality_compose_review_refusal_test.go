package cli

// SPEC-REVMODEL-001 T4 (REQ-RVM-12; acceptance S12): the line splice writer
// refuses every target the Line Splice Writer Contract calls unsafe. The
// refusal names the first refused path, and the file bytes and the run's
// orchestra and review gate stay as they were.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// reviewSpecLast moves the spec section to the end of an R3 file, so an
// alias inside it may name an anchor in orchestra.
func reviewSpecLast(t *testing.T, text string) string {
	t.Helper()
	start := strings.Index(text, "\nspec:\n") + 1
	end := start + strings.Index(text[start:], "\nmethodology:\n") + 1
	return text[:start] + text[end:] + text[start:end]
}

func reviewReplace(old, replacement string) func(*testing.T, string) string {
	return func(t *testing.T, text string) string {
		t.Helper()
		require.Contains(t, text, old)
		return strings.Replace(text, old, replacement, 1)
	}
}

func TestSaveReviewComposition_RefusesUnsafeTargets(t *testing.T) {
	t.Setenv("CODEX_REVIEW_MODEL", "openai-codex/gpt-6-astra:max")
	gemini := func(w *reviewComposition) { w.Providers = []string{"claude", "codex"} }
	luna := func(w *reviewComposition) { w.Models["codex"] = "openai-codex/gpt-6-luna:high" }
	cases := []struct {
		name string
		file func(*testing.T, string) string
		// late edits the file after the run loaded it: a shape no loader
		// accepts can still be what the writer reads.
		late   func(*testing.T, string) string
		change func(*reviewComposition)
		key    string
	}{
		{name: "S12(a) placeholder value", change: luna, key: "orchestra.providers.codex.model",
			file: reviewReplace("model: openai-codex/gpt-6-astra:max", "model: ${CODEX_REVIEW_MODEL}")},
		{name: "S12(b) alias value", change: gemini, key: "spec.review_gate.providers",
			file: func(t *testing.T, text string) string {
				text = reviewReplace("        review:\n            strategy: debate\n            providers: [claude, codex, gemini]",
					"        review:\n            strategy: debate\n            providers: &rp [claude, codex, gemini]")(t, text)
				return reviewSpecLast(t, reviewReplace("        providers: [claude, codex, gemini]\n        judge", "        providers: *rp\n        judge")(t, text))
			}},
		{name: "S12(c) anchored value", change: gemini, key: "spec.review_gate.providers",
			file: func(t *testing.T, text string) string {
				text = reviewReplace("        providers: [claude, codex, gemini]\n        judge", "        providers: &rp [claude, codex, gemini]\n        judge")(t, text)
				return reviewReplace("        review:\n            strategy: debate\n            providers: [claude, codex, gemini]",
					"        review:\n            strategy: debate\n            providers: *rp")(t, text)
			}},
		{name: "merge key in the parent", change: luna, key: "orchestra.providers.codex.model",
			file: func(t *testing.T, text string) string {
				text = reviewReplace("        claude:\n            backend: omp", "        claude: &entry\n            backend: omp")(t, text)
				return reviewReplace("        codex:\n            backend: omp", "        codex:\n            <<: *entry\n            backend: omp")(t, text)
			}},
		{name: "flow parent", change: luna, key: "orchestra.providers.codex.model",
			file: func(t *testing.T, text string) string {
				start := strings.Index(text, "        codex:\n            backend: omp")
				end := start + strings.Index(text[start:], "        gemini:\n")
				return text[:start] + "        codex: {backend: omp, model: openai-codex/gpt-6-astra:max, binary: \"\", args: [], subprocess: {timeout: 900}}\n" + text[end:]
			}},
		{name: "block scalar", change: func(w *reviewComposition) { w.Judge = "codex" }, key: "spec.review_gate.judge",
			file: reviewReplace("        judge: claude\n", "        judge: |-\n            claude\n")},
		{name: "comment inside a block list", change: gemini, key: "spec.review_gate.providers",
			file: reviewReplace("        providers: [claude, codex, gemini]\n",
				"        providers:\n            - claude\n            # keep codex\n            - codex\n            - gemini\n")},
		{name: "flow list over two lines", change: gemini, key: "spec.review_gate.providers",
			file: reviewReplace("        providers: [claude, codex, gemini]\n", "        providers: [claude, codex,\n            gemini]\n")},
		{name: "list item not a scalar", change: gemini, key: "spec.review_gate.providers",
			late: reviewReplace("        providers: [claude, codex, gemini]\n", "        providers: [claude, codex, [gemini]]\n")},
		{name: "value on the next line", change: luna, key: "orchestra.providers.codex.model",
			file: reviewReplace("            model: openai-codex/gpt-6-astra:max\n", "            model:\n                openai-codex/gpt-6-astra:max\n")},
		{name: "invalid selector", change: func(w *reviewComposition) { w.Models["codex"] = "openai-codex/gpt 6" },
			key: "orchestra.providers.codex.model"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeReviewR3(t)
			path := filepath.Join(dir, "autopus.yaml")
			if tc.file != nil {
				require.NoError(t, os.WriteFile(path, []byte(tc.file(t, readReviewFile(t, dir))), 0o644))
			}
			cfg, err := config.Load(dir)
			require.NoError(t, err)
			if tc.late != nil {
				require.NoError(t, os.WriteFile(path, []byte(tc.late(t, readReviewFile(t, dir))), 0o644))
			}
			before := readReviewFile(t, dir)
			orchestra, gate := cfg.Orchestra, cfg.Spec.ReviewGate
			baseline := newReviewBaseline(cfg, readReviewRaw(dir, cfg))
			working := baseline.reviewComposition.clone()
			tc.change(&working)

			err = saveReviewComposition(dir, cfg, baseline, working)

			var refusal reviewWriteRefusal
			require.True(t, errors.As(err, &refusal), "%v", err)
			assert.Equal(t, tc.key, refusal.key)
			assert.Equal(t, before, readReviewFile(t, dir))
			assert.True(t, reflect.DeepEqual(orchestra, cfg.Orchestra))
			assert.True(t, reflect.DeepEqual(gate, cfg.Spec.ReviewGate))
		})
	}
}

// A YAML-only line break (here U+2028) moves the node lines away from the
// \n lines the splice cuts, so the writer refuses before it edits.
func TestSaveReviewComposition_RefusesYAMLOnlyLineBreaks(t *testing.T) {
	dir := writeReviewR3(t)
	path := filepath.Join(dir, "autopus.yaml")
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	text := strings.Replace(readReviewFile(t, dir), "project_name: p\n", "project_name: \"p \"\n", 1)
	require.NoError(t, os.WriteFile(path, []byte(text), 0o644))
	baseline := newReviewBaseline(cfg, readReviewRaw(dir, cfg))
	working := baseline.reviewComposition.clone()
	working.Providers = []string{"claude", "codex"}

	err = saveReviewComposition(dir, cfg, baseline, working)

	var refusal reviewWriteRefusal
	require.True(t, errors.As(err, &refusal), "%v", err)
	assert.Equal(t, "spec.review_gate.providers", refusal.key)
	assert.Equal(t, text, readReviewFile(t, dir))
}
