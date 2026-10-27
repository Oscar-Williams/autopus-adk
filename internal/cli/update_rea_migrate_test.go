package cli

// SPEC-REA-001 MCP Entry Shape Contract (acceptance S4 rows C8, G4, X8): a
// project v0.50.127 registered holds entries that ran npx without
// --prefix=/. The real `auto update` moves them to the current shape, and
// --no-rea and `auto platform remove` take them out.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// retireReaShape rewrites the current managed entries of a target file into
// the v0.50.127 shape and changes no other byte.
func retireReaShape(text string) string {
	text = strings.ReplaceAll(text, "exec npx -y --prefix=/ rea-agents@", "exec npx -y rea-agents@")
	return strings.ReplaceAll(text, `"-y", "--prefix=/", "rea-agents@`, `"-y", "rea-agents@`)
}

// reaRetired returns acceptance fixture Q with every managed entry in the
// v0.50.127 shape, and the target files as the current release writes them.
func reaRetired(t *testing.T) (string, map[reamcp.Platform]string) {
	t.Helper()
	dir := reaQ(t)
	current := map[reamcp.Platform]string{}
	for _, p := range reamcp.Platforms() {
		current[p] = string(readReaFile(t, dir, p))
		retired := retireReaShape(current[p])
		require.NotEqual(t, current[p], retired, p)
		writeReaTarget(t, dir, p, []byte(retired))
		class, err := reamcp.ClassifyFile(p, []byte(retired))
		require.NoError(t, err, p)
		require.Equal(t, reamcp.ClassManaged, class, p)
	}
	return dir, current
}

// An update with the decision carried by the retired entries registers the
// current ones, changes no other byte of the target files or autopus.yaml,
// and leaves nothing for the next update to do.
func TestUpdateRea_MovesTheRetiredEntries(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir, current := reaRetired(t)
	yaml := reaSums(t, dir)["autopus.yaml"]

	out, err := reaUpdate(t, strings.NewReader(""), dir, "--yes")

	require.NoError(t, err, out)
	for _, p := range reamcp.Platforms() {
		assert.Equal(t, current[p], string(readReaFile(t, dir, p)), p)
	}
	assert.Contains(t, out, "REA 역공학 MCP를 등록했습니다: .mcp.json, .codex/config.toml, .agents/plugins/autopus/mcp_config.json (rea-agents@6.5.0)")
	assert.Equal(t, yaml, reaSums(t, dir)["autopus.yaml"])
	assert.Nil(t, reaKey(t, dir))
	settled := reaSums(t, dir)
	out, err = reaUpdate(t, strings.NewReader(""), dir, "--yes")
	require.NoError(t, err, out)
	assert.Equal(t, settled, reaSums(t, dir))
	assert.NotContains(t, out, "REA 역공학 MCP를 등록했습니다")
}

// --no-rea removes the retired entries and stores false; `auto platform
// remove codex` removes the retired Codex table.
func TestUpdateRea_OffPathsRemoveTheRetiredEntries(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir, _ := reaRetired(t)
	out, err := reaUpdate(t, strings.NewReader(""), dir, "--yes", "--no-rea")
	require.NoError(t, err, out)
	assertReaAbsent(t, dir)
	assert.Equal(t, reaBool(false), reaKey(t, dir))

	dir, _ = reaRetired(t)
	out, err = runReaCLI(t, "", "platform", "remove", "codex", "--dir", dir)
	require.NoError(t, err, out)
	assert.NotContains(t, string(readReaFile(t, dir, reamcp.Codex)), "mcp_servers.rea")
	assert.Contains(t, string(readReaFile(t, dir, reamcp.Claude)), "exec npx -y rea-agents@6.5.0 mcp", "other platforms keep their entry")
}
