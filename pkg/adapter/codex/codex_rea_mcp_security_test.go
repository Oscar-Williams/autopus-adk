package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// Security audit L4: the REA transform reads strings by TOML's rules, and
// every Codex text a REA action changes is re-validated before it is written.

func TestReaMCP_CodexDelimiterInOneLineStringKeepsUserTableSingle(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := reaCodexAdapter(root)
	_, err := a.Generate(context.Background(), reaCodexConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	userTable := codexReaX1 + "tool_timeout_sec = 120\n"
	writeReaCodexConfig(t, root, readReaCodexConfig(t, root)+"\n[mcp_servers.mine]\nx = \"it'''s\"\n\n"+userTable)

	_, err = a.Generate(context.Background(), reaCodexConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	text := readReaCodexConfig(t, root)
	assert.Equal(t, 1, strings.Count(text, "[mcp_servers.rea]"))
	assert.True(t, strings.HasSuffix(text, "\n\n"+userTable), text)
	assert.NoError(t, validateCodexTOMLStructure(text))
}

func TestReaMCP_CodexRemoveLeavesManagedLookingStringContent(t *testing.T) {
	t.Parallel()
	block := strings.TrimSuffix(codexReaX1, "\n")
	userTail := "\n[mcp_servers.mine]\nnote.text = \"\"\"\n" + block + "\n# \"\"\"\n"
	keepRoot, reaRoot := t.TempDir(), t.TempDir()
	for _, root := range []string{keepRoot, reaRoot} {
		_, err := reaCodexAdapter(root).Generate(context.Background(), reaCodexConfig(config.ReaMCPKeep))
		require.NoError(t, err)
		writeReaCodexConfig(t, root, readReaCodexConfig(t, root)+userTail)
	}
	_, err := reaCodexAdapter(keepRoot).Generate(context.Background(), reaCodexConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	_, err = reaCodexAdapter(reaRoot).Generate(context.Background(), reaCodexConfig(config.ReaMCPRemove))
	require.NoError(t, err)

	assert.Equal(t, readReaCodexConfig(t, keepRoot), readReaCodexConfig(t, reaRoot))
	assert.Contains(t, readReaCodexConfig(t, reaRoot), userTail)
}

func TestReaMCP_CodexRefusesAddThatFailsStructuralValidation(t *testing.T) {
	t.Parallel()
	// validateCodexTOMLStructure closes the string at \""" although TOML and
	// the REA scanner do not, so to the validator an appended rea table is a
	// duplicate. The merge admits the file; the REA write must not.
	original := "note = \"\"\"\na \\\"\"\"\n[mcp_servers.rea]\ncommand = \"x\"\n"
	require.NoError(t, validateCodexTOMLStructure(original))
	root := t.TempDir()
	writeReaCodexConfig(t, root, original)

	_, err := reaCodexAdapter(root).Generate(context.Background(), reaCodexConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Contains(t, err.Error(), ".codex/config.toml")
	assert.Contains(t, err.Error(), "duplicate table")
	assert.Equal(t, original, readReaCodexConfig(t, root))
}

func TestReaCodexTransform_ValidatesOnlyTextItChanged(t *testing.T) {
	t.Parallel()
	_, changed, err := reaCodexTransform("approval_policy\n\n"+codexReaX1, config.ReaMCPRemove)
	require.Error(t, err)
	assert.False(t, changed)
	assert.Contains(t, err.Error(), ".codex/config.toml")

	out, changed, err := reaCodexTransform("approval_policy\n", config.ReaMCPRemove)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, "approval_policy\n", out)
}

// reamcp.ApplyFile carries a copy of validateCodexTOMLStructure; on every
// output a REA add produces it must refuse exactly what the adapter refuses.
func TestReaCodexValidatorMatchesMerge(t *testing.T) {
	t.Parallel()
	corpus := []string{
		"a = 1\n", "[a]\nb = 2\n", "[[t]]\nx = 1\n[[t]]\nx = 2\n", "\"a.b\" = 1\n",
		"a = [\n1,\n2,\n]\n", "a = \"\"\"\nx\n\"\"\"\n", "a = { b = 1 } # c\n",
		"[a]\n[a]\n", "a = 1\na = 2\n", "[a\n", "a\n", "a$ = 1\n", "a =\n",
		"a = [1]]\n", "a = [1,\n", "a = \"x\" b = 1 }\n",
		"note = \"\"\"\na \\\"\"\"\n[mcp_servers.rea]\ncommand = \"x\"\n",
		"note = \"\"\"a\"\"\"\"\n",
	}
	for _, doc := range corpus {
		out, changed := reamcp.ApplyCodexTOML(doc, config.ReaMCPAdd)
		require.True(t, changed, doc)
		_, _, err := reamcp.ApplyFile(reamcp.Codex, []byte(doc), config.ReaMCPAdd)
		assert.Equal(t, validateCodexTOMLStructure(out) == nil, err == nil, doc)
	}
}
