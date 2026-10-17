package antigravity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// platform remove takes the managed entry out of the plugin file, keeps a
// server the user added there, and deletes a file that held the entry alone;
// it also sweeps a legacy settings entry whatever the current render holds.
func TestReaMCP_AntigravityCleanRemovesOnlyTheManagedEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	path := antigravityPluginMCPPath(root)
	withMine := `{"mcpServers": {"mine": {"command": "mine-mcp"}, "rea": ` + antigravityReaG1 + `}}`
	require.NoError(t, os.WriteFile(path, []byte(withMine), 0o644))
	settingsPath := filepath.Join(root, ".gemini", "settings.json")
	var settings map[string]any
	require.NoError(t, json.Unmarshal(readAntigravityBytes(t, settingsPath), &settings))
	settings["mcpServers"].(map[string]any)["context7"] = map[string]any{"command": "user-context-server"}
	settings["mcpServers"].(map[string]any)["rea"] = decodeAntigravityValue(t, antigravityReaG1)
	raw, err := json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, raw, 0o644))

	require.NoError(t, a.Clean(context.Background()))
	assert.Equal(t, map[string]any{"mine": map[string]any{"command": "mine-mcp"}}, antigravityServers(t, readAntigravityBytes(t, path)))
	legacy := antigravityServers(t, readAntigravityBytes(t, settingsPath))
	assert.NotContains(t, legacy, "rea")
	assert.Equal(t, map[string]any{"command": "user-context-server"}, legacy["context7"])

	aloneRoot := t.TempDir()
	alone := NewWithRoot(aloneRoot)
	_, err = alone.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	require.NoError(t, alone.Clean(context.Background()))
	assert.NoFileExists(t, antigravityPluginMCPPath(aloneRoot))
}

// A REA change whose path crosses a symlink is refused before anything is
// written, and the file outside the project keeps its bytes.
func TestReaMCP_AntigravityRefusesSymlinkedTargetWhenReaChangesContent(t *testing.T) {
	t.Parallel()
	outside := t.TempDir()
	o1 := filepath.Join(outside, "o1.json")
	original := []byte("{\"mcpServers\": {}}\n")
	require.NoError(t, os.WriteFile(o1, original, 0o644))

	// (a) the plugin MCP file is a symlink: Generate refuses before writing.
	fileRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(fileRoot, antigravityPluginDir), 0o755))
	if err := os.Symlink(o1, antigravityPluginMCPPath(fileRoot)); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := NewWithRoot(fileRoot).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Contains(t, err.Error(), reamcp.TargetRelPath(reamcp.Antigravity))
	assert.Equal(t, original, readAntigravityBytes(t, o1))
	assert.NoFileExists(t, filepath.Join(fileRoot, "GEMINI.md"), "the refusal comes before any write")

	// (b) .agents is a symlinked directory: Update refuses before its transaction.
	o2 := t.TempDir()
	dirRoot := t.TempDir()
	require.NoError(t, os.Symlink(o2, filepath.Join(dirRoot, ".agents")))
	_, err = NewWithRoot(dirRoot).Update(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.NoFileExists(t, filepath.Join(o2, "plugins", "autopus", "mcp_config.json"))

	// Clean through a symlinked plugin file that holds a managed entry.
	cleanRoot := t.TempDir()
	a := NewWithRoot(cleanRoot)
	_, err = a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	managed := []byte(`{"mcpServers": {"rea": ` + antigravityReaG1 + `}}`)
	o3 := filepath.Join(outside, "o3.json")
	require.NoError(t, os.WriteFile(o3, managed, 0o644))
	require.NoError(t, os.Symlink(o3, antigravityPluginMCPPath(cleanRoot)))
	err = a.Clean(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Equal(t, managed, readAntigravityBytes(t, o3))

	// (c) a symlinked .gemini/settings.json holding a legacy managed entry.
	legacyRoot := t.TempDir()
	o4 := filepath.Join(outside, "o4.json")
	require.NoError(t, os.WriteFile(o4, managed, 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(legacyRoot, ".gemini"), 0o755))
	require.NoError(t, os.Symlink(o4, filepath.Join(legacyRoot, ".gemini", "settings.json")))
	_, err = NewWithRoot(legacyRoot).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".gemini/settings.json")
	assert.Equal(t, managed, readAntigravityBytes(t, o4))
}

// The plugin file is read through reamcp.ReadTarget: an oversized file is
// refused for add and remove alike, with an error naming the file, and its
// bytes stay.
func TestReaMCP_AntigravityRefusesAnOversizedPluginFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	big := `{"mcpServers": {"rea": ` + antigravityReaG1 + `}, "pad": "` +
		strings.Repeat("x", reamcp.TargetLimit) + `"}`
	path := writeAntigravityPluginMCP(t, root, big)
	for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
		_, _, err := NewWithRoot(root).planReaPluginMCP(action)
		require.Error(t, err, action)
		assert.Contains(t, err.Error(), reamcp.TargetRelPath(reamcp.Antigravity)+": larger than 1 MiB", action)
	}
	_, changed, err := NewWithRoot(root).planReaPluginMCP(config.ReaMCPKeep)
	require.NoError(t, err)
	assert.False(t, changed, "a keep run never reads the file")
	assert.Equal(t, big, string(readAntigravityBytes(t, path)))
}

// Every agy agent loads the REA guard: it is an always_on plugin rule, since
// agy applies only always_on rules unconditionally and an agent's tools list
// does not keep it from the autopus_rea server.
func TestReaMCP_AntigravityPluginCarriesTheAlwaysOnGuardRule(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := NewWithRoot(root).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	rule := string(readAntigravityBytes(t, filepath.Join(root, antigravityPluginDir, "rules", "rea-mcp-guard.md")))
	assert.Contains(t, rule, "\ntrigger: always_on\n")
	for _, want := range []string{
		"Only `auto-agent-reverse-engineer` and the main session running the `/auto reverse` flow call `autopus_rea` tools.",
		"`허용` (`allow`)", "once per call", "never choose an \"always allow\" option",
	} {
		assert.Contains(t, rule, want)
	}
}
