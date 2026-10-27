package claude

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

// claudeReaC1 is the literal managed .mcp.json entry of the MCP Entry Shape
// Contract (acceptance S4 row C1).
const claudeReaC1 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`

// claudeReaC8 and claudeReaC9 are acceptance rows C8 and C9: the entry
// v0.50.127 wrote, without --prefix=/, with the managed pin and another pin.
const (
	claudeReaC8 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`
	claudeReaC9 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.4.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.4.0 mcp"]}`
)

func reaClaudeConfig(action config.ReaMCPAction) *config.HarnessConfig {
	cfg := config.DefaultFullConfig("rea")
	cfg.Runtime.ReaMCP = action
	return cfg
}

func decodeReaValue(t *testing.T, text string) any {
	t.Helper()
	var value any
	require.NoError(t, json.Unmarshal([]byte(text), &value))
	return value
}

func readReaFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

func mcpServers(t *testing.T, root string) map[string]any {
	t.Helper()
	servers, _ := readJSONObject(t, filepath.Join(root, ".mcp.json"))["mcpServers"].(map[string]any)
	return servers
}

func TestReaMCP_GenerateAddWritesOneManagedEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := NewWithRoot(root).Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)

	servers := mcpServers(t, root)
	assert.Equal(t, decodeReaValue(t, claudeReaC1), servers["rea"])
	assert.Contains(t, servers, "context7")
	assert.Contains(t, servers, "sequential-thinking")
	assert.Equal(t, 1, strings.Count(string(readReaFile(t, filepath.Join(root, ".mcp.json"))), `"rea": {`))
}

func TestReaMCP_KeepLeavesFileAsBaseMergeWritesIt(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	fresh := readReaFile(t, filepath.Join(root, ".mcp.json"))
	assert.NotContains(t, string(fresh), "rea-agents")

	// A managed entry an earlier run wrote survives a keep run unchanged.
	_, err = a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	added := readReaFile(t, filepath.Join(root, ".mcp.json"))
	_, err = a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	assert.Equal(t, added, readReaFile(t, filepath.Join(root, ".mcp.json")))
}

func TestReaMCP_AddThenRemoveRestoresBaseMergeBytes(t *testing.T) {
	t.Parallel()
	user := "{\n    \"mcpServers\": {\n        \"mine\": {\"note\": \"café\", \"command\": \"mine-mcp\"}\n    }\n}\n"
	keepRoot, reaRoot := t.TempDir(), t.TempDir()
	for _, root := range []string{keepRoot, reaRoot} {
		require.NoError(t, os.WriteFile(filepath.Join(root, ".mcp.json"), []byte(user), 0o644))
	}
	_, err := NewWithRoot(keepRoot).Generate(context.Background(), reaClaudeConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	base := readReaFile(t, filepath.Join(keepRoot, ".mcp.json"))

	a := NewWithRoot(reaRoot)
	_, err = a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	added := readReaFile(t, filepath.Join(reaRoot, ".mcp.json"))
	assert.Equal(t, decodeReaValue(t, `{"note": "café", "command": "mine-mcp"}`), mcpServers(t, reaRoot)["mine"])

	// The CLI's whole-file transform encodes exactly as this merge does.
	viaFile, changed, err := reamcp.ApplyFile(reamcp.Claude, base, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, string(added), string(viaFile))

	_, err = a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPRemove))
	require.NoError(t, err)
	assert.Equal(t, string(base), string(readReaFile(t, filepath.Join(reaRoot, ".mcp.json"))))
}

func TestReaMCP_MergeLeavesOtherPinAndUserOwnedEntriesValueIdentical(t *testing.T) {
	t.Parallel()
	rows := map[string]string{
		"C2": `{"$comment": "REA reverse-engineering MCP (rea-agents@6.4.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.4.0 mcp"]}`,
		"C3": `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}`,
		"C4": `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"], "env": {"REA_MCP_MAX_RESPONSE_BYTES": "1"}}`,
		"C5": `{"$comment": "mine", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`,
		"C7": `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.4.0 mcp"]}`,
		"C9": claudeReaC9,
	}
	for id, value := range rows {
		root := t.TempDir()
		writeJSONFile(t, filepath.Join(root, ".mcp.json"), map[string]any{"mcpServers": map[string]any{"rea": decodeReaValue(t, value)}})
		for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
			files, err := NewWithRoot(root).prepareMCPConfig(reaClaudeConfig(action))
			require.NoError(t, err, id)
			var doc map[string]any
			require.NoError(t, json.Unmarshal(files[0].Content, &doc), id)
			assert.Equal(t, decodeReaValue(t, value), doc["mcpServers"].(map[string]any)["rea"], "%s %s", id, action)
		}
	}

	// C6: rea-agents is another server, so add writes rea beside it.
	root := t.TempDir()
	writeJSONFile(t, filepath.Join(root, ".mcp.json"), map[string]any{"mcpServers": map[string]any{"rea-agents": decodeReaValue(t, claudeReaC1)}})
	files, err := NewWithRoot(root).prepareMCPConfig(reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal(files[0].Content, &doc))
	servers := doc["mcpServers"].(map[string]any)
	assert.Equal(t, decodeReaValue(t, claudeReaC1), servers["rea"])
	assert.Equal(t, decodeReaValue(t, claudeReaC1), servers["rea-agents"])
}

func TestReaMCP_CleanRemovesOnlyTheManagedEntry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	require.NoError(t, a.Clean(context.Background()))
	assert.NotContains(t, string(readReaFile(t, filepath.Join(root, ".mcp.json"))), "rea-agents")

	userRoot := t.TempDir()
	userRea := `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}`
	writeJSONFile(t, filepath.Join(userRoot, ".mcp.json"), map[string]any{"mcpServers": map[string]any{"rea": decodeReaValue(t, userRea)}})
	userAdapter := NewWithRoot(userRoot)
	_, err = userAdapter.Generate(context.Background(), reaClaudeConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	require.NoError(t, userAdapter.Clean(context.Background()))
	assert.Equal(t, decodeReaValue(t, userRea), mcpServers(t, userRoot)["rea"])
}

// A project holding the v0.50.127 entry (C8) gets the current entry from an
// add run and loses it on remove and Clean; a keep run leaves it as written.
func TestReaMCP_MovesTheRetiredEntry(t *testing.T) {
	t.Parallel()
	writeRetired := func(t *testing.T) string {
		root := t.TempDir()
		writeJSONFile(t, filepath.Join(root, ".mcp.json"), map[string]any{"mcpServers": map[string]any{
			"rea": decodeReaValue(t, claudeReaC8), "mine": map[string]any{"command": "mine-mcp"}}})
		return root
	}

	root := writeRetired(t)
	_, err := NewWithRoot(root).Generate(context.Background(), reaClaudeConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	assert.Equal(t, decodeReaValue(t, claudeReaC8), mcpServers(t, root)["rea"])
	_, err = NewWithRoot(root).Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	assert.Equal(t, decodeReaValue(t, claudeReaC1), mcpServers(t, root)["rea"])
	assert.Equal(t, map[string]any{"command": "mine-mcp"}, mcpServers(t, root)["mine"])

	root = writeRetired(t)
	_, err = NewWithRoot(root).Generate(context.Background(), reaClaudeConfig(config.ReaMCPRemove))
	require.NoError(t, err)
	assert.NotContains(t, mcpServers(t, root), "rea")

	root = writeRetired(t)
	require.NoError(t, NewWithRoot(root).Clean(context.Background()))
	assert.NotContains(t, string(readReaFile(t, filepath.Join(root, ".mcp.json"))), "rea-agents")
}

func TestReaMCP_RefusesSymlinkedTargetWhenReaChangesContent(t *testing.T) {
	t.Parallel()
	root, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "o3.json")
	original := []byte("{\"mcpServers\": {}}\n")
	require.NoError(t, os.WriteFile(target, original, 0o644))
	if err := os.Symlink(target, filepath.Join(root, ".mcp.json")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := NewWithRoot(root).Generate(context.Background(), reaClaudeConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Contains(t, err.Error(), ".mcp.json")
	assert.Equal(t, original, readReaFile(t, target))
}
