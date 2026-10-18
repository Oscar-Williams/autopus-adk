package cli

// SPEC-REA-001 T12 (acceptance S2 (4), (6)): an off decision through the
// real `auto update`, from the stored key and from --no-rea with the off
// sweep over an unselected target file.

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaQ returns acceptance fixture Q: OLD updated with --yes --rea, so the
// three target files hold managed entries and autopus.yaml holds no key.
func reaQ(t *testing.T) string {
	t.Helper()
	dir := reaOLD(t, reaAllTargets)
	out, err := reaUpdate(t, strings.NewReader(""), dir, "--yes", "--rea")
	require.NoError(t, err, out)
	assertReaManagedOnce(t, dir)
	return dir
}

// S2 (4): Q with enabled false removes every managed entry and leaves
// autopus.yaml byte-identical.
func TestUpdateRea_StoredOffRemovesEntries(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir := reaQ(t)
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	cfg.Features.Rea.Enabled = reaBool(false)
	require.NoError(t, config.Save(dir, cfg))
	before := reaSums(t, dir)["autopus.yaml"]

	out, err := reaUpdate(t, strings.NewReader(""), dir, "--yes")

	require.NoError(t, err, out)
	assertReaAbsent(t, dir)
	assert.Equal(t, before, reaSums(t, dir)["autopus.yaml"])
	assert.Equal(t, [3]int{0, 0, 0}, reaCounts(out), out)
}

// S2 (6): with managed only in the unselected .codex/config.toml the project
// is undecided; --no-rea sweeps that table, keeps every other line byte for
// byte, and stores false.
func TestUpdateRea_NoReaSweepsUnselectedTarget(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir := reaQ(t)
	for _, p := range []reamcp.Platform{reamcp.Claude, reamcp.Antigravity} {
		writeReaTarget(t, dir, p, reaApplied(t, p, readReaFile(t, dir, p), config.ReaMCPRemove))
	}
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	cfg.Platforms = []string{"claude-code"}
	require.NoError(t, config.Save(dir, cfg))
	require.Equal(t, reaDecision{reaUndecided, reaSourceNone}, resolveReaDecision(dir, cfg))
	codex := readReaFile(t, dir, reamcp.Codex)
	want := reaApplied(t, reamcp.Codex, codex, config.ReaMCPRemove)
	require.NotEqual(t, codex, want)

	out, err := reaUpdate(t, strings.NewReader(""), dir, "--yes", "--no-rea")

	require.NoError(t, err, out)
	assert.Equal(t, string(want), string(readReaFile(t, dir, reamcp.Codex)))
	assert.NotContains(t, string(readReaFile(t, dir, reamcp.Codex)), "mcp_servers.rea")
	assert.Equal(t, reaBool(false), reaKey(t, dir))
	assert.Equal(t, [3]int{0, 0, 1}, reaCounts(out), out)
}

// The off sweep shows in a preview: `--plan --no-rea` lists each unselected
// target file whose managed entry the run would remove and writes nothing; a
// preview that does not turn REA off lists none.
func TestUpdateRea_PlanShowsTheOffSweep(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir := reaQ(t)
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	cfg.Platforms = []string{"claude-code"}
	require.NoError(t, config.Save(dir, cfg))
	before := reaSums(t, dir)

	out, err := reaUpdate(t, strings.NewReader(""), dir, "--plan", "--no-rea")

	require.NoError(t, err, out)
	assert.Equal(t, before, reaSums(t, dir))
	const sweep = " — REA off would remove the managed rea entry of this unselected platform"
	assert.Contains(t, out, "update .codex/config.toml (codex)"+sweep)
	// The plugin MCP file held the managed entry alone, so the sweep deletes it.
	assert.Contains(t, out, "remove .agents/plugins/autopus/mcp_config.json (antigravity-cli)"+sweep)
	assert.NotContains(t, out, "(claude-code)"+sweep, "the selected file is the adapter's to plan")

	plain, err := reaUpdate(t, strings.NewReader(""), dir, "--plan")
	require.NoError(t, err, plain)
	assert.NotContains(t, plain, sweep)
}

// reaSettledOff stores the off decision in a settled composition fixture
// through the real non-interactive `--no-rea` update, so a later interactive
// update asks no REA question and prints no REA line: its output stays the
// composition step's own (SPEC-MODELUX-001 line goldens). The codex calls of
// that run are forgotten, as settled forgets its own.
func (r composeUpdateRun) reaSettledOff(t *testing.T, dir string) {
	t.Helper()
	interactive := composeTTYProbe
	composeTTYProbe = func() bool { return false }
	out, err := runComposeUpdate(t, strings.NewReader(""), "--dir", dir, "--no-rea")
	composeTTYProbe = interactive
	require.NoError(t, err, out)
	require.Equal(t, reaBool(false), reaKey(t, dir))
	require.NoError(t, os.RemoveAll(r.codexLog))
}
