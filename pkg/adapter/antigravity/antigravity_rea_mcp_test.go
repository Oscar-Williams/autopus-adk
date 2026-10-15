package antigravity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// antigravityReaG1 is the literal managed entry of the MCP Entry Shape
// Contract (acceptance S4 row G1).
const antigravityReaG1 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`

func reaAntigravityConfig(action config.ReaMCPAction) *config.HarnessConfig {
	cfg := config.DefaultFullConfig("rea")
	cfg.Runtime.ReaMCP = action
	return cfg
}

func decodeAntigravityValue(t *testing.T, text string) any {
	t.Helper()
	var value any
	require.NoError(t, json.Unmarshal([]byte(text), &value))
	return value
}

func readAntigravityBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func antigravityServers(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))
	servers, _ := settings["mcpServers"].(map[string]any)
	return servers
}

func writeAntigravitySettings(t *testing.T, root, text string) string {
	t.Helper()
	path := filepath.Join(root, ".gemini", "settings.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(text), 0o644))
	return path
}

func TestReaMCP_AntigravityGenerateAddWritesManagedEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	path := filepath.Join(root, ".gemini", "settings.json")
	added := readAntigravityBytes(t, path)
	servers := antigravityServers(t, added)
	assert.Equal(t, decodeAntigravityValue(t, antigravityReaG1), servers["rea"])
	assert.Contains(t, servers, "context7")

	// A keep run leaves the managed entry an earlier run wrote byte-identical.
	_, err = a.Update(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	assert.Equal(t, string(added), string(readAntigravityBytes(t, path)))
}

func TestReaMCP_AntigravityAddThenRemoveRestoresBaseMergeBytes(t *testing.T) {
	t.Parallel()
	user := "{\n    \"mcpServers\": {\n        \"mine\": {\"note\": \"café\", \"command\": \"mine-mcp\"}\n    }\n}\n"
	keepRoot, reaRoot := t.TempDir(), t.TempDir()
	writeAntigravitySettings(t, keepRoot, user)
	path := writeAntigravitySettings(t, reaRoot, user)

	_, err := NewWithRoot(keepRoot).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	base := readAntigravityBytes(t, filepath.Join(keepRoot, ".gemini", "settings.json"))
	assert.NotContains(t, string(base), "rea-agents")

	a := NewWithRoot(reaRoot)
	_, err = a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	added := readAntigravityBytes(t, path)
	assert.Equal(t, decodeAntigravityValue(t, `{"note": "café", "command": "mine-mcp"}`), antigravityServers(t, added)["mine"])

	viaFile, changed, err := reamcp.ApplyFile(reamcp.Antigravity, base, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, string(added), string(viaFile))

	// The merge alone cannot delete a key; remove deletes mcpServers.rea.
	_, err = a.Update(context.Background(), reaAntigravityConfig(config.ReaMCPRemove))
	require.NoError(t, err)
	assert.Equal(t, string(base), string(readAntigravityBytes(t, path)))
}

func TestReaMCP_AntigravityMergeLeavesOtherPinAndUserOwnedEntriesValueIdentical(t *testing.T) {
	t.Parallel()
	rows := map[string]string{
		"G2": `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`,
		"G3": `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@latest mcp"]}`,
	}
	for id, value := range rows {
		root := t.TempDir()
		writeAntigravitySettings(t, root, `{"mcpServers": {"rea": `+value+`}}`)
		for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
			files, err := NewWithRoot(root).generateSettings(reaAntigravityConfig(action))
			require.NoError(t, err, id)
			assert.Equal(t, decodeAntigravityValue(t, value), antigravityServers(t, files[0].Content)["rea"], "%s %s", id, action)
		}
	}
}

func TestReaMCP_AntigravityCleanRemovesManagedEntryRegardlessOfRender(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	path := filepath.Join(root, ".gemini", "settings.json")
	var settings map[string]any
	require.NoError(t, json.Unmarshal(readAntigravityBytes(t, path), &settings))
	// A user-edited context7 is no longer the current render; rea still goes.
	settings["mcpServers"].(map[string]any)["context7"] = map[string]any{"command": "user-context-server"}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, raw, 0o644))

	require.NoError(t, a.Clean(context.Background()))
	servers := antigravityServers(t, readAntigravityBytes(t, path))
	assert.NotContains(t, servers, "rea")
	assert.Equal(t, map[string]any{"command": "user-context-server"}, servers["context7"])

	userRoot := t.TempDir()
	userAdapter := NewWithRoot(userRoot)
	_, err = userAdapter.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	userPath := filepath.Join(userRoot, ".gemini", "settings.json")
	require.NoError(t, json.Unmarshal(readAntigravityBytes(t, userPath), &settings))
	userRea := `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}`
	settings["mcpServers"].(map[string]any)["rea"] = decodeAntigravityValue(t, userRea)
	raw, err = json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(userPath, raw, 0o644))
	require.NoError(t, userAdapter.Clean(context.Background()))
	assert.Equal(t, decodeAntigravityValue(t, userRea), antigravityServers(t, readAntigravityBytes(t, userPath))["rea"])
}

func TestReaMCP_AntigravityRefusesSymlinkedTargetWhenReaChangesContent(t *testing.T) {
	t.Parallel()
	outside := t.TempDir()
	o1 := filepath.Join(outside, "o1.json")
	original := []byte("{\"mcpServers\": {}}\n")
	require.NoError(t, os.WriteFile(o1, original, 0o644))

	// (a) .gemini/settings.json is a symlink: Generate writes settings directly.
	fileRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(fileRoot, ".gemini"), 0o755))
	if err := os.Symlink(o1, filepath.Join(fileRoot, ".gemini", "settings.json")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := NewWithRoot(fileRoot).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Contains(t, err.Error(), ".gemini/settings.json")
	assert.Equal(t, original, readAntigravityBytes(t, o1))

	// (b) .gemini is a symlinked directory: Update refuses before its transaction.
	o2 := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(o2, "settings.json"), original, 0o644))
	dirRoot := t.TempDir()
	require.NoError(t, os.Symlink(o2, filepath.Join(dirRoot, ".gemini")))
	_, err = NewWithRoot(dirRoot).Update(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Equal(t, original, readAntigravityBytes(t, filepath.Join(o2, "settings.json")))

	// Clean through a symlinked settings file that holds a managed entry.
	cleanRoot := t.TempDir()
	a := NewWithRoot(cleanRoot)
	_, err = a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	managed := []byte(`{"mcpServers": {"rea": ` + antigravityReaG1 + `}}`)
	o3 := filepath.Join(outside, "o3.json")
	require.NoError(t, os.WriteFile(o3, managed, 0o644))
	settingsPath := filepath.Join(cleanRoot, ".gemini", "settings.json")
	require.NoError(t, os.Remove(settingsPath))
	require.NoError(t, os.Symlink(o3, settingsPath))
	err = a.Clean(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Equal(t, managed, readAntigravityBytes(t, o3))
}
