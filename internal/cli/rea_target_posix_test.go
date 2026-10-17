//go:build !windows

package cli

// SPEC-REA-001 security audit L5 on POSIX: FIFO, symlink, and oversized
// target fixtures.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaUnsafeTarget plants an unsafe target for p under dir: "fifo", "zero" (a
// symlink to /dev/zero), "dirlink" (the target's directory is a symlink), or
// "large" (a regular file one byte over the cap).
func reaUnsafeTarget(t *testing.T, dir string, p reamcp.Platform, kind string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(reamcp.TargetRelPath(p)))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	switch kind {
	case "fifo":
		require.NoError(t, syscall.Mkfifo(path, 0o644))
	case "zero":
		require.NoError(t, os.Symlink("/dev/zero", path))
	case "dirlink":
		outside := t.TempDir()
		require.NoError(t, os.Remove(filepath.Dir(path)))
		require.NoError(t, os.Symlink(outside, filepath.Dir(path)))
	case "large":
		require.NoError(t, os.WriteFile(path, bytes.Repeat([]byte(" "), reaTargetLimit+1), 0o644))
	}
}

// withinReaDeadline fails the test instead of hanging when run blocks.
func withinReaDeadline(t *testing.T, run func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		run()
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("a REA read blocked on an unsafe target")
	}
}

func TestReadReaTarget_RefusesUnsafeFiles(t *testing.T) {
	t.Parallel()
	cases := map[string]reaRefusal{
		"fifo": reaRefusedIrregular, "zero": reaRefusedSymlink, "dirlink": reaRefusedSymlink, "large": reaRefusedTooLarge,
	}
	for kind, reason := range cases {
		dir := t.TempDir()
		reaUnsafeTarget(t, dir, reamcp.Antigravity, kind)
		withinReaDeadline(t, func() {
			data, err := readReaTarget(dir, reamcp.Antigravity)
			var refused *reaTargetError
			require.ErrorAs(t, err, &refused, kind)
			assert.Equal(t, &reaTargetError{Rel: ".agents/plugins/autopus/mcp_config.json", Reason: reason}, refused, kind)
			assert.Nil(t, data, kind)
			assert.Equal(t, reamcp.ClassUserOwned, reaFileClass(dir, reamcp.Antigravity), kind)
		})
	}
}

// A run reads only the target files its platform writes may change: an
// unselected FIFO target is never opened, so the run cannot hang on it.
func TestReaRun_SnapshotLeavesUnselectedTargetsUnread(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	reaUnsafeTarget(t, dir, reamcp.Antigravity, "fifo")
	run, _ := testReaRun(t, dir, cfg, "v24.11.0\n")
	var out bytes.Buffer

	withinReaDeadline(t, func() {
		run.decide(context.Background(), &out, cfg, reaOn)
		run.snapshot(cfg.Platforms)
		writeManagedRea(t, dir, reamcp.Claude)
		require.NoError(t, run.finish(&out, dir, cfg))
	})

	assert.NotContains(t, out.String(), ".agents/plugins/autopus/mcp_config.json")
	assert.Contains(t, out.String(), "REA 역공학 MCP를 등록했습니다: .mcp.json (rea-agents@6.5.0)")
}

// The off sweep prints one skip line for each unselected target it cannot
// inspect (FIFO, symlink to /dev/zero, oversized), writes nothing, does not
// fail the run, and still stores the off decision.
func TestReaRunFinish_SweepSkipsUninspectableTargets(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	reaUnsafeTarget(t, dir, reamcp.Antigravity, "fifo")
	reaUnsafeTarget(t, dir, reamcp.Codex, "zero")
	run, _ := testReaRun(t, dir, cfg, "v24.11.0\n")
	var out bytes.Buffer

	withinReaDeadline(t, func() {
		run.decide(context.Background(), &out, cfg, reaOff)
		require.NoError(t, run.finish(&out, dir, cfg))
	})

	assert.Equal(t, 1, strings.Count(out.String(), "  REA MCP 대상 파일을 건너뜁니다: .codex/config.toml (symlink를 따라가지 않습니다)\n"), out.String())
	assert.Equal(t, 1, strings.Count(out.String(), "  REA MCP 대상 파일을 건너뜁니다: .agents/plugins/autopus/mcp_config.json (일반 파일이 아닙니다)\n"), out.String())
	reloaded, err := config.Load(dir)
	require.NoError(t, err)
	assert.Equal(t, reaBool(false), reloaded.Features.Rea.Enabled)

	large, largeCfg := reaFixture(t, "claude-code")
	reaUnsafeTarget(t, large, reamcp.Codex, "large")
	run, _ = testReaRun(t, large, largeCfg, "v24.11.0\n")
	out.Reset()
	run.decide(context.Background(), &out, largeCfg, reaOff)
	require.NoError(t, run.finish(&out, large, largeCfg))
	assert.Contains(t, out.String(), "  REA MCP 대상 파일을 건너뜁니다: .codex/config.toml (1 MiB보다 큽니다)\n")
	info, err := os.Stat(filepath.Join(large, ".codex", "config.toml"))
	require.NoError(t, err)
	assert.EqualValues(t, reaTargetLimit+1, info.Size(), "nothing is written")
}

// The reproduced hang: a committed `.mcp.json` symlink to a FIFO in a project
// that does not select claude-code no longer stops a real `auto update`; the
// --no-rea sweep reports the file it left unread. (`.gemini/settings.json` and
// `.codex/config.toml` are also read by update paths outside REA: the Claude
// adapter's stale completion hook scan and the Codex supervisor migration.)
func TestUpdateRea_UnselectedFIFOTargetDoesNotHang(t *testing.T) {
	reaCLIEnv(t, "v24.11.0\n")
	stubDriverProbes(t, false, "")
	dir := reaOLD(t, "codex")
	fifo := filepath.Join(t.TempDir(), "fifo")
	require.NoError(t, syscall.Mkfifo(fifo, 0o644))
	require.NoError(t, os.Symlink(fifo, filepath.Join(dir, ".mcp.json")))

	for _, args := range [][]string{{"--yes", "--rea"}, {"--yes", "--no-rea"}} {
		var out string
		var err error
		withinReaDeadline(t, func() { out, err = reaUpdate(t, strings.NewReader(""), dir, args...) })
		require.NoError(t, err, out)
		skipped := strings.Count(out, "REA MCP 대상 파일을 건너뜁니다: .mcp.json (symlink를 따라가지 않습니다)")
		if args[1] == "--no-rea" {
			assert.Equal(t, 1, skipped, out)
		} else {
			assert.Zero(t, skipped, "an unselected target stays unread: "+out)
		}
	}
}
