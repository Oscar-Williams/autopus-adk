package cli

// Antigravity CLI reads plugin MCP servers from <plugin>/mcp_config.json and
// never from a project .gemini/settings.json, so REA's Antigravity target is
// the workspace plugin's file (SPEC-REA-001, agy 1.4 smoke of 2026-10-17).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaPluginMCPRel is the Antigravity target file.
const reaPluginMCPRel = ".agents/plugins/autopus/mcp_config.json"

// A real init registers Antigravity in the plugin file, leaves the settings
// file without a rea entry, and doctor names the file each class was read
// from.
func TestReaPluginMCP_InitRegistersInThePluginAndDoctorNamesTheFile(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	dir := t.TempDir()
	out, err := reaInit(t, dir, reaAllTargets, "--yes")
	require.NoError(t, err, out)

	assert.Equal(t, reaPluginMCPRel, reamcp.TargetRelPath(reamcp.Antigravity))
	assertReaManagedOnce(t, dir)
	settings, err := os.ReadFile(filepath.Join(dir, ".gemini", "settings.json"))
	require.NoError(t, err)
	assert.NotContains(t, string(settings), "rea-agents")

	text, _ := runReaCLI(t, "", "doctor", "--dir", dir)
	for _, line := range []string{
		"claude-code (.mcp.json): managed",
		"codex (.codex/config.toml): managed",
		"antigravity-cli (" + reaPluginMCPRel + "): managed",
	} {
		assert.Equal(t, 1, strings.Count(text, line), line+"\n"+text)
	}
}

// The off sweep deletes the plugin file when the managed entry was all it
// held, keeps a server the user added there, and leaves a file it cannot
// decode alone.
func TestReaPluginMCP_SweepRemovesOnlyTheManagedEntry(t *testing.T) {
	t.Parallel()
	alone := t.TempDir()
	writeManagedRea(t, alone, reamcp.Antigravity)
	require.NoError(t, removeReaEntry(alone, reamcp.Antigravity))
	assert.NoFileExists(t, filepath.Join(alone, filepath.FromSlash(reaPluginMCPRel)))

	shared := t.TempDir()
	mine := []byte(`{"mcpServers": {"mine": {"command": "mine-mcp"}}}`)
	writeReaTarget(t, shared, reamcp.Antigravity, reaApplied(t, reamcp.Antigravity, mine, config.ReaMCPAdd))
	require.NoError(t, removeReaEntry(shared, reamcp.Antigravity))
	assert.Equal(t, string(reaApplied(t, reamcp.Antigravity, reaApplied(t, reamcp.Antigravity, mine, config.ReaMCPAdd), config.ReaMCPRemove)),
		string(readReaFile(t, shared, reamcp.Antigravity)))
	assert.NotContains(t, string(readReaFile(t, shared, reamcp.Antigravity)), "rea-agents")

	jsonc := t.TempDir()
	commented := []byte("{\n  // mine\n  \"mcpServers\": {},\n}\n")
	writeReaTarget(t, jsonc, reamcp.Antigravity, commented)
	require.NoError(t, removeReaEntry(jsonc, reamcp.Antigravity))
	assert.Equal(t, commented, readReaFile(t, jsonc, reamcp.Antigravity))
	assert.Equal(t, reamcp.ClassUserOwned, reaFileClass(jsonc, reamcp.Antigravity))
}

// A preview names the plugin file the run would create or delete, which no
// planned mapping shows, and a run that keeps the entry names nothing.
func TestReaPluginMCP_PreviewItems(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "antigravity-cli")
	cfg.Runtime.ReaMCP = config.ReaMCPAdd
	items := reaPluginPreviewItems(dir, cfg)
	require.Len(t, items, 1)
	assert.Equal(t, previewItem{Path: reaPluginMCPRel, Kind: "create", Scope: "antigravity-cli",
		Reason: "REA would register its managed rea entry in this plugin MCP file"}, items[0])

	writeManagedRea(t, dir, reamcp.Antigravity)
	assert.Empty(t, reaPluginPreviewItems(dir, cfg), "the current entry needs no write")
	cfg.Runtime.ReaMCP = config.ReaMCPRemove
	items = reaPluginPreviewItems(dir, cfg)
	require.Len(t, items, 1)
	assert.Equal(t, "remove", items[0].Kind)
	cfg.Runtime.ReaMCP = config.ReaMCPKeep
	assert.Empty(t, reaPluginPreviewItems(dir, cfg))
	cfg.Runtime.ReaMCP = config.ReaMCPRemove
	cfg.Platforms = []string{"claude-code"}
	assert.Empty(t, reaPluginPreviewItems(dir, cfg), "an unselected platform belongs to the off sweep")
}
