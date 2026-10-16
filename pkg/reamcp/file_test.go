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
        "exec npx -y rea-agents@6.5.0 mcp"
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
        "exec npx -y rea-agents@6.5.0 mcp"
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
	assert.Contains(t, err.Error(), ".gemini/settings.json")

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
