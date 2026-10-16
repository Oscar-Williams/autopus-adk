package reamcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTOMLKeyPath(t *testing.T) {
	valid := map[string][]string{
		"mcp_servers.rea":          {"mcp_servers", "rea"},
		` mcp_servers . "rea" `:    {"mcp_servers", "rea"},
		`'mcp_servers'.'rea'.env`:  {"mcp_servers", "rea", "env"},
		`"a\"b".c`:                 {`a"b`, "c"},
		`plugins."browser@openai"`: {"plugins", "browser@openai"},
	}
	for raw, want := range valid {
		got, ok := tomlKeyPath(raw)
		assert.True(t, ok, raw)
		assert.Equal(t, want, got, raw)
	}
	for _, raw := range []string{"", "a.", "a b", `"unterminated`, `'x`, "a..b", `"a"b`, `"\q"`} {
		_, ok := tomlKeyPath(raw)
		assert.False(t, ok, raw)
	}
}

func TestTOMLValueWithoutComment(t *testing.T) {
	for value, want := range map[string]string{
		`"a#b" # c`:        `"a#b" `,
		`'a#b' # c`:        `'a#b' `,
		`"a\"#" #c`:        `"a\"#" `,
		`"""x#""" # y`:     `"""x#""" `,
		`'''x#''' #`:       `'''x#''' `,
		`"""a\"""#""" # z`: `"""a\"""#""" `,
		`30`:               `30`,
	} {
		assert.Equal(t, want, tomlValueWithoutComment(value), value)
	}
}

func TestScanCodex_EscapedDelimiterKeepsMultilineOpen(t *testing.T) {
	text := codexBase + "\n[mcp_servers.mine]\nnote = \"\"\"\nnot closed \\\"\"\"\n" + codexX1 + "\"\"\"\n"
	assert.Equal(t, ClassAbsent, ClassifyCodexTOML(text))

	oneLine := codexBase + "\n[mcp_servers.mine]\nnote = \"\"\"one line\"\"\"\n\n" + codexX1
	assert.Equal(t, ClassManaged, ClassifyCodexTOML(oneLine))
}
