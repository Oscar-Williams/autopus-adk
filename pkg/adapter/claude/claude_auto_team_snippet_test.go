//go:build !windows

package claude_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	snippetFailureLine = "Agent Team: off — auto workflow triage --topology failed; running subagents\n"
	snippetFactsUDJ    = `{"version":1,"flags":{"team":false,"workflow":false,"solo":false},"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/a"],"depends_on":[],"done":false},{"id":"T2","execution":"worker","owned_paths":["pkg/ab"],"depends_on":[],"done":false},{"id":"T3","execution":"worker","owned_paths":["pkg/c"],"depends_on":["T1"],"done":false},{"id":"T4","execution":"worker","owned_paths":["pkg/d"],"depends_on":[],"done":true}]}`
	snippetConfigKO    = "mode: full\nproject_name: proj\nplatforms:\n  - claude-code\nlanguage:\n  ai_responses: ko\n"
)

// extractAutoTeamSnippet returns the sh block of the generated section.
func extractAutoTeamSnippet(t *testing.T) string {
	t.Helper()
	autoGo, _ := generateAutoTeamSkills(t)
	section := between(t, autoGo, autoTeamSectionHeading, routeBHeading)
	block := between(t, section, "```sh\n", "\n```")
	return strings.TrimPrefix(block, "```sh\n") + "\n"
}

// writeExecutable writes an sh script at path.
func writeExecutable(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755))
}

// buildAutoForSnippet builds ./cmd/auto from this module into a temp dir.
func buildAutoForSnippet(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	bin := filepath.Join(t.TempDir(), "auto")
	build := exec.Command("go", "build", "-o", bin, "./cmd/auto")
	build.Dir = filepath.Join(filepath.Dir(file), "..", "..", "..")
	out, err := build.CombinedOutput()
	require.NoError(t, err, string(out))
	return bin
}

type snippetRun struct {
	stdout, stderr string
	teamJSON       []byte
	factsLeft      bool
}

// runAutoTeamSnippet runs the snippet with sh from project, resolving auto
// from autoDir only. The environment is built from scratch so the session
// running the test cannot leak its own Claude Code variables in.
func runAutoTeamSnippet(t *testing.T, snippet, autoDir, project string, extraEnv ...string) snippetRun {
	t.Helper()
	tmp := t.TempDir()
	facts, teamJSON := filepath.Join(tmp, "facts.json"), filepath.Join(tmp, "team.json")
	require.NoError(t, os.WriteFile(facts, []byte(snippetFactsUDJ), 0o600))
	require.NoError(t, os.WriteFile(teamJSON, nil, 0o600))
	cmd := exec.Command("/bin/sh", "-c", snippet)
	cmd.Dir = project
	cmd.Env = append([]string{"PATH=" + autoDir + ":/usr/bin:/bin", "HOME=" + tmp, "FACTS=" + facts, "TEAM_JSON=" + teamJSON}, extraEnv...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	require.NoError(t, cmd.Run(), stderr.String())
	data, err := os.ReadFile(teamJSON)
	require.NoError(t, err)
	_, statErr := os.Stat(facts)
	return snippetRun{stdout.String(), stderr.String(), data, statErr == nil}
}

func TestAutoTeamSnippetIsTheContractSnippet(t *testing.T) {
	t.Parallel()
	assert.Equal(t, autoTeamSnippet+"\n", extractAutoTeamSnippet(t))
}

func TestAutoTeamSnippetPrintsTheFailureLineWhenTheDecisionFails(t *testing.T) {
	t.Parallel()
	snippet := extractAutoTeamSnippet(t)
	project := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(project, "autopus.yaml"), []byte(snippetConfigKO), 0o644))
	fakes := map[string]string{
		"exit 1":                "exit 1",
		"exit 0 without output": "exit 0",
		// The CI stand-in for auto 0.50.125, which does not know the flag.
		"unknown flag": "printf 'Error: unknown flag: --topology\\n' >&2; exit 1",
	}
	for name, body := range fakes {
		dir := t.TempDir()
		writeExecutable(t, filepath.Join(dir, "auto"), body)
		run := runAutoTeamSnippet(t, snippet, dir, project)
		assert.Equal(t, snippetFailureLine, run.stdout, name)
		assert.Empty(t, run.stderr, name)
		assert.Empty(t, run.teamJSON, name)
		assert.False(t, run.factsLeft, name)
	}
}

func TestAutoTeamSnippetWithTheBranchBinary(t *testing.T) {
	t.Parallel()
	snippet := extractAutoTeamSnippet(t)
	bin := buildAutoForSnippet(t)
	probe := filepath.Join(t.TempDir(), "claude")
	writeExecutable(t, probe, "printf '2.1.288 (Claude Code)\\n'")
	capOK := []string{"CLAUDECODE=1", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1", "CLAUDE_CODE_ENTRYPOINT=cli", "CLAUDE_CODE_SESSION_ATTENDED=1", "CLAUDE_CODE_EXECPATH=" + probe}

	project := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(project, "autopus.yaml"), []byte(snippetConfigKO), 0o644))
	run := runAutoTeamSnippet(t, snippet, filepath.Dir(bin), project, capOK...)
	assert.Empty(t, run.stdout)
	assert.Empty(t, run.stderr)
	assert.False(t, run.factsLeft)
	var report struct {
		Decision struct {
			Route string `json:"route"`
		} `json:"decision"`
	}
	require.NoError(t, json.Unmarshal(run.teamJSON, &report), string(run.teamJSON))
	assert.Equal(t, "team", report.Decision.Route)

	broken := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(broken, "autopus.yaml"), []byte("workflow: [\n"), 0o644))
	run = runAutoTeamSnippet(t, snippet, filepath.Dir(bin), broken, capOK...)
	assert.Equal(t, snippetFailureLine, run.stdout)
	assert.Empty(t, run.stderr)
	assert.Empty(t, run.teamJSON)
	assert.False(t, run.factsLeft)
}
