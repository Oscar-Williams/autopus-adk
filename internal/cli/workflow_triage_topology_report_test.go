package cli

import (
	"crypto/sha256"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// treeDigest maps every file under root to its SHA-256, so a test can prove
// a run wrote, renamed, and deleted nothing.
func treeDigest(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	digest := map[string][32]byte{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		digest[rel] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return digest
}

func TestWorkflowTriageTopologyReportsWithoutWriting(t *testing.T) {
	ran := stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	stubTeamEnv(t, capOKEnv(writeProbeTarget(t)))
	project := topologyProject(t, cfgLEG)
	before := treeDigest(t, project)

	report := decodeTopology(t, runTopology(t, project, factsUDJ, "--format", "json"))
	if report.Mode != "topology" || report.UnitSource != "caller_declared" || report.CapabilitySource != "observed_process_env" || report.ConfigSource != "autopus.yaml" || report.ExecutionPerformed == nil || *report.ExecutionPerformed {
		t.Fatalf("%+v", report)
	}
	d := report.Decision
	if d.Route != "team" || d.Reason != "independent_disjoint_slices" || d.ReadyUnits == nil || *d.ReadyUnits != 2 || strings.Join(d.TeamUnits, ",") != "T1,T2" || d.FollowupUnits == nil || strings.Join(*d.FollowupUnits, ",") != "T3" {
		t.Fatalf("%+v", d)
	}
	if d.Line != "Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다" || d.ToolsFallbackLine == nil || *d.ToolsFallbackLine != "Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다" {
		t.Fatalf("%+v", d)
	}

	human := runTopology(t, project, factsUDJ, "--format", "human")
	if human.code != 0 || human.stderr != "" || human.stdout != "Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다\n" {
		t.Fatalf("%+v", human)
	}
	if after := treeDigest(t, project); !maps.Equal(before, after) {
		t.Fatalf("the project changed: %v -> %v", before, after)
	}
	if len(*ran) != 2 {
		t.Fatalf("probe ran %d times, want one per run", len(*ran))
	}
}

func TestWorkflowTriageTopologyExplicitFlagsSkipCapability(t *testing.T) {
	ran := stubVersionProbe(t, "2.1.288 (Claude Code)\n", nil)
	stubTeamEnv(t, capOKEnv(writeProbeTarget(t)))
	project := topologyProject(t, cfgKO)
	for flag, route := range map[string]string{"solo": "solo", "workflow": "workflow", "team": "team"} {
		facts := strings.Replace(factsUDJ, `"`+flag+`":false`, `"`+flag+`":true`, 1)
		report := decodeTopology(t, runTopology(t, project, facts))
		d := report.Decision
		if d.Route != route || d.Reason != "explicit_flag" || d.ReadyUnits != nil || d.TeamUnits != nil || d.FollowupUnits != nil || d.ToolsFallbackLine != nil || d.Line != "Agent Team: 자동 선택 안 함 — --"+flag+" 지정됨" {
			t.Fatalf("%s: %+v", flag, d)
		}
	}
	if len(*ran) != 0 {
		t.Fatalf("explicit flags probed %v", *ran)
	}
}

func TestWorkflowTriageHelpCarriesTheAutoTeamNotice(t *testing.T) {
	run := runTopologyArgs(t, t.TempDir(), "workflow", "triage", "--help")
	if run.code != 0 {
		t.Fatalf("%+v", run)
	}
	for _, item := range []string{"--topology", "on by default", "separate Claude session", "experimental", "workflow.auto_team: false", "--solo", "auto 0.50.113 through", "0.50.125 refuse", "auto 0.50.112 and older ignore the key and delete it on auto update", "CLAUDE_CODE_EXECPATH", "claude or claude.exe", "claude_version_unknown", "invalid_facts decision", "cannot tell a subagent from the top-level session", "tools_fallback_line"} {
		if !strings.Contains(run.stdout, item) {
			t.Fatalf("help lacks %q:\n%s", item, run.stdout)
		}
	}
	// Auto-selection cannot check for a top-level session, so help must not claim it does.
	if strings.Contains(run.stdout, "top-level Claude Code session") {
		t.Fatalf("help claims a top-level check:\n%s", run.stdout)
	}
}
