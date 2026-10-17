//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const versionScript = "printf '2.1.288 (Claude Code)\\n'"

// pathDecoy puts a claude reporting 9.9.9 first on PATH and returns the file it
// creates when it runs, so a test proves PATH is never consulted.
func pathDecoy(t *testing.T) string {
	t.Helper()
	other := t.TempDir()
	touched := filepath.Join(other, "path.touched")
	script := "#!/bin/sh\n: > '" + touched + "'\nprintf '9.9.9 (Claude Code)\\n'\n"
	if err := os.WriteFile(filepath.Join(other, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", other+string(os.PathListSeparator)+os.Getenv("PATH"))
	return touched
}

func requireAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s exists", path)
	}
}

func TestWorkflowTriageTopologyS1SessionWithoutTeams(t *testing.T) {
	probe, count := fakeClaudeProbe(t, "claude", versionScript)
	touched := pathDecoy(t)
	env := capOKEnv(probe)
	delete(env, "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS")
	stubTeamEnv(t, env)
	d := decodeTopology(t, runTopology(t, topologyProject(t, cfgKO), factsU3)).Decision
	if d.Route != "subagent" || d.Reason != "teams_env_missing" || d.ReadyUnits == nil || *d.ReadyUnits != 3 || d.TeamUnits != nil || d.ToolsFallbackLine != nil {
		t.Fatalf("%+v", d)
	}
	if d.Line != "Agent Team: 유지 — 세션 환경에 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1이 없어 subagent로 실행한다" {
		t.Fatalf("%q", d.Line)
	}
	if probeRuns(t, count) != 0 {
		t.Fatal("probe ran")
	}
	requireAbsent(t, touched)
}

func TestWorkflowTriageTopologyS2S3PlanKeepsSubagentsWithoutProbing(t *testing.T) {
	probe, count := fakeClaudeProbe(t, "claude", versionScript)
	stubTeamEnv(t, capOKEnv(probe))
	project := topologyProject(t, cfgKO)
	u1d := strings.Replace(factsU1, `"owned_paths":["pkg/alpha"]}`, `"owned_paths":["pkg/alpha"],"done":true}`, 1)
	u1dd := strings.Replace(u1d, `"depends_on":["T1"]}`, `"depends_on":["T1"],"done":true}`, 1)
	ugo := `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/alpha/**"]},{"id":"T2","execution":"worker","owned_paths":["pkg/alpha/b.go"]}]}`
	u3o := strings.TrimSuffix(factsU3, `]}`) + `,{"id":"T4","execution":"worker","owned_paths":["internal/beta/x"]}]}`
	cases := []struct {
		facts, reason, line string
		ready               int
	}{
		{factsU1, "fewer_than_two_ready", "Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다", 1},
		{u1d, "fewer_than_two_ready", "Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다", 1},
		{u1dd, "fewer_than_two_ready", "Agent Team: 유지 — 함께 실행할 준비된 slice가 0개라 subagent로 실행한다", 0},
		{factsUOV, "ownership_overlap", "Agent Team: 유지 — T1, T2의 owned_paths가 겹쳐 subagent로 실행한다", 2},
		{ugo, "ownership_overlap", "Agent Team: 유지 — T1, T2의 owned_paths가 겹쳐 subagent로 실행한다", 2},
		{u3o, "ownership_overlap", "Agent Team: 유지 — T2, T4의 owned_paths가 겹쳐 subagent로 실행한다", 4},
	}
	for _, tc := range cases {
		d := decodeTopology(t, runTopology(t, project, tc.facts)).Decision
		if d.Route != "subagent" || d.Reason != tc.reason || d.ReadyUnits == nil || *d.ReadyUnits != tc.ready || d.Line != tc.line {
			t.Fatalf("%s: %+v", tc.facts, d)
		}
	}
	if probeRuns(t, count) != 0 {
		t.Fatal("probe ran")
	}
}

func TestWorkflowTriageTopologyS4TeamProbesOncePerRun(t *testing.T) {
	probe, count := fakeClaudeProbe(t, "claude", versionScript)
	touched := pathDecoy(t)
	stubTeamEnv(t, capOKEnv(probe))
	project := topologyProject(t, cfgKO)
	ugl := `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/a/**"]},{"id":"T2","execution":"worker","owned_paths":["pkg/ab/**"]}]}`
	first := runTopology(t, project, factsUDJ)
	second := runTopology(t, project, factsUDJ)
	if first.stdout != second.stdout {
		t.Fatalf("UDJ runs differ:\n%s\n%s", first.stdout, second.stdout)
	}
	for facts, followup := range map[string]string{factsUDJ: "T3", ugl: ""} {
		d := decodeTopology(t, runTopology(t, project, facts)).Decision
		if d.Route != "team" || d.Reason != "independent_disjoint_slices" || d.ReadyUnits == nil || *d.ReadyUnits != 2 || strings.Join(d.TeamUnits, ",") != "T1,T2" || d.FollowupUnits == nil || strings.Join(*d.FollowupUnits, ",") != followup {
			t.Fatalf("%s: %+v", facts, d)
		}
		if d.Line != "Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다" || d.ToolsFallbackLine == nil || *d.ToolsFallbackLine != "Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다" {
			t.Fatalf("%+v", d)
		}
	}
	if runs := probeRuns(t, count); runs != 4 {
		t.Fatalf("probe ran %d times over 4 runs", runs)
	}
	requireAbsent(t, touched)
}

func TestWorkflowTriageTopologyS4bFollowupUnits(t *testing.T) {
	probe, _ := fakeClaudeProbe(t, "claude", versionScript)
	stubTeamEnv(t, capOKEnv(probe))
	project := topologyProject(t, cfgKO)
	ufu := `{"version":1,"units":[{"id":"T4","execution":"worker","owned_paths":["pkg/d"],"done":true},{"id":"T5","execution":"worker","owned_paths":["pkg/e"]},{"id":"T6","execution":"worker","owned_paths":["pkg/f"]},{"id":"T7","execution":"worker","owned_paths":["pkg/g"],"depends_on":["T5"]},{"id":"T8","execution":"inline","owned_paths":["docs/h.md"]}]}`
	ufu7 := strings.Replace(ufu, `"depends_on":["T5"]}`, `"depends_on":["T5"],"done":true}`, 1)
	ufu78 := strings.Replace(ufu7, `["docs/h.md"]}`, `["docs/h.md"],"done":true}`, 1)
	for facts, followup := range map[string]string{ufu: "T7,T8", ufu7: "T8", ufu78: ""} {
		d := decodeTopology(t, runTopology(t, project, facts)).Decision
		if d.Route != "team" || *d.ReadyUnits != 2 || strings.Join(d.TeamUnits, ",") != "T5,T6" || d.FollowupUnits == nil || strings.Join(*d.FollowupUnits, ",") != followup {
			t.Fatalf("%s: %+v", facts, d)
		}
	}
}

func TestWorkflowTriageTopologyS5CapabilitySignals(t *testing.T) {
	touched := pathDecoy(t)
	project := topologyProject(t, cfgKO)
	unknown := func(x string) string {
		return "Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=" + x + ") subagent로 실행한다"
	}
	notInteractive := func(e, a string) string {
		return "Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=" + e + ", attended=" + a + ") subagent로 실행한다"
	}
	relative := filepath.Join(project, "probe", "claude")
	if err := os.MkdirAll(filepath.Dir(relative), 0o755); err != nil {
		t.Fatal(err)
	}
	relativeCount := filepath.Join(project, "probe", "probe.count")
	if err := os.WriteFile(relative, []byte("#!/bin/sh\nprintf 'run\\n' >> '"+relativeCount+"'\n"+versionScript+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, script, probeName string
		mutate                  func(env map[string]string)
		reason, line            string
		runs                    int
	}{
		{"sdk entrypoint", versionScript, "claude", func(env map[string]string) { env["CLAUDE_CODE_ENTRYPOINT"] = "sdk-cli" }, "not_interactive", notInteractive("sdk-cli", "1"), 0},
		{"unattended", versionScript, "claude", func(env map[string]string) { delete(env, "CLAUDE_CODE_SESSION_ATTENDED") }, "not_interactive", notInteractive("cli", "unset"), 0},
		{"entrypoint injection", versionScript, "claude", func(env map[string]string) { env["CLAUDE_CODE_ENTRYPOINT"] = "cli; rm -rf /" }, "not_interactive", notInteractive("unknown", "1"), 0},
		// CAP-OK carries CLAUDE_CODE_CHILD_SESSION=1 like a real top-level
		// session; the marker is not read, so dropping it changes nothing.
		{"child session marker absent", versionScript, "claude", func(env map[string]string) { delete(env, "CLAUDE_CODE_CHILD_SESSION") }, "independent_disjoint_slices", "Agent Team: 전환 — 준비된 slice 3개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다", 1},
		{"not claude code", versionScript, "claude", func(env map[string]string) { delete(env, "CLAUDECODE") }, "not_claude_code", "Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다", 0},
		{"rule 5 before 6", versionScript, "claude", func(env map[string]string) {
			delete(env, "CLAUDECODE")
			delete(env, "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS")
		}, "not_claude_code", "Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다", 0},
		{"old", "printf '2.1.245 (Claude Code)\\n'", "claude", nil, "claude_version_old", "Agent Team: 유지 — Claude Code 2.1.245 버전은 2.1.246보다 낮아 subagent로 실행한다", 1},
		{"minimum", "printf '2.1.246 (Claude Code)\\n'", "claude", nil, "independent_disjoint_slices", "Agent Team: 전환 — 준비된 slice 3개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다", 1},
		{"integer minor", "printf '2.10.0 (Claude Code)\\n'", "claude.exe", nil, "independent_disjoint_slices", "Agent Team: 전환 — 준비된 slice 3개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다", 1},
		{"node", "printf 'v24.20.0\\n'", "node", nil, "claude_version_unknown", unknown("node"), 0},
		{"relative", versionScript, "claude", func(env map[string]string) { env["CLAUDE_CODE_EXECPATH"] = filepath.Join("probe", "claude") }, "claude_version_unknown", unknown("claude"), 0},
		{"hang", "sleep 30", "claude", nil, "claude_version_unknown", unknown("claude"), 1},
		{"unset", versionScript, "claude", func(env map[string]string) { delete(env, "CLAUDE_CODE_EXECPATH") }, "claude_version_unknown", unknown("unset"), 0},
	}
	for _, tc := range cases {
		probe, count := fakeClaudeProbe(t, tc.probeName, tc.script)
		env := capOKEnv(probe)
		if tc.mutate != nil {
			tc.mutate(env)
		}
		stubTeamEnv(t, env)
		started := time.Now()
		d := decodeTopology(t, runTopology(t, project, factsU3)).Decision
		if elapsed := time.Since(started); elapsed > 20*time.Second {
			t.Fatalf("%s: took %s", tc.name, elapsed)
		}
		if d.Reason != tc.reason || d.Line != tc.line || probeRuns(t, count) != tc.runs {
			t.Fatalf("%s: %+v, runs %d", tc.name, d, probeRuns(t, count))
		}
		if tc.reason == "independent_disjoint_slices" && (d.Route != "team" || *d.ReadyUnits != 3) {
			t.Fatalf("%s: %+v", tc.name, d)
		}
	}
	if probeRuns(t, relativeCount) != 0 {
		t.Fatal("the relative EXECPATH ran")
	}
	requireAbsent(t, touched)
}
