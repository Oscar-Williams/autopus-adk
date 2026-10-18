package cli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/insajin/autopus-adk/pkg/taskroute"
)

// probeExecutableName is the Claude Code file name the guard runs on this
// host: Windows runs only claude.exe, so PATHEXT never picks a sibling script.
var probeExecutableName = map[bool]string{true: "claude.exe", false: "claude"}[runtime.GOOS == "windows"]

// stubTeamEnv replaces the environment seam with a fixed map and records
// every variable name the snapshot asks for.
func stubTeamEnv(t *testing.T, values map[string]string) *[]string {
	t.Helper()
	var asked []string
	previous := teamCapabilityEnv
	teamCapabilityEnv = func(name string) (string, bool) {
		asked = append(asked, name)
		value, ok := values[name]
		return value, ok
	}
	t.Cleanup(func() { teamCapabilityEnv = previous })
	return &asked
}

// stubVersionProbe replaces the probe execution seam and records each path.
func stubVersionProbe(t *testing.T, output string, err error) *[]string {
	t.Helper()
	var ran []string
	previous := runClaudeVersionProbe
	runClaudeVersionProbe = func(path string) (string, error) {
		ran = append(ran, path)
		return output, err
	}
	t.Cleanup(func() { runClaudeVersionProbe = previous })
	return &ran
}

func TestObserveTeamCapabilityReadsOnlyTheFiveSessionVariables(t *testing.T) {
	asked := stubTeamEnv(t, map[string]string{
		"CLAUDECODE": "1", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1", "CLAUDE_CODE_ENTRYPOINT": "cli",
		"CLAUDE_CODE_SESSION_ATTENDED": "1", "CLAUDE_CODE_EXECPATH": "/opt/claude/bin/claude.exe",
		"CLAUDE_CODE_MESSAGING_TOKEN": "secret-token", "CLAUDE_CODE_CHILD_SESSION": "1",
	})
	snapshot, execPath := observeTeamCapability()
	want := taskroute.Capability{ClaudeCode: "1", AgentTeams: "1", Entrypoint: "cli", SessionAttended: "1", ExecPathBase: "claude.exe"}
	if snapshot != want || execPath != "/opt/claude/bin/claude.exe" {
		t.Fatalf("%+v %q", snapshot, execPath)
	}
	// CLAUDE_CODE_CHILD_SESSION is 1 in the top-level session's Bash tool too,
	// so it cannot mark a subagent and is never read.
	allowed := []string{"CLAUDECODE", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS", "CLAUDE_CODE_ENTRYPOINT", "CLAUDE_CODE_SESSION_ATTENDED", "CLAUDE_CODE_EXECPATH"}
	for _, name := range *asked {
		if !slices.Contains(allowed, name) {
			t.Fatalf("snapshot read %s", name)
		}
	}
	if len(*asked) != len(allowed) {
		t.Fatalf("asked %v", *asked)
	}
}

func TestObserveTeamCapabilityLeavesUnsetValuesEmpty(t *testing.T) {
	stubTeamEnv(t, map[string]string{"CLAUDE_CODE_ENTRYPOINT": "cli", "CLAUDE_CODE_CHILD_SESSION": "1"})
	snapshot, execPath := observeTeamCapability()
	if snapshot != (taskroute.Capability{Entrypoint: "cli"}) || execPath != "" {
		t.Fatalf("%+v %q", snapshot, execPath)
	}
}

func TestProbeClaudeVersionGuardRefusesBeforeExecution(t *testing.T) {
	root := t.TempDir()
	write := func(rel string) string {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return path
	}
	node := write("bin/node")
	claudeDir := filepath.Join(root, "dir", "claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	refused := []string{
		"",
		"claude",
		filepath.Join("probe", "claude"),
		node,
		claudeDir,
		filepath.Join(root, "missing", "claude"),
		write("bin/Claude"),
		write("bin/claude.sh"),
	}
	// A symlink named claude counts as what it resolves to: here a directory.
	link := filepath.Join(root, "links", "claude")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err == nil && os.Symlink(filepath.Join(root, "dir"), link) == nil {
		refused = append(refused, link)
	}
	ran := stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	for _, path := range refused {
		if got := probeClaudeVersion(path); got != "" {
			t.Fatalf("%q probed: %q", path, got)
		}
	}
	if len(*ran) != 0 {
		t.Fatalf("guard let these run: %v", *ran)
	}
}

func TestProbeClaudeVersionRunsTheGuardedPathOnce(t *testing.T) {
	root := t.TempDir()
	claude := filepath.Join(root, probeExecutableName)
	claudeExe := filepath.Join(root, "win", "claude.exe")
	for _, path := range []string{claude, claudeExe} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	paths := []string{claude, claudeExe}
	link := filepath.Join(root, "links", probeExecutableName)
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err == nil && os.Symlink(claude, link) == nil {
		paths = append(paths, link)
	}
	for _, path := range paths {
		ran := stubVersionProbe(t, "2.1.288 (Claude Code)\r\nsecond line\n", nil)
		if got := probeClaudeVersion(path); got != "2.1.288 (Claude Code)" {
			t.Fatalf("%s: %q", path, got)
		}
		if !slices.Equal(*ran, []string{path}) {
			t.Fatalf("%s: ran %v", path, *ran)
		}
	}
	stubVersionProbe(t, "2.1.288 (Claude Code)\n", errors.New("exit status 1"))
	if got := probeClaudeVersion(claude); got != "" {
		t.Fatalf("failed probe reported %q", got)
	}
}
