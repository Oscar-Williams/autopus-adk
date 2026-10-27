//go:build !windows

package cli

// SPEC-REA-001 security audit L6: the Node probe captures bounded stdout and
// kills its whole process group, so neither a chatty login shell nor a
// grandchild that holds the stdout pipe can outlive or flood the probe.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A grandchild that holds stdout dies with the probe's process group when the
// deadline kills the probe.
func TestReaNodeExec_KillsTheGrandchildHoldingStdout(t *testing.T) {
	t.Parallel()
	pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, exit, err := execReaNode(ctx, "/bin/sh", "-c", "sleep 30 & echo $! > '"+pidFile+"'; sleep 30")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.NotEqual(t, 0, exit)
	assert.Less(t, elapsed, 5*time.Second)
	data, err := os.ReadFile(pidFile)
	require.NoError(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	require.NoError(t, err)
	assert.Eventually(t, func() bool {
		return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH)
	}, 5*time.Second, 50*time.Millisecond, "the grandchild holding stdout must not survive the probe")
}

// A login shell that prints without end is stopped at the output cap and
// judged unknown, with no more than the cap captured.
func TestReaNodeExec_BoundsStdout(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	stdout, exit, err := execReaNode(ctx, "/bin/sh", "-c", "yes v24.11.0")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(stdout), reaNodeOutputLimit)
	assert.Equal(t, -1, exit, "a probe stopped at the cap is never a clean exit")
	assert.Less(t, elapsed, 5*time.Second)
}
