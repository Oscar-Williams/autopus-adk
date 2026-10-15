package cli

// SPEC-REVMODEL-001 T8 (REQ-RVM-15; acceptance S7): the files this branch's
// review writer leaves add no key path except a missing judge, and the
// v0.50.125 binary still loads them while it refuses an unknown review key.

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// reviewCompatEnv names the v0.50.125 `auto` binary (git archive v0.50.125,
// dd6d6ea2) the load half of S7 runs. Without it that half skips; the key
// path half always runs.
const reviewCompatEnv = "AUTOPUS_V050125_AUTO"

// reviewCompatCase is one S7 file: the start fixture and the edit the tab
// applies to it.
type reviewCompatCase struct {
	name  string
	start func(*testing.T) string
	edit  func(*reviewComposition)
	added []string
}

var reviewCompatCases = []reviewCompatCase{
	{name: "S1", start: func(t *testing.T) string { return writeReviewR3(t) },
		edit: func(w *reviewComposition) { w.Providers = []string{"claude", "codex"} }},
	{name: "S2", start: func(t *testing.T) string { return writeReviewR3(t) }, edit: func(w *reviewComposition) {
		w.Models["claude"], w.Models["codex"] = "anthropic/claude-opus-5-5:max", "openai-codex/gpt-6-luna:high"
	}},
	{name: "S3(b)", start: func(t *testing.T) string { return writeReviewR3(t) },
		edit: func(w *reviewComposition) { w.Providers, w.Judge = []string{"codex", "gemini"}, "codex" }},
	{name: "S5(3)", start: func(t *testing.T) string { return writeReviewR3(t) },
		edit: func(w *reviewComposition) { w.Models["gemini"] = "google-antigravity/gemini-3.8-flash:high" }},
	{name: "S9(b)", start: func(t *testing.T) string { return writeReviewR3Judge(t, "") },
		edit: func(w *reviewComposition) { w.Judge = "codex" }, added: []string{"spec.review_gate.judge"}},
}

// writeReviewCompatFile applies the case's edit to its start file through
// saveReviewComposition and returns the project and the start text.
func writeReviewCompatFile(t *testing.T, tc reviewCompatCase) (string, string) {
	t.Helper()
	dir := tc.start(t)
	start := readReviewFile(t, dir)
	_, err := saveReviewEdit(t, dir, tc.edit)
	require.NoError(t, err)
	return dir, start
}

// reviewKeyPaths lists every mapping key path of a YAML text, list items as
// [], sorted and without repeats.
func reviewKeyPaths(t *testing.T, text string) []string {
	t.Helper()
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(text), &doc))
	var paths []string
	var walk func(prefix string, node *yaml.Node)
	walk = func(prefix string, node *yaml.Node) {
		switch node.Kind {
		case yaml.DocumentNode:
			for _, child := range node.Content {
				walk(prefix, child)
			}
		case yaml.SequenceNode:
			for _, child := range node.Content {
				walk(prefix+"[]", child)
			}
		case yaml.MappingNode:
			for i := 0; i+1 < len(node.Content); i += 2 {
				path := node.Content[i].Value
				if prefix != "" {
					path = prefix + "." + path
				}
				paths = append(paths, path)
				walk(path, node.Content[i+1])
			}
		}
	}
	walk("", &doc)
	slices.Sort(paths)
	return slices.Compact(paths)
}

func TestReviewCompat_S7KeyPathsStayTheStartFiles(t *testing.T) {
	for _, tc := range reviewCompatCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, start := writeReviewCompatFile(t, tc)
			want := slices.Sorted(slices.Values(append(reviewKeyPaths(t, start), tc.added...)))
			assert.Equal(t, want, reviewKeyPaths(t, readReviewFile(t, dir)))
		})
	}
}

func TestReviewCompat_S7V050125LoadsTheWrittenFiles(t *testing.T) {
	binary := os.Getenv(reviewCompatEnv)
	if binary == "" {
		t.Skipf("set %s to an auto built from git archive v0.50.125", reviewCompatEnv)
	}
	load := func(t *testing.T, dir string) (string, error) {
		t.Helper()
		cmd := exec.Command(binary, "config", "get", "usage_profile")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	var s1 string
	for _, tc := range reviewCompatCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, _ := writeReviewCompatFile(t, tc)
			out, err := load(t, dir)
			require.NoError(t, err, out)
			assert.Equal(t, "developer\n", out)
			if tc.name == "S1" {
				s1 = readReviewFile(t, dir)
			}
		})
	}
	require.NotEmpty(t, s1)
	control := writeBlockFixture(t, strings.Replace(s1, "        providers: [claude, codex]\n",
		"        providers: [claude, codex]\n        reviewers: [claude]\n", 1))
	out, err := load(t, control)
	var exit *exec.ExitError
	require.ErrorAs(t, err, &exit, out)
	assert.Equal(t, 1, exit.ExitCode())
	assert.Contains(t, out, "field reviewers not found in type config.ReviewGateConf")
}
