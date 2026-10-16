package reamcp

import (
	"errors"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJudgeNode_GateContract pins every Node Gate Contract row with the exec
// results of acceptance S6 and the extra oracle cases of plan.md T8.
func TestJudgeNode_GateContract(t *testing.T) {
	t.Parallel()

	exit0 := func(stdout string) NodeExec { return NodeExec{Stdout: stdout} }
	cases := []struct {
		name    string
		run     NodeExec
		status  NodeStatus
		version string
	}{
		{"v22.19.0 is the 22.x floor", exit0("v22.19.0\n"), NodePass, "v22.19.0"},
		{"v22.20.3", exit0("v22.20.3\n"), NodePass, "v22.20.3"},
		{"v24.11.0 is the 24.x floor", exit0("v24.11.0\n"), NodePass, "v24.11.0"},
		{"v24.20.0", exit0("v24.20.0\n"), NodePass, "v24.20.0"},
		{"v26.0.0 opens the open range", exit0("v26.0.0\n"), NodePass, "v26.0.0"},
		{"v27.3.1", exit0("v27.3.1\n"), NodePass, "v27.3.1"},
		{"login banner before the version", exit0("Welcome\nv24.11.0\n"), NodePass, "v24.11.0"},
		{"trailing blank lines and spaces", exit0("  v24.11.0  \r\n\n \n"), NodePass, "v24.11.0"},

		{"v20.11.0", exit0("v20.11.0\n"), NodeUnsupported, "v20.11.0"},
		{"v22.18.0 is below the 22.x floor", exit0("v22.18.0\n"), NodeUnsupported, "v22.18.0"},
		{"v23.1.0 is an odd major", exit0("v23.1.0\n"), NodeUnsupported, "v23.1.0"},
		{"v24.10.0 is below the 24.x floor", exit0("v24.10.0\n"), NodeUnsupported, "v24.10.0"},
		{"v25.9.0 is an odd major", exit0("v25.9.0\n"), NodeUnsupported, "v25.9.0"},

		{"shell could not start", NodeExec{StartErr: syscall.ENOENT}, NodeAbsent, ""},
		{"node could not start on Windows", NodeExec{StartErr: errors.New(`exec: "node": executable file not found in %PATH%`)}, NodeAbsent, ""},
		{"exit 127 with empty stdout", NodeExec{Exit: 127}, NodeAbsent, ""},
		// Blank lines are no output: the stdout holds no non-empty line to read.
		{"exit 127 with only blank lines", NodeExec{Stdout: "\n \n", Exit: 127}, NodeAbsent, ""},

		{"garbage", exit0("garbage\n"), NodeUnknown, ""},
		{"major only", exit0("v26\n"), NodeUnknown, ""},
		{"nightly prerelease tail", exit0("v26.0.0-nightly20260101\n"), NodeUnknown, ""},
		{"rc prerelease tail", exit0("v22.19.0-rc.1\n"), NodeUnknown, ""},
		{"build metadata tail", exit0("v24.11.0+build\n"), NodeUnknown, ""},
		{"no v prefix", exit0("24.11.0\n"), NodeUnknown, ""},
		{"signed component", exit0("v+24.11.0\n"), NodeUnknown, ""},
		{"empty component", exit0("v24..0\n"), NodeUnknown, ""},
		{"four components", exit0("v24.11.0.1\n"), NodeUnknown, ""},
		{"overflowing component", exit0("v99999999999999999999.0.0\n"), NodeUnknown, ""},
		{"version not on the last line", exit0("v24.11.0\nbye\n"), NodeUnknown, ""},
		{"exit 0 with empty stdout", exit0(""), NodeUnknown, ""},
		{"exit 1 with a valid version", NodeExec{Stdout: "v24.11.0\n", Exit: 1}, NodeUnknown, ""},
		{"exit 127 after rc output", NodeExec{Stdout: "bash: rc\n", Exit: 127}, NodeUnknown, ""},
		{"killed by a signal", NodeExec{Stdout: "v24.11.0\n", Exit: -1}, NodeUnknown, ""},
		{"5 second timeout", NodeExec{Exit: -1, TimedOut: true}, NodeUnknown, ""},
		{"timeout wins over printed version", NodeExec{Stdout: "v24.11.0\n", TimedOut: true}, NodeUnknown, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := JudgeNode(tc.run)
			assert.Equal(t, NodeGate{Status: tc.status, Version: tc.version}, got)
		})
	}
}

// TestNodeConstants pins the stored verdict strings doctor reports and the
// range every Node notice names.
func TestNodeConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "^22.19.0 || ^24.11.0 || >=26.0.0", NodeRange)
	assert.Equal(t, []string{"absent", "pass", "unsupported", "unknown"}, []string{
		string(NodeAbsent), string(NodePass), string(NodeUnsupported), string(NodeUnknown),
	})
}
