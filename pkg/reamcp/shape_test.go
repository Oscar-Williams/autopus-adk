package reamcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// The JSON values below are the literal S4 rows of acceptance.md, decoded the
// way the adapters decode a file, so the oracle never comes from this package.
// C8, C9, and G4 hold the v0.50.127 shape, which ran npx without --prefix=/.
const (
	jsonC1 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`
	jsonC2 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.4.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.4.0 mcp"]}`
	jsonC3 = `{"command": "npx", "args": ["-y", "rea-agents@6.5.0", "mcp"]}`
	jsonC4 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"], "env": {"REA_MCP_MAX_RESPONSE_BYTES": "1"}}`
	jsonC5 = `{"$comment": "mine", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`
	jsonC7 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.4.0 mcp"]}`
	jsonC8 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`
	jsonC9 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.4.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.4.0 mcp"]}`
	jsonG1 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`
	jsonG2 = `{"$comment": "REA reverse-engineering MCP (rea-agents@6.5.0)", "command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp"]}`
	jsonG3 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@latest mcp"]}`
	jsonG4 = `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 mcp"]}`
)

func decodeValue(t *testing.T, text string) any {
	t.Helper()
	var value any
	require.NoError(t, json.Unmarshal([]byte(text), &value))
	return value
}

func serversWith(t *testing.T, name, text string) map[string]any {
	t.Helper()
	return map[string]any{
		"context7": decodeValue(t, `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y @upstash/context7-mcp@latest"]}`),
		name:       decodeValue(t, text),
	}
}

func TestClassifyJSON_S4Rows(t *testing.T) {
	cases := []struct {
		id       string
		platform Platform
		name     string
		value    string
		want     Class
	}{
		{"C1", Claude, "rea", jsonC1, ClassManaged},
		{"C2", Claude, "rea", jsonC2, ClassOtherPin},
		{"C3", Claude, "rea", jsonC3, ClassUserOwned},
		{"C4", Claude, "rea", jsonC4, ClassUserOwned},
		{"C5", Claude, "rea", jsonC5, ClassUserOwned},
		{"C6", Claude, "rea-agents", jsonC1, ClassAbsent},
		{"C7", Claude, "rea", jsonC7, ClassUserOwned},
		{"C8", Claude, "rea", jsonC8, ClassManaged},
		{"C9", Claude, "rea", jsonC9, ClassOtherPin},
		{"G1", Antigravity, "rea", jsonG1, ClassManaged},
		{"G2", Antigravity, "rea", jsonG2, ClassUserOwned},
		{"G3", Antigravity, "rea", jsonG3, ClassOtherPin},
		{"G4", Antigravity, "rea", jsonG4, ClassManaged},
		{"Claude without comment", Claude, "rea", jsonG1, ClassUserOwned},
		{"null value", Claude, "rea", `null`, ClassUserOwned},
		{"string value", Antigravity, "rea", `"npx"`, ClassUserOwned},
		{"pin with an extra argument", Antigravity, "rea", `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 --x mcp"]}`, ClassUserOwned},
		{"retired pin with an extra argument", Antigravity, "rea", `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y rea-agents@6.5.0 --x mcp"]}`, ClassUserOwned},
		{"another prefix", Antigravity, "rea", `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/tmp rea-agents@6.5.0 mcp"]}`, ClassUserOwned},
		{"empty pin", Antigravity, "rea", `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@ mcp"]}`, ClassUserOwned},
		{"fourth arg", Antigravity, "rea", `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.5.0 mcp", "x"]}`, ClassUserOwned},
		{"missing args", Antigravity, "rea", `{"command": "/bin/bash", "env": {}}`, ClassUserOwned},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			assert.Equal(t, tc.want, ClassifyJSON(tc.platform, serversWith(t, tc.name, tc.value)))
		})
	}
	assert.Equal(t, ClassAbsent, ClassifyJSON(Claude, nil))
	assert.Equal(t, ClassAbsent, ClassifyJSON(Antigravity, map[string]any{}))
}

func TestManagedJSONEntry_MatchesContractShape(t *testing.T) {
	assert.Equal(t, decodeValue(t, jsonC1), any(ManagedJSONEntry(Claude)))
	assert.Equal(t, decodeValue(t, jsonG1), any(ManagedJSONEntry(Antigravity)))
	assert.Nil(t, ManagedJSONEntry(Codex))
	assert.Equal(t, "6.5.0", Pin)
	assert.Equal(t, []string{"6.5.0"}, ManagedPins)
}

func TestApplyJSON_AddRemoveKeep(t *testing.T) {
	for _, tc := range []struct {
		platform Platform
		managed  string
	}{{Claude, jsonC1}, {Antigravity, jsonG1}} {
		t.Run(string(tc.platform), func(t *testing.T) {
			base := map[string]any{"mine": decodeValue(t, `{"note": "café", "command": "mine-mcp"}`)}

			added, changed := ApplyJSON(tc.platform, base, config.ReaMCPAdd)
			require.True(t, changed)
			assert.Equal(t, decodeValue(t, tc.managed), added["rea"])
			assert.Equal(t, base["mine"], added["mine"])
			assert.NotContains(t, base, "rea", "ApplyJSON must not edit its input")

			again, changed := ApplyJSON(tc.platform, added, config.ReaMCPAdd)
			assert.False(t, changed)
			assert.Equal(t, added, again)

			kept, changed := ApplyJSON(tc.platform, added, config.ReaMCPKeep)
			assert.False(t, changed)
			assert.Equal(t, added, kept)

			removed, changed := ApplyJSON(tc.platform, added, config.ReaMCPRemove)
			require.True(t, changed)
			assert.Equal(t, base, removed)
			assert.Contains(t, added, "rea", "ApplyJSON must not edit its input")

			_, changed = ApplyJSON(tc.platform, base, config.ReaMCPRemove)
			assert.False(t, changed)
		})
	}

	fromNil, changed := ApplyJSON(Claude, nil, config.ReaMCPAdd)
	require.True(t, changed)
	assert.Equal(t, decodeValue(t, jsonC1), fromNil["rea"])
	_, changed = ApplyJSON(Codex, map[string]any{}, config.ReaMCPAdd)
	assert.False(t, changed, "Codex has no JSON shape")
}

func TestApplyJSON_LeavesOtherPinAndUserOwnedValueIdentical(t *testing.T) {
	cases := []struct {
		id       string
		platform Platform
		name     string
		value    string
	}{
		{"C2", Claude, "rea", jsonC2}, {"C3", Claude, "rea", jsonC3}, {"C4", Claude, "rea", jsonC4},
		{"C5", Claude, "rea", jsonC5}, {"C7", Claude, "rea", jsonC7}, {"C9", Claude, "rea", jsonC9},
		{"G2", Antigravity, "rea", jsonG2}, {"G3", Antigravity, "rea", jsonG3},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			servers := serversWith(t, tc.name, tc.value)
			for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
				out, changed := ApplyJSON(tc.platform, servers, action)
				assert.False(t, changed, action)
				assert.Equal(t, decodeValue(t, tc.value), out["rea"], action)
			}
		})
	}

	// C6: a rea-agents server is a different name, so add inserts rea beside it.
	servers := serversWith(t, "rea-agents", jsonC1)
	out, changed := ApplyJSON(Claude, servers, config.ReaMCPAdd)
	require.True(t, changed)
	assert.Equal(t, ClassManaged, ClassifyJSON(Claude, out))
	assert.Equal(t, decodeValue(t, jsonC1), out["rea-agents"])
}

func TestApplyJSON_ReplacesOlderManagedPin(t *testing.T) {
	previous := ManagedPins
	ManagedPins = []string{"6.4.0", "6.5.0"}
	t.Cleanup(func() { ManagedPins = previous })

	for _, tc := range []struct {
		platform Platform
		older    string
		current  string
	}{
		{Claude, jsonC2, jsonC1},
		{Antigravity, `{"command": "/bin/bash", "args": ["-l", "-c", "exec npx -y --prefix=/ rea-agents@6.4.0 mcp"]}`, jsonG1},
		// An older pin in the retired shape moves to both at once.
		{Claude, jsonC9, jsonC1},
	} {
		servers := serversWith(t, "rea", tc.older)
		require.Equal(t, ClassManaged, ClassifyJSON(tc.platform, servers))
		out, changed := ApplyJSON(tc.platform, servers, config.ReaMCPAdd)
		require.True(t, changed)
		assert.Equal(t, decodeValue(t, tc.current), out["rea"])
	}
}

// The v0.50.127 entry is Autopus-managed: add replaces it with the current
// shape, remove deletes it, and keep leaves it, as for an older managed pin.
func TestApplyJSON_MovesTheRetiredShapeToTheCurrentOne(t *testing.T) {
	for _, tc := range []struct {
		platform Platform
		retired  string
		current  string
	}{{Claude, jsonC8, jsonC1}, {Antigravity, jsonG4, jsonG1}} {
		t.Run(string(tc.platform), func(t *testing.T) {
			servers := serversWith(t, "rea", tc.retired)
			require.Equal(t, ClassManaged, ClassifyJSON(tc.platform, servers))

			added, changed := ApplyJSON(tc.platform, servers, config.ReaMCPAdd)
			require.True(t, changed)
			assert.Equal(t, decodeValue(t, tc.current), added["rea"])
			assert.Equal(t, servers["context7"], added["context7"])
			assert.Equal(t, decodeValue(t, tc.retired), servers["rea"], "ApplyJSON must not edit its input")
			_, changed = ApplyJSON(tc.platform, added, config.ReaMCPAdd)
			assert.False(t, changed, "the current shape is final")

			kept, changed := ApplyJSON(tc.platform, servers, config.ReaMCPKeep)
			assert.False(t, changed)
			assert.Equal(t, servers, kept)

			removed, changed := ApplyJSON(tc.platform, servers, config.ReaMCPRemove)
			require.True(t, changed)
			assert.Equal(t, map[string]any{"context7": servers["context7"]}, removed)
		})
	}
}

func TestPlatformTargets(t *testing.T) {
	assert.Equal(t, []Platform{Claude, Codex, Antigravity}, Platforms())
	assert.Equal(t, ".mcp.json", TargetRelPath(Claude))
	assert.Equal(t, ".codex/config.toml", TargetRelPath(Codex))
	// agy reads plugin MCP servers from <plugin>/mcp_config.json, never from a
	// project .gemini/settings.json, which release candidates wrote.
	assert.Equal(t, ".agents/plugins/autopus/mcp_config.json", TargetRelPath(Antigravity))
	assert.Equal(t, "", TargetRelPath(Platform("omp")))
	assert.Equal(t, ".gemini/settings.json", LegacyTargetRelPath(Antigravity))
	assert.Equal(t, "", LegacyTargetRelPath(Claude))
	assert.Equal(t, "", LegacyTargetRelPath(Codex))
	assert.Equal(t, Platform("claude-code"), Claude)
	assert.Equal(t, Platform("antigravity-cli"), Antigravity)
	assert.Equal(t, Platform("codex"), Codex)
}
