package reamcp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// codexX3 is a user-owned rea table (acceptance S4 row X3).
const codexX3 = codexX1 + "tool_timeout_sec = 120\n"

// A multiline delimiter inside a one-line string opens nothing, so the rea
// table after it is real (security audit L4, exploit 1: a missed table made
// add append a duplicate [mcp_servers.rea]).
func TestCodexScan_DelimiterInsideOneLineStringOpensNothing(t *testing.T) {
	lines := []string{
		`x = "it'''s"`,
		`x = 'say """'`,
		`x = "a \" '''"`,
		`x = ["'''", '"""']`,
		`"it'''s" = 1`,
		`x = "#" # '''`,
	}
	for _, line := range lines {
		head := codexBase + "\n[mcp_servers.mine]\n" + line + "\n\n"

		user := head + codexX3
		assert.Equal(t, ClassUserOwned, ClassifyCodexTOML(user), line)
		out, changed := ApplyCodexTOML(user, config.ReaMCPAdd)
		assert.False(t, changed, line)
		assert.Equal(t, user, out, line)

		managed := head + codexX1
		assert.Equal(t, ClassManaged, ClassifyCodexTOML(managed), line)
		out, changed = ApplyCodexTOML(managed, config.ReaMCPRemove)
		require.True(t, changed, line)
		assert.Equal(t, strings.TrimSuffix(head, "\n"), out, line)
	}
}

// A managed-looking block inside a multiline string is string content
// whatever opens the string, so remove never edits it (exploit 2).
func TestCodexScan_ManagedLookingBlockInsideAnyMultilineString(t *testing.T) {
	block := strings.TrimSuffix(codexX1, "\n")
	bodies := map[string]string{
		"dotted key":         "note.text = \"\"\"\n" + block + "\n# \"\"\"\n",
		"quoted dotted key":  "\"a.b\".c = \"\"\"\n" + block + "\n# \"\"\"\n",
		"literal dotted key": "note.text = '''\n" + block + "\n# '''\n",
		"array element":      "notes = [\n\"\"\"\n" + block + "\n# \"\"\",\n]\n",
		"reopened":           "notes = [\"\"\"\na\"\"\", \"\"\"\n" + block + "\n# \"\"\"]\n",
		"reopened after run": "notes = [\"\"\"\na\"\"\"\"\", \"\"\"\n" + block + "\n# \"\"\"]\n",
	}
	for name, body := range bodies {
		text := codexBase + "\n[mcp_servers.mine]\n" + body
		assert.Equal(t, ClassAbsent, ClassifyCodexTOML(text), name)

		out, changed := ApplyCodexTOML(text, config.ReaMCPRemove)
		assert.False(t, changed, name)
		assert.Equal(t, text, out, name)

		out, changed = ApplyCodexTOML(text, config.ReaMCPAdd)
		require.True(t, changed, name)
		assert.Equal(t, text+"\n"+codexX1, out, name)
		assert.NoError(t, tomlValidateStructure(out), name)
	}
}

// codexEscapedDelimiter passes the structural validator, which closes a basic
// multiline string at \""" although TOML does not, so the validator sees a
// rea table that the scanner (rightly) reads as string content.
const codexEscapedDelimiter = "note = \"\"\"\na \\\"\"\"\n[mcp_servers.rea]\ncommand = \"x\"\n"

func TestApplyFile_CodexRefusesOutputThatFailsStructuralValidation(t *testing.T) {
	require.NoError(t, tomlValidateStructure(codexEscapedDelimiter))
	require.Equal(t, ClassAbsent, ClassifyCodexTOML(codexEscapedDelimiter))

	for name, tc := range map[string]struct {
		data   string
		action config.ReaMCPAction
	}{
		"duplicate table after add": {codexEscapedDelimiter, config.ReaMCPAdd},
		"invalid line before add":   {"approval_policy\n", config.ReaMCPAdd},
		"invalid line after remove": {"approval_policy\n\n" + codexX1, config.ReaMCPRemove},
	} {
		out, changed, err := ApplyFile(Codex, []byte(tc.data), tc.action)
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), ".codex/config.toml", name)
		assert.False(t, changed, name)
		assert.Nil(t, out, name)
	}

	// An action that changes nothing is not validated: keep stays a no-op.
	out, changed, err := ApplyFile(Codex, []byte("approval_policy\n"), config.ReaMCPRemove)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, []byte("approval_policy\n"), out)
}

func TestTOMLValidateStructure(t *testing.T) {
	valid := []string{
		"", "a = 1\n", "[a]\nb = 2\n", "[[t]]\nx = 1\n[[t]]\nx = 2\n", "\"a.b\" = 1\n",
		"a = [\n1,\n2,\n]\n", "a = \"\"\"\nx\n\"\"\"\n", "a = '''\n'''\n", "a = { b = 1 } # c\n",
		"[a]\nb = 1\n[c]\nb = 1\n", codexBase + "\n" + codexX1,
	}
	for _, content := range valid {
		assert.NoError(t, tomlValidateStructure(content), content)
	}
	invalid := map[string]string{
		"[a]\nb = 1\n[a]\n":   "line 3: duplicate table",
		"a = 1\na = 2\n":      "line 2: duplicate key",
		"[a\n":                "line 1: invalid table header",
		"a\n":                 "line 1: invalid assignment",
		"a$ = 1\n":            "line 1: invalid assignment",
		"a =\n":               "line 1: missing value",
		"a = [1]]\n":          "line 1: unbalanced value",
		"a = [\n1]]\n":        "line 2: unbalanced value",
		"a = [1,\n":           "unterminated value",
		"a = \"\"\"\nopen\n":  "unterminated value",
		"a = \"x\" b = 1 }\n": "line 1: unbalanced value",
	}
	for content, want := range invalid {
		err := tomlValidateStructure(content)
		require.Error(t, err, content)
		assert.Equal(t, want, err.Error(), content)
	}
}
