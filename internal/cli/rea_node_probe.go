package cli

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaNodeTimeout bounds one Node probe run (Node Gate Contract).
const reaNodeTimeout = 5 * time.Second

// reaNodeWaitDelay bounds how long a killed probe may wait for its output
// pipes, so a child process that keeps stdout open cannot hold the run past
// the deadline.
const reaNodeWaitDelay = 200 * time.Millisecond

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

// execReaNode is the real exec boundary. The probe reads no stdin and drops
// stderr; the exit status is -1 when the process was killed.
func execReaNode(ctx context.Context, name string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.WaitDelay = reaNodeWaitDelay
	if err := cmd.Start(); err != nil {
		return "", 0, err
	}
	// The exit status carries the verdict, so Wait's own error adds nothing;
	// a WaitDelay overrun keeps the stdout captured before the pipes closed.
	_ = cmd.Wait()
	exit := -1
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	return stdout.String(), exit, nil
}
