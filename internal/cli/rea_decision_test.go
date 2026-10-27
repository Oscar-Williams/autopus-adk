package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaCodexBase is a Codex config with one user server and no rea table.
const reaCodexBase = "model = \"gpt-6\"\n\n[mcp_servers.mine]\ncommand = \"mine-mcp\"  # keep\n"

// reaC3 is acceptance S4 row C3: a user-owned .mcp.json rea entry.
const reaC3 = "{\n  \"mcpServers\": {\n    \"rea\": {\n      \"args\": [\"-y\", \"rea-agents@6.5.0\", \"mcp\"],\n      \"command\": \"npx\"\n    }\n  }\n}\n"

// reaFixture returns a temp project whose autopus.yaml selects platforms and
// holds no REA decision.
func reaFixture(t *testing.T, platforms ...string) (string, *config.HarnessConfig) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.DefaultFullConfig("p")
	cfg.Platforms = platforms
	require.NoError(t, config.Save(dir, cfg))
	return dir, cfg
}

// writeReaTarget writes p's target file under dir.
func writeReaTarget(t *testing.T, dir string, p reamcp.Platform, data []byte) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(reamcp.TargetRelPath(p)))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

// readReaFile returns p's target file under dir, nil when missing.
func readReaFile(t *testing.T, dir string, p reamcp.Platform) []byte {
	t.Helper()
	data, err := readReaTarget(dir, p)
	require.NoError(t, err)
	return data
}

// reaApplied applies action to data the way the adapters do.
func reaApplied(t *testing.T, p reamcp.Platform, data []byte, action config.ReaMCPAction) []byte {
	t.Helper()
	out, _, err := reamcp.ApplyFile(p, data, action)
	require.NoError(t, err)
	return out
}

// writeManagedRea gives every listed platform's target file the managed entry.
func writeManagedRea(t *testing.T, dir string, platforms ...reamcp.Platform) {
	t.Helper()
	for _, p := range platforms {
		base := []byte(nil)
		if p == reamcp.Codex {
			base = []byte(reaCodexBase)
		}
		writeReaTarget(t, dir, p, reaApplied(t, p, base, config.ReaMCPAdd))
	}
}

// Acceptance S2: each fixture resolves to the Decision State Contract row.
func TestResolveReaDecision_DecisionState(t *testing.T) {
	t.Parallel()
	all := []string{"claude-code", "codex", "antigravity-cli"}
	cases := []struct {
		name  string
		setup func(t *testing.T) (string, *config.HarnessConfig)
		want  reaDecision
	}{
		{"(1) OLD", func(t *testing.T) (string, *config.HarnessConfig) {
			return reaFixture(t, all...)
		}, reaDecision{reaUndecided, reaSourceNone}},
		{"(2) Q", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, all...)
			writeManagedRea(t, dir, reamcp.Platforms()...)
			return dir, cfg
		}, reaDecision{reaOn, reaSourceManaged}},
		{"(3) user-owned C3 only", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, all...)
			writeReaTarget(t, dir, reamcp.Claude, []byte(reaC3))
			return dir, cfg
		}, reaDecision{reaUndecided, reaSourceNone}},
		{"(4) Q with enabled false", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, all...)
			writeManagedRea(t, dir, reamcp.Platforms()...)
			cfg.Features.Rea.Enabled = reaBool(false)
			return dir, cfg
		}, reaDecision{reaOff, reaSourceKey}},
		{"(5) OLD with enabled true", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, all...)
			cfg.Features.Rea.Enabled = reaBool(true)
			return dir, cfg
		}, reaDecision{reaOn, reaSourceKey}},
		{"(6) managed only in an unselected file", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, "claude-code")
			writeManagedRea(t, dir, reamcp.Codex)
			return dir, cfg
		}, reaDecision{reaUndecided, reaSourceNone}},
		{"other pin only", func(t *testing.T) (string, *config.HarnessConfig) {
			dir, cfg := reaFixture(t, all...)
			writeReaTarget(t, dir, reamcp.Codex, []byte("[mcp_servers.rea]\ncommand = \"npx\"\nargs = [\"-y\", \"--prefix=/\", \"rea-agents@6.4.0\", \"mcp\"]\nstartup_timeout_sec = 120\n"))
			return dir, cfg
		}, reaDecision{reaUndecided, reaSourceNone}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir, cfg := tc.setup(t)
			assert.Equal(t, tc.want, resolveReaDecision(dir, cfg))
		})
	}
}

// The Decision State Contract write table, row by row.
func TestReaKeyAfterRun_WriteTable(t *testing.T) {
	t.Parallel()
	yes, no := reaBool(true), reaBool(false)
	cases := []struct {
		name    string
		state   reaState
		stored  *bool
		managed bool
		want    *bool
	}{
		{"on, M, no key: the entry carries it", reaOn, nil, true, nil},
		{"on, M, false is corrected", reaOn, no, true, yes},
		{"on, M, true stays", reaOn, yes, true, yes},
		{"on, no M, no key: true is written", reaOn, nil, false, yes},
		{"on, no M, false is corrected", reaOn, no, false, yes},
		{"off writes false", reaOff, nil, true, no},
		{"off over true writes false", reaOff, yes, false, no},
		{"off over false stays", reaOff, no, false, no},
		{"undecided writes nothing", reaUndecided, nil, true, nil},
		{"undecided keeps a stored value", reaUndecided, yes, false, yes},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, reaKeyAfterRun(tc.state, tc.stored, tc.managed), tc.name)
	}
	assert.True(t, reaSameKey(nil, nil))
	assert.True(t, reaSameKey(reaBool(true), yes))
	assert.False(t, reaSameKey(nil, no))
	assert.False(t, reaSameKey(yes, no))
}

// Acceptance S2 (6): the off sweep removes the managed table from the
// unselected Codex file and keeps every other line byte for byte, and it
// leaves selected and user-owned files alone.
func TestSweepReaUnselected_RemovesManagedOutsideSelection(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	writeManagedRea(t, dir, reamcp.Claude, reamcp.Codex)
	writeReaTarget(t, dir, reamcp.Antigravity, []byte(reaC3))
	claudeBefore := readReaFile(t, dir, reamcp.Claude)

	require.NoError(t, sweepReaUnselected(dir, cfg.Platforms))

	assert.Equal(t, reaCodexBase, string(readReaFile(t, dir, reamcp.Codex)))
	assert.Equal(t, claudeBefore, readReaFile(t, dir, reamcp.Claude), "a selected file is the adapter's to change")
	assert.Equal(t, reaC3, string(readReaFile(t, dir, reamcp.Antigravity)))
	require.NoError(t, sweepReaUnselected(dir, cfg.Platforms), "a second sweep finds nothing to remove")
}

// Acceptance S16 (d) for the sweep: a target directory that is a symlink out
// of the project refuses the write with the path, and the outside file keeps
// its bytes.
func TestSweepReaUnselected_RefusesSymlinkedTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture")
	}
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	outside := t.TempDir()
	managed := reaApplied(t, reamcp.Codex, []byte(reaCodexBase), config.ReaMCPAdd)
	require.NoError(t, os.WriteFile(filepath.Join(outside, "config.toml"), managed, 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, ".codex")))

	err := sweepReaUnselected(dir, cfg.Platforms)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ".codex/config.toml")
	data, readErr := os.ReadFile(filepath.Join(outside, "config.toml"))
	require.NoError(t, readErr)
	assert.Equal(t, managed, data)
}

// A preview reads M from the planned mappings, and a tracked file the
// operator deleted is skipped instead of planned.
func TestReaPlannedManaged(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	assert.True(t, reaPlannedManaged(dir, cfg.Platforms, config.ReaMCPAdd), "add plans an entry into a new file")
	assert.False(t, reaPlannedManaged(dir, cfg.Platforms, config.ReaMCPKeep))

	writeReaTarget(t, dir, reamcp.Claude, []byte(reaC3))
	assert.False(t, reaPlannedManaged(dir, cfg.Platforms, config.ReaMCPAdd), "a user-owned entry is never replaced")

	writeManagedRea(t, dir, reamcp.Claude)
	assert.True(t, reaPlannedManaged(dir, cfg.Platforms, config.ReaMCPKeep))
	assert.False(t, reaPlannedManaged(dir, cfg.Platforms, config.ReaMCPRemove))

	skipped, skippedCfg := reaFixture(t, "claude-code")
	manifest := filepath.Join(skipped, ".autopus", "claude-code-manifest.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(manifest), 0o755))
	require.NoError(t, os.WriteFile(manifest, []byte(`{"platform":"claude-code","files":{".mcp.json":{"checksum":"x"}}}`), 0o644))
	assert.True(t, reaTargetSkipped(skipped, reamcp.Claude))
	assert.False(t, reaPlannedManaged(skipped, skippedCfg.Platforms, config.ReaMCPAdd))
}

// An unreadable or undecodable target file is user_owned, never managed.
func TestReaFileClass_UndecodableIsUserOwned(t *testing.T) {
	t.Parallel()
	dir, _ := reaFixture(t, "claude-code")
	assert.Equal(t, reamcp.ClassAbsent, reaFileClass(dir, reamcp.Claude))
	writeReaTarget(t, dir, reamcp.Claude, []byte("{not json"))
	assert.Equal(t, reamcp.ClassUserOwned, reaFileClass(dir, reamcp.Claude))
	assert.Equal(t, []string{"omp", "opencode"}, reaGuidancePlatforms([]string{"opencode", "codex", "omp"}))
	assert.Equal(t, []string{}, reaGuidancePlatforms([]string{"codex"}))
}
