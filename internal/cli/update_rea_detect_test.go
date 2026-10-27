package cli

// SPEC-REA-001 T12 (acceptance S7 (7), (15)): a workspace update hints once
// per repository, and a platform the update newly detects gets the same REA
// action as the configured ones.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S7 (7): `--workspace --yes` runs every repository with --yes, so each
// undecided repository prints the hint once, nothing reads stdin, and no
// repository gets a rea entry or a key.
func TestUpdateRea_WorkspaceHintsOncePerRepo(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, true, "")
	root := reaOLD(t, reaAllTargets)
	child := filepath.Join(root, "service")
	data, err := os.ReadFile(filepath.Join(root, "autopus.yaml"))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(child, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(child, "autopus.yaml"), data, 0o644))
	for _, repo := range []string{root, child} {
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644))
	}

	out, err := runComposeUpdate(t, driverFailingReader{t}, "--dir", root, "--workspace", "--yes")

	require.NoError(t, err, out)
	assert.Equal(t, 2, strings.Count(out, reaTextKo.Hint), out)
	assert.Zero(t, strings.Count(out, reaTextKo.Question))
	for _, repo := range []string{root, child} {
		assertReaAbsent(t, repo)
		assert.Nil(t, reaKey(t, repo), repo)
	}
}

// S7 (15): with only claude-code configured and codex detected, --rea gives
// .mcp.json and the newly generated .codex/config.toml one managed entry
// each, and the entries carry the decision without a key. The run is
// interactive so the detected Codex generates from the composition step's
// catalog snapshot; the flag means no REA question, and Enter keeps the
// composition.
func TestUpdateRea_DetectedPlatformGetsTheAction(t *testing.T) {
	newComposeDetectRun(t, "codex")
	dir := writeComposeFixture(t, "claude-code")

	out, err := runComposeUpdate(t, strings.NewReader("\n"), "--dir", dir, "--rea")

	require.NoError(t, err, out)
	assert.Zero(t, strings.Count(out, reaTextKo.Question))
	assert.Contains(t, out, "새 플랫폼 감지: codex")
	assert.Contains(t, out, "REA 역공학 MCP를 등록했습니다: .mcp.json, .codex/config.toml (rea-agents@6.5.0)")
	for _, rel := range []string{".mcp.json", ".codex/config.toml"} {
		assert.FileExists(t, filepath.Join(dir, rel))
	}
	assert.Equal(t, 1, strings.Count(string(readReaFile(t, dir, "codex")), "[mcp_servers.rea]"))
	assert.Equal(t, "managed", string(reaFileClass(dir, "claude-code")))
	assert.Equal(t, "managed", string(reaFileClass(dir, "codex")))
	assert.Nil(t, reaKey(t, dir))
}
