//go:build !windows

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeClaudeProbe writes an executable script called name into a fresh probe
// directory. Every run appends one line to probe.count before body runs, so a
// test counts real executions instead of trusting a seam. The script runs once
// before it is returned, uncounted: the decision gives the probe 5 seconds,
// and under load the first exec of a new file can spend that on macOS
// first-run assessment.
func fakeClaudeProbe(t *testing.T, name, body string) (path, count string) {
	t.Helper()
	dir := t.TempDir()
	count = filepath.Join(dir, "probe.count")
	path = filepath.Join(dir, name)
	script := "#!/bin/sh\n[ \"${AUTOPUS_PROBE_WARMUP:-}\" = 1 ] && exit 0\nprintf 'run\\n' >> '" + count + "'\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	warm := exec.Command(path)
	warm.Env = []string{"AUTOPUS_PROBE_WARMUP=1"}
	if err := warm.Run(); err != nil {
		t.Fatal(err)
	}
	return path, count
}

// probeRuns counts the lines a fake probe appended; a missing file is zero.
func probeRuns(t *testing.T, count string) int {
	t.Helper()
	data, err := os.ReadFile(count)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return strings.Count(string(data), "\n")
}

func TestProbeClaudeVersionExecutesTheScriptWithoutShellOrPath(t *testing.T) {
	path, count := fakeClaudeProbe(t, "claude", `[ "$#" = 1 ] && [ "$1" = --version ] || exit 7
printf '2.1.288 (Claude Code)\n'`)
	if got := probeClaudeVersion(path); got != "2.1.288 (Claude Code)" || probeRuns(t, count) != 1 {
		t.Fatalf("%q, runs %d", got, probeRuns(t, count))
	}
}

func TestProbeClaudeVersionFailuresReportNoVersion(t *testing.T) {
	cases := map[string]string{
		"non-zero exit":   "printf '2.1.288 (Claude Code)\\n'; exit 3",
		"stdout overflow": "i=0; while [ $i -lt 300 ]; do printf '0123456789abcdef'; i=$((i+1)); done",
		"stderr overflow": "i=0; while [ $i -lt 300 ]; do printf '0123456789abcdef' >&2; i=$((i+1)); done; printf '2.1.288 (Claude Code)\\n'",
	}
	for name, body := range cases {
		path, count := fakeClaudeProbe(t, "claude", body)
		if got := probeClaudeVersion(path); got != "" || probeRuns(t, count) != 1 {
			t.Fatalf("%s: %q, runs %d", name, got, probeRuns(t, count))
		}
	}
}

func TestProbeClaudeVersionTimesOutAHangingBinary(t *testing.T) {
	path, count := fakeClaudeProbe(t, "claude", "sleep 30")
	started := time.Now()
	got := probeClaudeVersion(path)
	elapsed := time.Since(started)
	if got != "" || probeRuns(t, count) != 1 || elapsed > 20*time.Second {
		t.Fatalf("%q, runs %d, elapsed %s", got, probeRuns(t, count), elapsed)
	}
}

func TestProbeClaudeVersionNeverSearchesPath(t *testing.T) {
	other := t.TempDir()
	touched := filepath.Join(other, "path.touched")
	script := "#!/bin/sh\n: > '" + touched + "'\nprintf '9.9.9 (Claude Code)\\n'\n"
	if err := os.WriteFile(filepath.Join(other, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", other+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, value := range []string{"", "claude", "claude.exe"} {
		if got := probeClaudeVersion(value); got != "" {
			t.Fatalf("%q: %q", value, got)
		}
	}
	if _, err := os.Stat(touched); !os.IsNotExist(err) {
		t.Fatal("the claude on PATH ran")
	}
}
