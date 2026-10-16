package cli

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// testReaNodeV24 is the Node probe seam TestMain pins for the whole CLI test
// binary (spec.md Existing Test Changes rule 6): node v24.11.0, exit 0. No
// CLI test runs a real node.
func testReaNodeV24(context.Context, string, ...string) (string, int, error) {
	return "v24.11.0\n", 0, nil
}

// reaNodeExecRecorder is a fake exec boundary that records each probe command.
type reaNodeExecRecorder struct {
	mu       sync.Mutex
	commands [][]string
	stdout   string
	exit     int
	startErr error
}

func (r *reaNodeExecRecorder) exec(_ context.Context, name string, args ...string) (string, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, append([]string{name}, args...))
	return r.stdout, r.exit, r.startErr
}

// The probe runs the REA server's login shell on POSIX and node from PATH on
// Windows (Node Gate Contract).
func TestReaNodeCommand_ByOS(t *testing.T) {
	t.Parallel()
	for _, goos := range []string{"darwin", "linux"} {
		name, args := reaNodeCommand(goos)
		assert.Equal(t, "/bin/bash", name, goos)
		assert.Equal(t, []string{"-l", "-c", "node --version"}, args, goos)
	}
	name, args := reaNodeCommand("windows")
	assert.Equal(t, "node", name)
	assert.Equal(t, []string{"--version"}, args)
}

// One run asks its probe as often as it likes; the command runs once and
// every answer is the first verdict (REQ-REA-05).
func TestReaNodeProbe_RunsOncePerRun(t *testing.T) {
	t.Parallel()
	for _, goos := range []string{"darwin", "windows"} {
		recorder := &reaNodeExecRecorder{stdout: "v24.11.0\n"}
		probe := &reaNodeProbe{exec: recorder.exec, goos: goos}

		first := probe.Gate(context.Background())
		recorder.stdout = "v20.11.0\n"
		second := probe.Gate(context.Background())

		want := reamcp.NodeGate{Status: reamcp.NodePass, Version: "v24.11.0"}
		assert.Equal(t, want, first, goos)
		assert.Equal(t, want, second, goos)
		name, args := reaNodeCommand(goos)
		assert.Equal(t, [][]string{append([]string{name}, args...)}, recorder.commands, goos)
	}
}

// The probe judges only what the exec boundary returned (acceptance S6).
func TestReaNodeProbe_JudgesExecBoundary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		recorder *reaNodeExecRecorder
		want     reamcp.NodeGate
	}{
		{"N24", &reaNodeExecRecorder{stdout: "v24.11.0\n"}, reamcp.NodeGate{Status: reamcp.NodePass, Version: "v24.11.0"}},
		{"N2218", &reaNodeExecRecorder{stdout: "v22.18.0\n"}, reamcp.NodeGate{Status: reamcp.NodeUnsupported, Version: "v22.18.0"}},
		{"ENOENT", &reaNodeExecRecorder{startErr: syscall.ENOENT}, reamcp.NodeGate{Status: reamcp.NodeAbsent}},
		{"exit 127 empty", &reaNodeExecRecorder{exit: 127}, reamcp.NodeGate{Status: reamcp.NodeAbsent}},
		{"garbage", &reaNodeExecRecorder{stdout: "garbage\n"}, reamcp.NodeGate{Status: reamcp.NodeUnknown}},
	}
	for _, tc := range cases {
		got := probeReaNode(context.Background(), "linux", reaNodeTimeout, tc.recorder.exec)
		assert.Equal(t, tc.want, got, tc.name)
		assert.Len(t, tc.recorder.commands, 1, tc.name)
	}
}

// A probe that outlives its deadline is unknown even when it printed a
// supported version, and the deadline is 5 seconds.
func TestReaNodeProbe_TimeoutIsUnknown(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 5*time.Second, reaNodeTimeout)

	blocked := func(ctx context.Context, _ string, _ ...string) (string, int, error) {
		<-ctx.Done()
		return "v24.11.0\n", -1, nil
	}
	got := probeReaNode(context.Background(), "linux", 20*time.Millisecond, blocked)
	assert.Equal(t, reamcp.NodeGate{Status: reamcp.NodeUnknown}, got)
}

// TestMain pins the seam, so a run's probe in any CLI test sees N24.
func TestReaNodeExec_TestMainPinsV24(t *testing.T) {
	got := newReaNodeProbe().Gate(context.Background())
	assert.Equal(t, reamcp.NodeGate{Status: reamcp.NodePass, Version: "v24.11.0"}, got)
}

// The real exec boundary reports stdout, the exit status, and a start error.
// It runs the test binary's --version fixture and /bin/sh, never node.
func TestReaNodeExec_RealBoundary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh")
	}
	t.Parallel()
	ctx := context.Background()

	stdout, exit, err := execReaNode(ctx, os.Args[0], "--version")
	require.NoError(t, err)
	assert.Equal(t, "omp/17.2.7\n", stdout)
	assert.Equal(t, 0, exit)

	stdout, exit, err = execReaNode(ctx, "/bin/sh", "-c", "printf 'bash: rc\\n'; echo noise >&2; exit 127")
	require.NoError(t, err)
	assert.Equal(t, "bash: rc\n", stdout, "stderr is not part of stdout")
	assert.Equal(t, 127, exit)

	_, _, err = execReaNode(ctx, filepath.Join(t.TempDir(), "missing-shell"))
	assert.Error(t, err, "a command that cannot start reports a start error")
}

// A killed probe returns soon after its deadline even when a child process
// keeps stdout open, so the run stays within the probe bound.
func TestReaNodeExec_DeadlineBoundsChildHoldingStdout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses /bin/sh")
	}
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, exit, err := execReaNode(ctx, "/bin/sh", "-c", "sleep 10 & sleep 10")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.NotEqual(t, 0, exit)
	assert.Less(t, elapsed, 5*time.Second)
}
