package cli

// The Claude permission ledger is ignored like the other local records, sync
// verify keeps classifying it generated/runtime, and update names a ledger git
// still tracks instead of untracking it.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
)

func TestGitignorePatterns_IgnoreThePermissionLedger(t *testing.T) {
	t.Parallel()
	count := 0
	for _, pattern := range gitignorePatterns {
		if pattern == permissionLedgerRel {
			count++
		}
	}
	assert.Equal(t, 1, count, "gitignorePatterns holds the ledger exactly once")
	assert.True(t, isGeneratedRuntime(permissionLedgerRel), "sync verify classifies the ledger generated/runtime")
	families := classifyTrackedIgnored([]string{permissionLedgerRel})
	require.Len(t, families, 1)
	assert.Equal(t, trackedIgnoredUntrack, families[0].Disposition, "doctor offers untracking, not a rule change")
}

func runLedgerGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s: %s", strings.Join(args, " "), out)
}

// The managed .gitignore hides a new ledger from git, so update prints no
// hint; once git tracks the ledger, update prints the one hint line on every
// run and leaves the index alone; outside a repository there is no hint.
func TestApplyGitignoreUpdate_HintsAtATrackedLedger(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	dir := t.TempDir()
	runLedgerGit(t, dir, "init", "-q")
	ledger := filepath.Join(dir, filepath.FromSlash(permissionLedgerRel))
	require.NoError(t, os.MkdirAll(filepath.Dir(ledger), 0o755))
	require.NoError(t, os.WriteFile(ledger, []byte("{\"version\": 1}\n"), 0o600))
	hint := "  " + permissionLedgerTrackedHint + "\n"

	var out bytes.Buffer
	require.NoError(t, applyGitignoreUpdate(&out, dir))
	assert.NotContains(t, out.String(), hint, "an untracked ledger needs no hint")
	runLedgerGit(t, dir, "check-ignore", "-q", "--", permissionLedgerRel)

	runLedgerGit(t, dir, "add", "-f", "--", permissionLedgerRel)
	for run := 0; run < 2; run++ {
		out.Reset()
		require.NoError(t, applyGitignoreUpdate(&out, dir))
		assert.Equal(t, hint, out.String(), "run %d", run)
	}
	runLedgerGit(t, dir, "ls-files", "--error-unmatch", "--", permissionLedgerRel)

	out.Reset()
	require.NoError(t, applyGitignoreUpdate(&out, t.TempDir()))
	assert.NotContains(t, out.String(), hint, "outside a repository")
	assert.Equal(t, "참고: .autopus/claude-code-permissions.json은 이제 .gitignore 대상인데 git이 아직 추적합니다. "+
		"로컬 파일은 두고 추적만 멈추려면 git rm --cached .autopus/claude-code-permissions.json", permissionLedgerTrackedHint)
}

// The update preview resolves a deleted ledger the way the Claude update does,
// so it lists the write instead of promising the file stays deleted.
func TestResolvePreviewAction_FollowsTheClaudeUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifest := adapter.NewManifest("claude-code")
	manifest.Files[permissionLedgerRel] = adapter.ManifestFile{Checksum: "x", Policy: adapter.OverwriteMerge}
	ledger := adapter.FileMapping{TargetPath: permissionLedgerRel, OverwritePolicy: adapter.OverwriteMerge}

	assert.Equal(t, adapter.ActionCreate, resolvePreviewAction(dir, "claude-code", ledger, manifest))
	assert.Equal(t, adapter.ActionSkip, resolvePreviewAction(dir, "codex", ledger, manifest))
}
