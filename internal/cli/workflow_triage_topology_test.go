package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Facts and configs named in acceptance.md, spelled as the session writes them.
const (
	factsU3  = `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/alpha"]},{"id":"T2","execution":"worker","owned_paths":["internal/beta"]},{"id":"T3","execution":"worker","owned_paths":["docs/gamma.md"]}]}`
	factsU1  = `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/alpha"]},{"id":"T2","execution":"worker","owned_paths":["pkg/beta"],"depends_on":["T1"]},{"id":"T3","execution":"inline","owned_paths":["README.md"]}]}`
	factsUOV = `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/alpha"]},{"id":"T2","execution":"worker","owned_paths":["pkg/alpha/b.go"]}]}`
	factsUDJ = `{"version":1,"flags":{"team":false,"workflow":false,"solo":false},"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/a"],"depends_on":[],"done":false},{"id":"T2","execution":"worker","owned_paths":["pkg/ab"],"depends_on":[],"done":false},{"id":"T3","execution":"worker","owned_paths":["pkg/c"],"depends_on":["T1"],"done":false},{"id":"T4","execution":"worker","owned_paths":["pkg/d"],"depends_on":[],"done":true}]}`

	cfgBase = "mode: full\nproject_name: proj\nplatforms:\n  - claude-code\n"
	cfgKO   = cfgBase + "language:\n  ai_responses: ko\n"
	cfgEN   = cfgBase + "language:\n  ai_responses: en\n"
	cfgJA   = cfgBase + "language:\n  ai_responses: ja\n"
	cfgLEG  = "mode: full\nproject_name: proj\nplatforms: [claude]\nlanguage:\n  ai_responses: ko\n"
)

// capOKEnv is CAP-OK without the EXECPATH, which each test points at its own
// probe. Like a real top-level session's Bash tool, it has the child-session
// marker Claude Code sets for every process a session spawns.
func capOKEnv(execPath string) map[string]string {
	env := map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1", "CLAUDE_CODE_ENTRYPOINT": "cli", "CLAUDE_CODE_SESSION_ATTENDED": "1", "CLAUDE_CODE_CHILD_SESSION": "1"}
	if execPath != "" {
		env["CLAUDE_CODE_EXECPATH"] = execPath
	}
	return env
}

type topologyRun struct {
	stdout, stderr string
	code           int
}

// topologyProject returns a project directory holding autopus.yaml, or no
// config at all when yaml is empty.
func topologyProject(t *testing.T, yaml string) string {
	t.Helper()
	dir := t.TempDir()
	if yaml != "" {
		if err := os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// runTopology runs `auto workflow triage --topology` through the real root
// command from dir, so stderr carries the binary's `Error: <msg>` line.
func runTopology(t *testing.T, dir, facts string, args ...string) topologyRun {
	t.Helper()
	factsPath := filepath.Join(t.TempDir(), "f.json")
	if err := os.WriteFile(factsPath, []byte(facts), 0o600); err != nil {
		t.Fatal(err)
	}
	return runTopologyArgs(t, dir, append([]string{"workflow", "triage", "--topology", "--facts-json", factsPath}, args...)...)
}

func runTopologyArgs(t *testing.T, dir string, args ...string) topologyRun {
	t.Helper()
	t.Chdir(dir)
	root := NewRootCmd()
	root.SetArgs(args)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	code, _ := runRootCommand(context.Background(), root, &stderr)
	return topologyRun{stdout.String(), stderr.String(), code}
}

type topologyDecision struct {
	Route             string    `json:"route"`
	Reason            string    `json:"reason"`
	FactsError        string    `json:"facts_error"`
	ReadyUnits        *int      `json:"ready_units"`
	TeamUnits         []string  `json:"team_units"`
	FollowupUnits     *[]string `json:"followup_units"`
	Line              string    `json:"line"`
	ToolsFallbackLine *string   `json:"tools_fallback_line"`
}

type topologyReport struct {
	Mode               string           `json:"mode"`
	UnitSource         string           `json:"unit_source"`
	CapabilitySource   string           `json:"capability_source"`
	ConfigSource       string           `json:"config_source"`
	ExecutionPerformed *bool            `json:"execution_performed"`
	Decision           topologyDecision `json:"decision"`
}

func decodeTopology(t *testing.T, run topologyRun) topologyReport {
	t.Helper()
	if run.code != 0 || run.stderr != "" {
		t.Fatalf("exit %d stderr %q", run.code, run.stderr)
	}
	var report topologyReport
	if err := json.Unmarshal([]byte(run.stdout), &report); err != nil {
		t.Fatalf("%v: %s", err, run.stdout)
	}
	return report
}

// Rejected facts are a subagent decision whose line names the rejected field:
// the generated snippet discards stderr, where an error would go unseen.
func TestWorkflowTriageTopologyReportsInvalidFactsAsSubagent(t *testing.T) {
	ran := stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	stubTeamEnv(t, capOKEnv("/opt/claude"))
	worker := func(fields string) string {
		return `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["a"]` + fields + `}]}`
	}
	owned := func(path string) string {
		raw, err := json.Marshal(map[string]any{"version": 1, "units": []any{map[string]any{"id": "T1", "execution": "worker", "owned_paths": []string{path}}}})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	many := make([]string, 33)
	for i := range many {
		many[i] = fmt.Sprintf(`{"id":"U%d","execution":"worker","owned_paths":["p/%d"]}`, i, i)
	}
	cases := map[string][]string{
		"version":           {`{"version":2,"units":[]}`},
		"depends_on":        {worker(`,"depends_on":["T9"]`), worker(`,"depends_on":["T1"]`)},
		"json":              {worker(`,"id":"T2"`), `{"version":1,"units":[],"extra":1}`, worker(`,"OWNED_PATHS":["pkg/z"]`), `{"version":1,"flags":{"TEAM":true},"units":[]}`},
		"owned_paths":       {`{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":[]}]}`, owned("../x"), owned("/abs"), owned("pkg/*/x"), owned("pkg/a/**/**"), owned("./**"), owned("pkg/$(id)/x"), owned("pkg/" + "`" + "id" + "`" + "/x"), owned("pkg/a;b")},
		"execution":         {`{"version":1,"units":[{"id":"T1","execution":"parallel","owned_paths":["a"]}]}`},
		"unit_count":        {`{"version":1,"units":[` + strings.Join(many, ",") + `]}`},
		"unit_id":           {`{"version":1,"units":[{"id":"T 1","execution":"worker","owned_paths":["a"]}]}`},
		"topology_conflict": {`{"version":1,"flags":{"solo":true,"team":true},"units":[]}`},
	}
	project := topologyProject(t, cfgKO)
	for code, documents := range cases {
		for _, document := range documents {
			d := decodeTopology(t, runTopology(t, project, document)).Decision
			want := "Agent Team: 유지 — topology facts가 잘못되어(" + code + ") subagent로 실행한다"
			if d.Route != "subagent" || d.Reason != "invalid_facts" || d.FactsError != code || d.Line != want || d.ReadyUnits != nil || d.TeamUnits != nil || d.FollowupUnits != nil || d.ToolsFallbackLine != nil {
				t.Fatalf("%s %s: %+v", code, document, d)
			}
		}
	}
	human := runTopology(t, project, `{"version":2}`, "--format", "human")
	if human.code != 0 || human.stderr != "" || human.stdout != "Agent Team: 유지 — topology facts가 잘못되어(version) subagent로 실행한다\n" {
		t.Fatalf("%+v", human)
	}
	if len(*ran) != 0 {
		t.Fatalf("invalid facts probed %v", *ran)
	}
}

func TestWorkflowTriageTopologyKeepsFileGuardMessages(t *testing.T) {
	project := topologyProject(t, cfgKO)
	oversize := filepath.Join(t.TempDir(), "big.json")
	if err := os.WriteFile(oversize, make([]byte, (1<<20)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{project, filepath.Join(project, "missing.json"), oversize} {
		run := runTopologyArgs(t, project, "workflow", "triage", "--topology", "--facts-json", path)
		if run.code != 1 || run.stdout != "" || run.stderr != "Error: facts file must be a readable regular file at most 1 MiB\n" {
			t.Fatalf("%s: %+v", path, run)
		}
	}
	for _, args := range [][]string{{"--format", "xml"}, {}} {
		full := append([]string{"workflow", "triage", "--topology"}, args...)
		if len(args) > 0 {
			full = append(full, "--facts-json", oversize)
		}
		if run := runTopologyArgs(t, project, full...); run.code != 1 || run.stdout != "" {
			t.Fatalf("%v: %+v", args, run)
		}
	}
}

func TestWorkflowTriageTopologyWithoutFlagKeepsLegacyMode(t *testing.T) {
	project := topologyProject(t, cfgKO)
	factsPath := filepath.Join(t.TempDir(), "f.json")
	if err := os.WriteFile(factsPath, []byte(factsUDJ), 0o600); err != nil {
		t.Fatal(err)
	}
	run := runTopologyArgs(t, project, "workflow", "triage", "--facts-json", factsPath)
	if run.code != 1 || run.stdout != "" || run.stderr != "Error: invalid workflow triage facts JSON\n" {
		t.Fatalf("%+v", run)
	}
}

func TestWorkflowTriageTopologyConfigSources(t *testing.T) {
	stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	stubTeamEnv(t, capOKEnv(writeProbeTarget(t)))
	teamEN := "Agent Team: on — 2 ready slices have disjoint owned_paths; running them as a native Agent Team"
	fallbackEN := "Agent Team: off — SendMessage/task tool schemas could not be loaded; running subagents"
	for name, yaml := range map[string]string{"en": cfgEN, "ja": cfgJA, "none": ""} {
		report := decodeTopology(t, runTopology(t, topologyProject(t, yaml), factsUDJ))
		if report.Decision.Line != teamEN || report.Decision.ToolsFallbackLine == nil || *report.Decision.ToolsFallbackLine != fallbackEN {
			t.Fatalf("%s: %+v", name, report.Decision)
		}
		want := "autopus.yaml"
		if yaml == "" {
			want = "default"
		}
		if report.ConfigSource != want {
			t.Fatalf("%s: config_source %q", name, report.ConfigSource)
		}
	}
	if line := decodeTopology(t, runTopology(t, topologyProject(t, cfgEN), factsU1)).Decision.Line; line != "Agent Team: off — 1 ready slice(s) can run together; running subagents" {
		t.Fatalf("%q", line)
	}
	off := decodeTopology(t, runTopology(t, topologyProject(t, cfgEN+"workflow:\n  auto_team: false\n"), factsUDJ))
	if off.Decision.Route != "subagent" || off.Decision.Reason != "config_off" || off.Decision.Line != "Agent Team: off — workflow.auto_team is false in autopus.yaml; running subagents" {
		t.Fatalf("%+v", off.Decision)
	}
}

func TestWorkflowTriageTopologyConfigValues(t *testing.T) {
	ran := stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	stubTeamEnv(t, capOKEnv(writeProbeTarget(t)))
	for _, value := range []string{"false", "FALSE"} {
		report := decodeTopology(t, runTopology(t, topologyProject(t, cfgKO+"workflow:\n  auto_team: "+value+"\n"), factsUDJ))
		if report.Decision.Reason != "config_off" || report.Decision.Line != "Agent Team: 유지 — autopus.yaml의 workflow.auto_team이 false라 subagent로 실행한다" {
			t.Fatalf("%s: %+v", value, report.Decision)
		}
	}
	if len(*ran) != 0 {
		t.Fatalf("config_off probed %v", *ran)
	}
	for _, value := range []string{"true", `!!bool "true"`, ""} {
		yaml := cfgKO
		if value != "" {
			yaml += "workflow:\n  auto_team: " + value + "\n"
		}
		if report := decodeTopology(t, runTopology(t, topologyProject(t, yaml), factsUDJ)); report.Decision.Route != "team" {
			t.Fatalf("%q: %+v", value, report.Decision)
		}
	}
	// cfgKO spans six lines, so the key under workflow: sits on line 8.
	for _, value := range []string{`"false"`, "null", "~", "", "yes", `"off"`, "1"} {
		run := runTopology(t, topologyProject(t, cfgKO+"workflow:\n  auto_team: "+value+"\n"), factsUDJ)
		if run.code != 1 || run.stdout != "" || !strings.HasPrefix(run.stderr, "Error: config load failed: ") || !strings.Contains(run.stderr, "workflow.auto_team: must be true or false (line 8)") {
			t.Fatalf("%q: %+v", value, run)
		}
	}
}

// writeProbeTarget returns an existing regular file the guard runs on this
// host, so the stubbed probe seam decides the version.
func writeProbeTarget(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), probeExecutableName)
	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
