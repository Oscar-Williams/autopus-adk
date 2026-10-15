package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/insajin/autopus-adk/pkg/processprobe"
	"github.com/insajin/autopus-adk/pkg/taskroute"
)

const (
	claudeVersionProbeTimeout  = 5 * time.Second
	claudeVersionProbeMaxBytes = 4096
)

// teamCapabilityEnv reads one variable of the process environment, which the
// Bash tool inherits from the Claude Code session. Tests replace it.
var teamCapabilityEnv = os.LookupEnv

// runClaudeVersionProbe executes `<path> --version` for a path the guard in
// probeClaudeVersion accepted and returns its stdout. Tests replace it.
var runClaudeVersionProbe = execClaudeVersionProbe

// observeTeamCapability builds the capability snapshot from the six session
// variables the Capability Observation Contract names and returns the raw
// CLAUDE_CODE_EXECPATH for the version probe. No other variable is read, so a
// secret such as CLAUDE_CODE_MESSAGING_TOKEN never enters the snapshot.
func observeTeamCapability() (taskroute.Capability, string) {
	read := func(name string) string {
		value, _ := teamCapabilityEnv(name)
		return value
	}
	snapshot := taskroute.Capability{
		ClaudeCode:      read("CLAUDECODE"),
		AgentTeams:      read("CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS"),
		Entrypoint:      read("CLAUDE_CODE_ENTRYPOINT"),
		SessionAttended: read("CLAUDE_CODE_SESSION_ATTENDED"),
		ChildSession:    read("CLAUDE_CODE_CHILD_SESSION"),
	}
	execPath := read("CLAUDE_CODE_EXECPATH")
	if execPath != "" {
		snapshot.ExecPathBase = filepath.Base(execPath)
	}
	return snapshot, execPath
}

// probeClaudeVersion returns the first stdout line of the running Claude Code
// binary's --version, or "" when the guard refuses the path or the probe
// fails. Only an absolute path whose os.Stat (symlinks followed) is a regular
// file named claude or claude.exe runs. PATH is never searched, because the
// claude found there need not be the binary running this session.
func probeClaudeVersion(execPath string) string {
	if !filepath.IsAbs(execPath) {
		return ""
	}
	if base := filepath.Base(execPath); base != "claude" && base != "claude.exe" {
		return ""
	}
	if info, err := os.Stat(execPath); err != nil || !info.Mode().IsRegular() {
		return ""
	}
	output, err := runClaudeVersionProbe(execPath)
	if err != nil {
		return ""
	}
	first, _, _ := strings.Cut(output, "\n")
	return strings.TrimSuffix(first, "\r")
}

// execClaudeVersionProbe runs path directly, without a shell, with --version
// as its only argument, a 5 second deadline, and 4096 bytes per stream.
func execClaudeVersionProbe(path string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), claudeVersionProbeTimeout)
	defer cancel()
	output, err := processprobe.OutputLimited(exec.CommandContext(ctx, path, "--version"), claudeVersionProbeMaxBytes)
	return string(output), err
}
