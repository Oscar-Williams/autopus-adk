package cli

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/insajin/autopus-adk/pkg/processprobe"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaNodeTimeout bounds one Node probe run (Node Gate Contract).
const reaNodeTimeout = 5 * time.Second

// reaNodeWaitDelay bounds how long a killed probe may wait for its output
// pipes, so a child process that keeps stdout open cannot hold the run past
// the deadline.
const reaNodeWaitDelay = 200 * time.Millisecond

// reaNodeOutputLimit bounds what one probe captures: a login shell that
// prints more is stopped and judged unknown.
const reaNodeOutputLimit = 64 << 10

// reaNodeExecFunc runs name with args and reports stdout, the exit status,
// and an error when the command could not start.
type reaNodeExecFunc func(ctx context.Context, name string, args ...string) (stdout string, exit int, startErr error)

// reaNodeExec is the Node probe's exec boundary seam (REQ-REA-20). Tests
// replace it and restore it with t.Cleanup; TestMain pins it to v24.11.0.
var reaNodeExec reaNodeExecFunc = execReaNode

// reaNodeCommand returns the probe command for goos: the login shell the REA
// server starts under on POSIX, and node from PATH on Windows.
func reaNodeCommand(goos string) (string, []string) {
	if goos == "windows" {
		return "node", []string{"--version"}
	}
	return "/bin/bash", []string{"-l", "-c", "node --version"}
}

// reaNodeProbe runs the Node probe at most once per run. A run creates one
// probe and asks it only when the REA decision is on or the REA question was
// answered yes; doctor asks it once on every run.
type reaNodeProbe struct {
	exec reaNodeExecFunc
	goos string
	once sync.Once
	gate reamcp.NodeGate
}

// newReaNodeProbe returns a run's probe over the current exec seam.
func newReaNodeProbe() *reaNodeProbe {
	return &reaNodeProbe{exec: reaNodeExec, goos: runtime.GOOS}
}

// Gate returns the run's Node gate. Only the first call runs the probe.
func (p *reaNodeProbe) Gate(ctx context.Context) reamcp.NodeGate {
	p.once.Do(func() {
		p.gate = probeReaNode(ctx, p.goos, reaNodeTimeout, p.exec)
	})
	return p.gate
}

// probeReaNode runs one probe under timeout and judges its result.
func probeReaNode(ctx context.Context, goos string, timeout time.Duration, run reaNodeExecFunc) reamcp.NodeGate {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	name, args := reaNodeCommand(goos)
	stdout, exit, startErr := run(ctx, name, args...)
	return reamcp.JudgeNode(reamcp.NodeExec{
		Stdout:   stdout,
		Exit:     exit,
		StartErr: startErr,
		TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded),
	})
}

// execReaNode is the real exec boundary. The probe reads no stdin, runs in its
// own process group, and captures at most reaNodeOutputLimit bytes of stdout
// (stderr is bounded and dropped). A deadline, an output overrun, or any
// failed run kills the whole group, so a grandchild that holds the stdout
// pipe dies with it. The exit status is -1 when the process was killed or
// stopped at the cap.
func execReaNode(ctx context.Context, name string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = reaNodeWaitDelay
	stdout, err := processprobe.OutputLimited(cmd, reaNodeOutputLimit)
	if cmd.Process == nil {
		return "", 0, err
	}
	// The exit status carries the verdict; only an overrun overrides it.
	exit := -1
	if cmd.ProcessState != nil && !errors.Is(err, processprobe.ErrOutputLimit) {
		exit = cmd.ProcessState.ExitCode()
	}
	return string(stdout), exit, nil
}
