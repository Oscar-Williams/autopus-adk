package taskroute

import (
	"strings"
	"testing"
)

func TestDecideTeamLinesFollowReasonLineContract(t *testing.T) {
	overlap := []TeamUnit{teamWorker("T1", "pkg/alpha"), teamWorker("T2", "pkg/alpha/b.go")}
	u3o := append(shapeU3(), teamWorker("T4", "internal/beta/x"))
	u1dd := shapeU1()
	u1dd[0].Done, u1dd[1].Done = true, true
	withCap := func(mutate func(*Capability)) Capability {
		c := capOK()
		mutate(&c)
		return c
	}
	cases := []struct {
		name   string
		in     TeamInput
		ko, en string
	}{
		{"team flag", TeamInput{Flags: TopologyFlags{Team: true}, Units: shapeUDJ()},
			"Agent Team: 자동 선택 안 함 — --team 지정됨", "Agent Team: auto-selection skipped — --team was given"},
		{"workflow flag", TeamInput{Flags: TopologyFlags{Workflow: true}},
			"Agent Team: 자동 선택 안 함 — --workflow 지정됨", "Agent Team: auto-selection skipped — --workflow was given"},
		{"solo flag", TeamInput{Flags: TopologyFlags{Solo: true}},
			"Agent Team: 자동 선택 안 함 — --solo 지정됨", "Agent Team: auto-selection skipped — --solo was given"},
		{"config off", TeamInput{AutoTeam: boolRef(false), Units: shapeUDJ(), Capability: capOK()},
			"Agent Team: 유지 — autopus.yaml의 workflow.auto_team이 false라 subagent로 실행한다", "Agent Team: off — workflow.auto_team is false in autopus.yaml; running subagents"},
		{"one ready", TeamInput{Units: shapeU1(), Capability: capOK()},
			"Agent Team: 유지 — 함께 실행할 준비된 slice가 1개라 subagent로 실행한다", "Agent Team: off — 1 ready slice(s) can run together; running subagents"},
		{"zero ready", TeamInput{Units: u1dd, Capability: capOK()},
			"Agent Team: 유지 — 함께 실행할 준비된 slice가 0개라 subagent로 실행한다", "Agent Team: off — 0 ready slice(s) can run together; running subagents"},
		{"overlap first pair", TeamInput{Units: overlap, Capability: capOK()},
			"Agent Team: 유지 — T1, T2의 owned_paths가 겹쳐 subagent로 실행한다", "Agent Team: off — T1 and T2 have overlapping owned_paths; running subagents"},
		{"overlap later pair", TeamInput{Units: u3o, Capability: capOK()},
			"Agent Team: 유지 — T2, T4의 owned_paths가 겹쳐 subagent로 실행한다", "Agent Team: off — T2 and T4 have overlapping owned_paths; running subagents"},
		{"not claude code", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.ClaudeCode = "" })},
			"Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다", "Agent Team: off — not a Claude Code session; running subagents"},
		{"teams env missing", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.AgentTeams = "" })},
			"Agent Team: 유지 — 세션 환경에 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1이 없어 subagent로 실행한다", "Agent Team: off — CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is not in the session environment; running subagents"},
		{"sdk entrypoint", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.Entrypoint = "sdk-cli" })},
			"Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=sdk-cli, attended=1) subagent로 실행한다", "Agent Team: off — not an interactive CLI session (entrypoint=sdk-cli, attended=1); running subagents"},
		{"attended unset", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.SessionAttended = "" })},
			"Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=cli, attended=unset) subagent로 실행한다", "Agent Team: off — not an interactive CLI session (entrypoint=cli, attended=unset); running subagents"},
		{"entrypoint injection", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.Entrypoint = "cli; rm -rf /" })},
			"Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint=unknown, attended=1) subagent로 실행한다", "Agent Team: off — not an interactive CLI session (entrypoint=unknown, attended=1); running subagents"},
		{"version unknown node", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.ExecPathBase, c.Version = "node", "" })},
			"Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=node) subagent로 실행한다", "Agent Team: off — the running Claude Code version could not be read (execpath=node); running subagents"},
		{"version unknown unset", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.ExecPathBase, c.Version = "", "" })},
			"Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=unset) subagent로 실행한다", "Agent Team: off — the running Claude Code version could not be read (execpath=unset); running subagents"},
		{"version unknown odd name", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.ExecPathBase, c.Version = "my claude", "" })},
			"Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath=unknown) subagent로 실행한다", "Agent Team: off — the running Claude Code version could not be read (execpath=unknown); running subagents"},
		{"version old", TeamInput{Units: shapeU3(), Capability: withCap(func(c *Capability) { c.Version = "2.1.245 (Claude Code)" })},
			"Agent Team: 유지 — Claude Code 2.1.245 버전은 2.1.246보다 낮아 subagent로 실행한다", "Agent Team: off — Claude Code 2.1.245 is older than 2.1.246; running subagents"},
		{"team", TeamInput{Units: shapeUDJ(), Capability: capOK()},
			"Agent Team: 전환 — 준비된 slice 2개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다", "Agent Team: on — 2 ready slices have disjoint owned_paths; running them as a native Agent Team"},
	}
	for _, tc := range cases {
		for _, locale := range []string{"ko", "en", "ja", "", "KO"} {
			in := tc.in
			in.Locale = locale
			want := tc.en
			if locale == "ko" {
				want = tc.ko
			}
			if got := DecideTeam(in).Line; got != want {
				t.Fatalf("%s/%q:\n got %q\nwant %q", tc.name, locale, got, want)
			}
		}
	}
}

func TestDecideTeamToolsFallbackLineOnlyForAutomaticTeam(t *testing.T) {
	ko := DecideTeam(TeamInput{Units: shapeUDJ(), Capability: capOK(), Locale: "ko"})
	en := DecideTeam(TeamInput{Units: shapeUDJ(), Capability: capOK(), Locale: "en"})
	if ko.ToolsFallbackLine != "Agent Team: 유지 — SendMessage 도구나 Agent의 name 인자를 쓸 수 없어 subagent로 실행한다" {
		t.Fatalf("%q", ko.ToolsFallbackLine)
	}
	if en.ToolsFallbackLine != "Agent Team: off — the SendMessage tool or the Agent name parameter is unavailable; running subagents" {
		t.Fatalf("%q", en.ToolsFallbackLine)
	}
	for _, in := range []TeamInput{{Flags: TopologyFlags{Team: true}, Units: shapeUDJ(), Capability: capOK()}, {Units: shapeU1(), Capability: capOK()}} {
		if d := DecideTeam(in); d.ToolsFallbackLine != "" {
			t.Fatalf("%+v", d)
		}
	}
}

func TestTeamLineTokensAreNormalized(t *testing.T) {
	for value, want := range map[string]string{"": "unset", "cli": "cli", "sdk-cli": "sdk-cli", "1": "1", "CLI": "unknown", "cli\n": "unknown", strings.Repeat("a", 32): strings.Repeat("a", 32), strings.Repeat("a", 33): "unknown", "a_b": "unknown"} {
		if got := sessionToken(value); got != want {
			t.Fatalf("session %q: %q", value, got)
		}
	}
	for value, want := range map[string]string{"": "unset", "claude": "claude", "claude.exe": "claude.exe", "Claude_1.2-x": "Claude_1.2-x", "my claude": "unknown", "/": "unknown", strings.Repeat("c", 33): "unknown", "클로드": "unknown"} {
		if got := execPathToken(value); got != want {
			t.Fatalf("execpath %q: %q", value, got)
		}
	}
}

func TestEveryReasonLineIsOneTaggedLine(t *testing.T) {
	if len(teamLineText) != 12 {
		t.Fatalf("contract has 12 reasons, table has %d", len(teamLineText))
	}
	// A child-session marker cannot tell the top-level session from a subagent.
	if _, ok := teamLineText["not_top_level"]; ok {
		t.Fatal("not_top_level is no longer a reason")
	}
	for reason := range teamLineText {
		for _, locale := range []string{"ko", "en"} {
			line := renderTeamLine(locale, reason, lineValues{n: 2, a: "T1", b: "T2", entrypoint: "x\ny", execBase: "a\nb", version: "2.1.0", flag: "--solo", code: "owned_paths"})
			if !strings.HasPrefix(line, "Agent Team: ") || strings.ContainsAny(line, "\r\n{}") {
				t.Fatalf("%s/%s: %q", reason, locale, line)
			}
		}
	}
}

func TestInvalidFactsDecisionNamesTheRejectedField(t *testing.T) {
	for _, code := range []string{"json", "version", "topology_conflict", "unit_id", "unit_count", "execution", "owned_paths", "depends_on"} {
		ko, en := InvalidFactsDecision("ko", code), InvalidFactsDecision("en", code)
		if ko.Version != 1 || ko.Route != "subagent" || ko.Reason != "invalid_facts" || ko.FactsError != code || ko.ReadyUnits != nil || ko.TeamUnits != nil || ko.FollowupUnits != nil || ko.ToolsFallbackLine != "" {
			t.Fatalf("%+v", ko)
		}
		if ko.Line != "Agent Team: 유지 — topology facts가 잘못되어("+code+") subagent로 실행한다" {
			t.Fatalf("%q", ko.Line)
		}
		if en.Line != "Agent Team: off — the topology facts are invalid ("+code+"); running subagents" {
			t.Fatalf("%q", en.Line)
		}
	}
	if d := InvalidFactsDecision("en", "x\ny"); d.FactsError != "unknown" || d.Line != "Agent Team: off — the topology facts are invalid (unknown); running subagents" {
		t.Fatalf("an unlisted code must not reach the line: %+v", d)
	}
}
