package cli

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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

// observeTeamCapability builds the capability snapshot from the five session
// variables the Capability Observation Contract names and returns the raw
// CLAUDE_CODE_EXECPATH for the version probe. No other variable is read, so a
// secret such as CLAUDE_CODE_MESSAGING_TOKEN never enters the snapshot.
// CLAUDE_CODE_CHILD_SESSION is not one of them: Claude Code sets it to 1 for
// every process a session spawns, the top-level session's Bash tool included.
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
	}
	execPath := read("CLAUDE_CODE_EXECPATH")
	if execPath != "" {
		snapshot.ExecPathBase = filepath.Base(execPath)
	}
	return snapshot, execPath
}

// probeClaudeVersion returns the first stdout line of the running Claude Code
// binary's --version, or "" when the guard refuses the path or the probe
// fails. Only a path claudeVersionProbePathAllowed accepts whose os.Stat
// (symlinks followed) is a regular file runs. PATH is never searched, because
// the claude found there need not be the binary running this session.
func probeClaudeVersion(execPath string) string {
	if !claudeVersionProbePathAllowed(runtime.GOOS, execPath) {
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

// claudeVersionProbePathAllowed judges the path text alone, so the rule of
// every platform is testable on any host. Elsewhere it is an absolute path
// named claude or claude.exe. On Windows, exec appends the PATHEXT extensions
// to a name without one, so an extensionless claude could run a sibling
// claude.cmd through cmd.exe: only a drive-letter path naming claude.exe runs,
// and UNC, \\?\ and \\.\ device paths, and alternate data streams are refused.
func claudeVersionProbePathAllowed(goos, execPath string) bool {
	if goos != "windows" {
		base := path.Base(execPath)
		return strings.HasPrefix(execPath, "/") && (base == "claude" || base == "claude.exe")
	}
	if len(execPath) < 3 || execPath[1] != ':' || (execPath[2] != '\\' && execPath[2] != '/') {
		return false
	}
	drive := execPath[0]
	if !(drive >= 'a' && drive <= 'z' || drive >= 'A' && drive <= 'Z') || strings.Contains(execPath[2:], ":") {
		return false
	}
	return execPath[strings.LastIndexAny(execPath, `\/`)+1:] == "claude.exe"
}

// execClaudeVersionProbe runs path directly, without a shell, with --version
// as its only argument, a 5 second deadline, and 4096 bytes per stream.
func execClaudeVersionProbe(path string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), claudeVersionProbeTimeout)
	defer cancel()
	output, err := processprobe.OutputLimited(exec.CommandContext(ctx, path, "--version"), claudeVersionProbeMaxBytes)
	return string(output), err
}
