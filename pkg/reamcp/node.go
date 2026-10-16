package reamcp

import (
	"strconv"
	"strings"
)

// NodeRange is the Node range rea-agents supports; every Node notice names it
// (Node Gate Contract).
const NodeRange = "^22.19.0 || ^24.11.0 || >=26.0.0"

// NodeStatus is the Node gate verdict of one probe run. The values are the
// strings doctor reports.
type NodeStatus string

const (
	// NodeAbsent: the probe command could not start, or it exited 127 with
	// no output on stdout.
	NodeAbsent NodeStatus = "absent"
	// NodePass: exit 0 and a version inside NodeRange.
	NodePass NodeStatus = "pass"
	// NodeUnsupported: exit 0 and a well-formed version outside NodeRange.
	NodeUnsupported NodeStatus = "unsupported"
	// NodeUnknown: every other result, including a prerelease tail, any other
	// nonzero exit even with output, and a timeout.
	NodeUnknown NodeStatus = "unknown"
)

// NodeExec is the raw result of one `node --version` probe run. Its output
// is untrusted: the gate reads only the exit status and the last non-empty
// stdout line.
type NodeExec struct {
	Stdout   string
	Exit     int   // exit status; -1 when the process was killed
	StartErr error // set when the shell (node on Windows) could not start
	TimedOut bool  // the run hit the probe deadline
}

// NodeGate is a judged probe run. Version is the matched `v<MAJOR>.<MINOR>.<PATCH>`
// line for pass and unsupported and empty otherwise, so a notice never
// echoes anything else the probe printed.
type NodeGate struct {
	Status  NodeStatus
	Version string
}

// JudgeNode applies the Node Gate Contract to one probe run.
func JudgeNode(run NodeExec) NodeGate {
	line := nodeLastLine(run.Stdout)
	switch {
	case run.StartErr != nil:
		return NodeGate{Status: NodeAbsent}
	case run.TimedOut:
		return NodeGate{Status: NodeUnknown}
	case run.Exit == 127 && line == "":
		return NodeGate{Status: NodeAbsent}
	case run.Exit != 0:
		return NodeGate{Status: NodeUnknown}
	}
	major, minor, ok := nodeVersion(line)
	switch {
	case !ok:
		return NodeGate{Status: NodeUnknown}
	case nodeSupported(major, minor):
		return NodeGate{Status: NodePass, Version: line}
	}
	return NodeGate{Status: NodeUnsupported, Version: line}
}

// nodeLastLine returns the last stdout line that is not empty after trimming
// surrounding whitespace, trimmed, or "" when there is none.
func nodeLastLine(stdout string) string {
	lines := strings.Split(stdout, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

// nodeVersion parses exactly `v<MAJOR>.<MINOR>.<PATCH>` with ASCII digits
// only. A prerelease or build tail, a sign, or a missing part fails.
func nodeVersion(line string) (major, minor int, ok bool) {
	rest, found := strings.CutPrefix(line, "v")
	parts := strings.Split(rest, ".")
	if !found || len(parts) != 3 {
		return 0, 0, false
	}
	var nums [3]int
	for i, part := range parts {
		if part == "" || strings.Trim(part, "0123456789") != "" {
			return 0, 0, false
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, false
		}
		nums[i] = n
	}
	return nums[0], nums[1], true
}

// nodeSupported reports whether a version is inside NodeRange. Every range
// starts at patch 0, so major and minor decide it.
func nodeSupported(major, minor int) bool {
	switch {
	case major == 22:
		return minor >= 19
	case major == 24:
		return minor >= 11
	}
	return major >= 26
}
