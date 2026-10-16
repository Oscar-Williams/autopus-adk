package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const reverseArtifactPattern = ".autopus/reverse/"

func countGitignoreLine(t *testing.T, dir, pattern string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	count := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == pattern {
			count++
		}
	}
	return count
}

// S3 (SPEC-REA-001): a new init ignores the behaviour-spec artifacts of
// /auto reverse exactly once, a later update does not repeat the line, and git
// keeps those third-party-derived documents local while SPEC documents stay
// trackable.
func TestInitGitignore_ReverseArtifactsStayLocal(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	require.NoError(t, exec.Command("git", "-C", dir, "init", "-q").Run())

	var out bytes.Buffer
	initCmd := newTestRootCmd()
	initCmd.SetOut(&out)
	initCmd.SetErr(&out)
	initCmd.SetArgs([]string{"init", "--dir", dir, "--yes", "--project", "reverse-gitignore",
		"--platforms", "claude-code,codex,antigravity-cli"})
	require.NoError(t, initCmd.Execute(), out.String())
	assert.Equal(t, 1, countGitignoreLine(t, dir, reverseArtifactPattern))

	out.Reset()
	updateCmd := newTestRootCmd()
	updateCmd.SetOut(&out)
	updateCmd.SetErr(&out)
	updateCmd.SetArgs([]string{"update", "--dir", dir})
	require.NoError(t, updateCmd.Execute(), out.String())
	assert.Equal(t, 1, countGitignoreLine(t, dir, reverseArtifactPattern))

	ignored := exec.Command("git", "-C", dir, "check-ignore", "-q",
		".autopus/reverse/login-flow/behaviour-spec.md")
	assert.NoError(t, ignored.Run(), "behaviour specs must stay out of git")
	tracked := exec.Command("git", "-C", dir, "check-ignore", "-q",
		".autopus/specs/SPEC-REA-001/spec.md")
	assert.Error(t, tracked.Run(), "SPEC documents must stay trackable")
}
