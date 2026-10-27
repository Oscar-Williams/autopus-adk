package antigravity

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// antigravityReaG1 is the literal managed entry of the MCP Entry Shape
// Contract (acceptance S4 row G1).
const antigravityReaG1 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`

// antigravityReaG4 is acceptance row G4: the entry v0.50.127 and the
// v0.50.126 release candidates wrote, which ran npx without --prefix=/.
const antigravityReaG4 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`

// antigravityPluginMCPCreated is the plugin MCP file an add run creates.
const antigravityPluginMCPCreated = "{\n  \"mcpServers\": {\n    \"rea\": {\n      \"args\": [\n        \"-l\",\n        \"-c\",\n" +
	"        \"exec npx -y --prefix=/ rea-agents@6.5.0 mcp\"\n      ],\n      \"command\": \"/bin/bash\"\n    }\n  }\n}\n"

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

// antigravityServers returns the mcpServers object of a settings or plugin
// MCP document.
func antigravityServers(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	require.NoError(t, json.Unmarshal(data, &doc))
	servers, _ := doc["mcpServers"].(map[string]any)
	return servers
}

func writeAntigravitySettings(t *testing.T, root, text string) string {
	t.Helper()
	return writeAntigravityFile(t, root, filepath.Join(".gemini", "settings.json"), text)
}

func writeAntigravityPluginMCP(t *testing.T, root, text string) string {
	t.Helper()
	return writeAntigravityFile(t, root, filepath.FromSlash(reamcp.TargetRelPath(reamcp.Antigravity)), text)
}

func writeAntigravityFile(t *testing.T, root, rel, text string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(text), 0o644))
	return path
}

func antigravityPluginMCPPath(root string) string {
	return filepath.Join(root, filepath.FromSlash(reamcp.TargetRelPath(reamcp.Antigravity)))
}

// agy reads plugin MCP servers from the plugin's own mcp_config.json and never
// from .gemini/settings.json: add creates the plugin file with the managed
// entry alone, settings get none, and the manifest does not own the file.
func TestReaMCP_AntigravityAddWritesThePluginMCPFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)

	path := antigravityPluginMCPPath(root)
	assert.Equal(t, antigravityPluginMCPCreated, string(readAntigravityBytes(t, path)))
	assert.Equal(t, filepath.Join(root, antigravityPluginDir, "mcp_config.json"), path)
	settings := antigravityServers(t, readAntigravityBytes(t, filepath.Join(root, ".gemini", "settings.json")))
	assert.NotContains(t, settings, "rea")
	assert.Contains(t, settings, "context7")
	manifest, err := adapter.LoadManifest(root, adapterName)
	require.NoError(t, err)
	assert.NotContains(t, manifest.Files, reamcp.TargetRelPath(reamcp.Antigravity))

	// A keep run leaves the file an earlier run wrote byte-identical.
	_, err = a.Update(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	assert.Equal(t, antigravityPluginMCPCreated, string(readAntigravityBytes(t, path)))
}

// Removing the entry an add created deletes the file again, and the settings
// bytes stay what a keep run writes (REQ-REA-04, REQ-REA-16).
func TestReaMCP_AntigravityAddThenRemoveRestoresTheBase(t *testing.T) {
	t.Parallel()
	keepRoot, reaRoot := t.TempDir(), t.TempDir()
	_, err := NewWithRoot(keepRoot).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	base := readAntigravityBytes(t, filepath.Join(keepRoot, ".gemini", "settings.json"))
	assert.NoFileExists(t, antigravityPluginMCPPath(keepRoot))

	a := NewWithRoot(reaRoot)
	_, err = a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	require.FileExists(t, antigravityPluginMCPPath(reaRoot))
	_, err = a.Update(context.Background(), reaAntigravityConfig(config.ReaMCPRemove))
	require.NoError(t, err)

	assert.NoFileExists(t, antigravityPluginMCPPath(reaRoot))
	assert.FileExists(t, filepath.Join(reaRoot, antigravityPluginDir, "plugin.json"))
	assert.Equal(t, string(base), string(readAntigravityBytes(t, filepath.Join(reaRoot, ".gemini", "settings.json"))))
}

// Servers a user added to the plugin file stay value-identical through add
// and remove, and the file stays once it holds anything but the entry.
func TestReaMCP_AntigravityKeepsServersTheUserAdded(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	user := "{\n    \"mcpServers\": {\n        \"mine\": {\"note\": \"café\", \"command\": \"mine-mcp\"}\n    }\n}\n"
	path := writeAntigravityPluginMCP(t, root, user)
	a := NewWithRoot(root)

	_, err := a.Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	added := readAntigravityBytes(t, path)
	viaFile, changed, err := reamcp.ApplyFile(reamcp.Antigravity, []byte(user), config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, string(viaFile), string(added))
	assert.Equal(t, decodeAntigravityValue(t, `{"note": "café", "command": "mine-mcp"}`), antigravityServers(t, added)["mine"])
	assert.Equal(t, decodeAntigravityValue(t, antigravityReaG1), antigravityServers(t, added)["rea"])

	_, err = a.Update(context.Background(), reaAntigravityConfig(config.ReaMCPRemove))
	require.NoError(t, err)
	removed := antigravityServers(t, readAntigravityBytes(t, path))
	assert.Equal(t, map[string]any{"mine": decodeAntigravityValue(t, `{"note": "café", "command": "mine-mcp"}`)}, removed)
}

// A plugin file v0.50.127 created for G4 alone is rewritten to the file a
// fresh add creates, kept by a keep run, and deleted by a remove run.
func TestReaMCP_AntigravityMovesTheRetiredEntry(t *testing.T) {
	t.Parallel()
	retired := `{"mcpServers": {"rea": ` + antigravityReaG4 + `}}`
	for _, tc := range []struct {
		action config.ReaMCPAction
		want   string // "" when the file is gone
	}{
		{config.ReaMCPAdd, antigravityPluginMCPCreated},
		{config.ReaMCPKeep, retired},
		{config.ReaMCPRemove, ""},
	} {
		root := t.TempDir()
		path := writeAntigravityPluginMCP(t, root, retired)
		_, err := NewWithRoot(root).Generate(context.Background(), reaAntigravityConfig(tc.action))
		require.NoError(t, err, tc.action)
		if tc.want == "" {
			assert.NoFileExists(t, path, tc.action)
			continue
		}
		assert.Equal(t, tc.want, string(readAntigravityBytes(t, path)), tc.action)
	}
}

// other_pin and user-owned rea entries in the plugin file are never touched.
func TestReaMCP_AntigravityLeavesOtherPinAndUserOwnedEntries(t *testing.T) {
	t.Parallel()
	rows := map[string]string{
		"G2": `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`,
		"G3": `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@latest mcp"]}`,
	}
	for id, value := range rows {
		root := t.TempDir()
		original := `{"mcpServers": {"rea": ` + value + `}}`
		path := writeAntigravityPluginMCP(t, root, original)
		for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
			_, changed, err := NewWithRoot(root).planReaPluginMCP(action)
			require.NoError(t, err, id)
			assert.False(t, changed, "%s %s", id, action)
		}
		_, err := NewWithRoot(root).Generate(context.Background(), reaAntigravityConfig(config.ReaMCPAdd))
		require.NoError(t, err, id)
		assert.Equal(t, original, string(readAntigravityBytes(t, path)), id)
	}
}

// A plugin file that does not decode (agy also reads comments and trailing
// commas) is user-owned: add, remove, and keep leave its bytes alone and none
// of them fails the run; the CLI prints the line that names it.
func TestReaMCP_AntigravityLeavesAMalformedPluginFile(t *testing.T) {
	t.Parallel()
	const malformed = "{\n  // mine\n  \"mcpServers\": {\"rea\": " + antigravityReaG1 + ",},\n}\n"
	root := t.TempDir()
	path := writeAntigravityPluginMCP(t, root, malformed)
	a := NewWithRoot(root)
	for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPKeep, config.ReaMCPRemove} {
		_, err := a.Generate(context.Background(), reaAntigravityConfig(action))
		require.NoError(t, err, action)
		assert.Equal(t, malformed, string(readAntigravityBytes(t, path)), action)
		_, err = a.Update(context.Background(), reaAntigravityConfig(action))
		require.NoError(t, err, action)
		assert.Equal(t, malformed, string(readAntigravityBytes(t, path)), action)
	}
	require.NoError(t, a.Clean(context.Background()))
	assert.Equal(t, malformed, string(readAntigravityBytes(t, path)))
}

// A managed rea entry release candidates left in .gemini/settings.json goes
// on any run, in the shape they wrote (G4) or the current one; other_pin and
// user-owned entries there stay.
func TestReaMCP_AntigravitySweepsTheLegacySettingsEntry(t *testing.T) {
	t.Parallel()
	for _, entry := range []string{antigravityReaG4, antigravityReaG1} {
		root := t.TempDir()
		path := writeAntigravitySettings(t, root, `{"mcpServers": {"rea": `+entry+`, "mine": {"command": "mine-mcp"}}}`)
		_, err := NewWithRoot(root).Update(context.Background(), reaAntigravityConfig(config.ReaMCPKeep))
		require.NoError(t, err)
		servers := antigravityServers(t, readAntigravityBytes(t, path))
		assert.NotContains(t, servers, "rea", entry)
		assert.Equal(t, map[string]any{"command": "mine-mcp"}, servers["mine"], entry)
		assert.NoFileExists(t, antigravityPluginMCPPath(root), "a keep run adds no entry")
	}

	userRea := `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}`
	userRoot := t.TempDir()
	userPath := writeAntigravitySettings(t, userRoot, `{"mcpServers": {"rea": `+userRea+`}}`)
	for _, action := range []config.ReaMCPAction{config.ReaMCPKeep, config.ReaMCPAdd, config.ReaMCPRemove} {
		files, err := NewWithRoot(userRoot).generateSettings(reaAntigravityConfig(action))
		require.NoError(t, err, action)
		assert.Equal(t, decodeAntigravityValue(t, userRea), antigravityServers(t, files[0].Content)["rea"], action)
	}
	assert.Equal(t, decodeAntigravityValue(t, userRea), antigravityServers(t, readAntigravityBytes(t, userPath))["rea"])
}
