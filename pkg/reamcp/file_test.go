package reamcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// claudeOnlyRea is the exact file the Claude adapter's merge would write for a
// document holding only the managed entry: two-space indent, sorted keys, and
// a trailing newline.
const claudeOnlyRea = `{
  "mcpServers": {
    "rea": {
      "$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)",
      "args": [
        "-l",
        "-c",
        "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"
      ],
      "command": "/bin/bash"
    }
  }
}
`

const antigravityWithRea = `{
  "hooks": {},
  "mcpServers": {
    "context7": {
      "command": "npx"
    },
    "rea": {
      "args": [
        "-l",
        "-c",
        "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"
      ],
      "command": "/bin/bash"
    }
  }
}
`

const antigravityWithoutRea = `{
  "hooks": {},
  "mcpServers": {
    "context7": {
      "command": "npx"
    }
  }
}
`

func TestClassifyFile_MissingAndPresentFiles(t *testing.T) {
	for _, platform := range Platforms() {
		class, err := ClassifyFile(platform, nil)
		require.NoError(t, err, platform)
		assert.Equal(t, ClassAbsent, class, platform)
	}

	class, err := ClassifyFile(Claude, []byte(claudeOnlyRea))
	require.NoError(t, err)
	assert.Equal(t, ClassManaged, class)

	class, err = ClassifyFile(Antigravity, []byte(antigravityWithoutRea))
	require.NoError(t, err)
	assert.Equal(t, ClassAbsent, class)

	class, err = ClassifyFile(Codex, []byte(codexBase+"\n"+codexX1))
	require.NoError(t, err)
	assert.Equal(t, ClassManaged, class)

	class, err = ClassifyFile(Claude, []byte("null"))
	require.NoError(t, err)
	assert.Equal(t, ClassAbsent, class)
}

func TestClassifyFile_RejectsUnreadableJSON(t *testing.T) {
	_, err := ClassifyFile(Claude, []byte("{not json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".mcp.json")

	_, err = ClassifyFile(Antigravity, []byte(`{"mcpServers": "x"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".agents/plugins/autopus/mcp_config.json")

	// agy accepts comments and trailing commas there; Autopus refuses the file
	// rather than re-encode it without them.
	_, _, err = ApplyFile(Antigravity, []byte("{\n  // mine\n  \"mcpServers\": {},\n}\n"), config.ReaMCPAdd)
	require.Error(t, err)

	_, err = ClassifyFile(Platform("omp"), nil)
	require.Error(t, err)
}

func TestApplyFile_JSONEncodesLikeTheAdapterMerge(t *testing.T) {
	out, changed, err := ApplyFile(Claude, nil, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, claudeOnlyRea, string(out))

	compact := []byte(`{"hooks":{},"mcpServers":{"context7":{"command":"npx"}}}`)
	out, changed, err = ApplyFile(Antigravity, compact, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, antigravityWithRea, string(out))

	out, changed, err = ApplyFile(Antigravity, out, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, antigravityWithoutRea, string(out))
}

// A file holding the v0.50.127 entry comes out of add exactly as a fresh add
// writes it, and out of remove without the entry.
func TestApplyFile_MovesTheRetiredShape(t *testing.T) {
	retired := []byte(`{"mcpServers": {"rea": ` + jsonC8 + `}}`)
	out, changed, err := ApplyFile(Claude, retired, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, claudeOnlyRea, string(out))

	retired = []byte(`{"hooks": {}, "mcpServers": {"context7": {"command": "npx"}, "rea": ` + jsonG4 + `}}`)
	out, changed, err = ApplyFile(Antigravity, retired, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, antigravityWithRea, string(out))
	out, changed, err = ApplyFile(Antigravity, retired, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, antigravityWithoutRea, string(out))

	out, changed, err = ApplyFile(Codex, []byte(codexBase+"\n"+codexX8), config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, codexBase+"\n"+codexX1, string(out))
}

// The plugin file Autopus creates for the entry alone goes when the entry
// goes; a file with another server or key, and every other platform's file,
// stays.
func TestRemovesFile_OnlyTheEmptiedAntigravityPluginFile(t *testing.T) {
	created, changed, err := ApplyFile(Antigravity, nil, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	emptied, changed, err := ApplyFile(Antigravity, created, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.True(t, RemovesFile(Antigravity, emptied))
	assert.False(t, RemovesFile(Claude, emptied), "only the plugin file is Autopus-created")
	assert.False(t, RemovesFile(Antigravity, created))

	withMine := []byte(`{"mcpServers": {"mine": {"command": "mine-mcp"}, "rea": ` + jsonG1 + `}}`)
	kept, changed, err := ApplyFile(Antigravity, withMine, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.False(t, RemovesFile(Antigravity, kept))
	withKey := []byte(`{"note": "mine", "mcpServers": {"rea": ` + jsonG1 + `}}`)
	kept, changed, err = ApplyFile(Antigravity, withKey, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.False(t, RemovesFile(Antigravity, kept))
	assert.False(t, RemovesFile(Antigravity, []byte("{not json")))
}

func TestApplyFile_UnchangedFilesReturnTheirOwnBytes(t *testing.T) {
	userOwned := []byte("{\n    \"mcpServers\": {\"rea\": {\"command\": \"npx\"}}\n}")
	for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove, config.ReaMCPKeep} {
		out, changed, err := ApplyFile(Claude, userOwned, action)
		require.NoError(t, err, action)
		assert.False(t, changed, action)
		assert.Equal(t, userOwned, out, action)
	}

	out, changed, err := ApplyFile(Claude, []byte("{not json"), config.ReaMCPKeep)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, []byte("{not json"), out)

	_, _, err = ApplyFile(Claude, []byte("{not json"), config.ReaMCPRemove)
	require.Error(t, err)

	_, _, err = ApplyFile(Platform("opencode"), nil, config.ReaMCPAdd)
	require.Error(t, err)
}

func TestApplyFile_CodexMatchesTextTransform(t *testing.T) {
	out, changed, err := ApplyFile(Codex, nil, config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, codexX1, string(out))

	data := []byte(codexBase + "\n" + codexX1)
	out, changed, err = ApplyFile(Codex, data, config.ReaMCPRemove)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, codexBase, string(out))

	out, changed, err = ApplyFile(Codex, data, config.ReaMCPAdd)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, data, out)
}
