package taskroute

import (
	"reflect"
	"slices"
	"testing"
)

func teamWorker(id string, paths ...string) TeamUnit {
	return TeamUnit{ID: id, Execution: "worker", OwnedPaths: paths}
}

func teamInline(id string, paths ...string) TeamUnit {
	return TeamUnit{ID: id, Execution: "inline", OwnedPaths: paths}
}

func dependsOn(unit TeamUnit, ids ...string) TeamUnit {
	unit.DependsOn = ids
	return unit
}

func done(unit TeamUnit) TeamUnit {
	unit.Done = true
	return unit
}

// capOK is the CAP-OK snapshot of acceptance.md after a successful probe.
func capOK() Capability {
	return Capability{ClaudeCode: "1", AgentTeams: "1", Entrypoint: "cli", SessionAttended: "1", ExecPathBase: "claude", Version: "2.1.288 (Claude Code)"}
}

// Unit shapes named in acceptance.md.
func shapeU3() []TeamUnit {
	return []TeamUnit{teamWorker("T1", "pkg/alpha"), teamWorker("T2", "internal/beta"), teamWorker("T3", "docs/gamma.md")}
}
func shapeU1() []TeamUnit {
	return []TeamUnit{teamWorker("T1", "pkg/alpha"), dependsOn(teamWorker("T2", "pkg/beta"), "T1"), teamInline("T3", "README.md")}
}
func shapeUDJ() []TeamUnit {
	return []TeamUnit{teamWorker("T1", "pkg/a"), teamWorker("T2", "pkg/ab"), dependsOn(teamWorker("T3", "pkg/c"), "T1"), done(teamWorker("T4", "pkg/d"))}
}
func shapeUFU() []TeamUnit {
	return []TeamUnit{done(teamWorker("T4", "pkg/d")), teamWorker("T5", "pkg/e"), teamWorker("T6", "pkg/f"), dependsOn(teamWorker("T7", "pkg/g"), "T5"), teamInline("T8", "docs/h.md")}
}

func boolRef(value bool) *bool { return &value }

func TestDecideTeamExplicitFlagWinsWithoutReadingAnything(t *testing.T) {
	flags := []struct {
		flags TopologyFlags
		route string
	}{{TopologyFlags{Team: true}, "team"}, {TopologyFlags{Workflow: true}, "workflow"}, {TopologyFlags{Solo: true}, "solo"}}
	for _, flag := range flags {
		for _, autoTeam := range []*bool{nil, boolRef(true), boolRef(false)} {
			for _, units := range [][]TeamUnit{nil, shapeU1(), shapeU3(), shapeUDJ()} {
				for _, capability := range []Capability{{}, capOK()} {
					in := TeamInput{Flags: flag.flags, AutoTeam: autoTeam, Units: units, Capability: capability}
					d := DecideTeam(in)
					if d.Version != 1 || d.Route != flag.route || d.Reason != "explicit_flag" || d.ReadyUnits != nil || d.TeamUnits != nil || d.FollowupUnits != nil || d.ToolsFallbackLine != "" {
						t.Fatalf("%+v -> %+v", in, d)
					}
					if NeedsVersionProbe(in) {
						t.Fatalf("explicit flag must not probe: %+v", in)
					}
				}
			}
		}
	}
}

func TestDecideTeamConfigOffPrecedesPlanAndCapability(t *testing.T) {
	for _, units := range [][]TeamUnit{shapeU3(), shapeU1(), nil} {
		in := TeamInput{AutoTeam: boolRef(false), Units: units, Capability: capOK()}
		d := DecideTeam(in)
		if d.Route != "subagent" || d.Reason != "config_off" || d.ReadyUnits == nil || d.TeamUnits != nil || d.FollowupUnits != nil || NeedsVersionProbe(in) {
			t.Fatalf("%+v", d)
		}
	}
	for _, autoTeam := range []*bool{nil, boolRef(true)} {
		d := DecideTeam(TeamInput{AutoTeam: autoTeam, Units: shapeUDJ(), Capability: capOK()})
		if d.Route != "team" {
			t.Fatalf("enabled config must allow a team: %+v", d)
		}
	}
}

func TestDecideTeamReadySet(t *testing.T) {
	u1d := shapeU1()
	u1d[0].Done = true
	u1dd := shapeU1()
	u1dd[0].Done, u1dd[1].Done = true, true
	cases := []struct {
		name  string
		units []TeamUnit
		ready int
	}{
		{"U1 dependent worker waits", shapeU1(), 1},
		{"U1d dependency done", u1d, 1},
		{"U1dd only inline left", u1dd, 0},
		{"no units", nil, 0},
		{"inline never ready", []TeamUnit{teamInline("A", "a"), teamInline("B", "b")}, 0},
		{"done never ready", []TeamUnit{done(teamWorker("A", "a")), teamWorker("B", "b")}, 1},
		{"dependency on unfinished inline", []TeamUnit{teamInline("A", "a"), dependsOn(teamWorker("B", "b"), "A"), teamWorker("C", "c")}, 1},
		{"dependency on a missing id", []TeamUnit{dependsOn(teamWorker("A", "a"), "Z"), teamWorker("B", "b")}, 1},
	}
	for _, tc := range cases {
		in := TeamInput{Units: tc.units, Capability: capOK()}
		d := DecideTeam(in)
		if d.Route != "subagent" || d.Reason != "fewer_than_two_ready" || d.ReadyUnits == nil || *d.ReadyUnits != tc.ready || NeedsVersionProbe(in) {
			t.Fatalf("%s: %+v", tc.name, d)
		}
	}
}

func TestDecideTeamOwnershipOverlap(t *testing.T) {
	u3o := append(shapeU3(), teamWorker("T4", "internal/beta/x"))
	cases := []struct {
		name  string
		units []TeamUnit
		a, b  string
	}{
		{"UOV nested file", []TeamUnit{teamWorker("T1", "pkg/alpha"), teamWorker("T2", "pkg/alpha/b.go")}, "T1", "T2"},
		{"UGO glob root", []TeamUnit{teamWorker("T1", "pkg/alpha/**"), teamWorker("T2", "pkg/alpha/b.go")}, "T1", "T2"},
		{"U3o later pair", u3o, "T2", "T4"},
		{"same cleaned path", []TeamUnit{teamWorker("T1", "pkg/a/"), teamWorker("T2", "pkg//a")}, "T1", "T2"},
		// Pairs are ordered by unit position, not by the position of the path inside a unit.
		{"facts order of pairs", []TeamUnit{teamWorker("T1", "pkg/x", "pkg/y"), teamWorker("T2", "pkg/y/z"), teamWorker("T3", "pkg/x/w")}, "T1", "T2"},
	}
	for _, tc := range cases {
		in := TeamInput{Units: tc.units, Capability: capOK()}
		d := DecideTeam(in)
		if d.Route != "subagent" || d.Reason != "ownership_overlap" || NeedsVersionProbe(in) {
			t.Fatalf("%s: %+v", tc.name, d)
		}
		a, b, ok := firstOverlapPair(readyUnits(tc.units))
		if !ok || a != tc.a || b != tc.b {
			t.Fatalf("%s: pair %s %s %v", tc.name, a, b, ok)
		}
	}
	// Overlap only counts among ready units: a done or waiting unit shares paths freely.
	units := []TeamUnit{teamWorker("T1", "pkg/a"), teamWorker("T2", "pkg/b"), done(teamWorker("T3", "pkg/a")), dependsOn(teamWorker("T4", "pkg/b/c"), "T1")}
	if d := DecideTeam(TeamInput{Units: units, Capability: capOK()}); d.Route != "team" {
		t.Fatalf("non-ready units must not block a team: %+v", d)
	}
}

func TestDecideTeamAutomaticTeam(t *testing.T) {
	ufu7 := shapeUFU()
	ufu7[3].Done = true
	ufu78 := shapeUFU()
	ufu78[3].Done, ufu78[4].Done = true, true
	cases := []struct {
		name     string
		units    []TeamUnit
		team     []string
		followup []string
	}{
		{"UDJ prefix sibling is disjoint", shapeUDJ(), []string{"T1", "T2"}, []string{"T3"}},
		{"UGL trailing globs", []TeamUnit{teamWorker("T1", "pkg/a/**"), teamWorker("T2", "pkg/ab/**")}, []string{"T1", "T2"}, []string{}},
		{"UFU", shapeUFU(), []string{"T5", "T6"}, []string{"T7", "T8"}},
		{"UFU T7 done", ufu7, []string{"T5", "T6"}, []string{"T8"}},
		{"UFU T7 T8 done", ufu78, []string{"T5", "T6"}, []string{}},
		{"U3", shapeU3(), []string{"T1", "T2", "T3"}, []string{}},
	}
	for _, tc := range cases {
		in := TeamInput{Units: tc.units, Capability: capOK()}
		if !NeedsVersionProbe(in) {
			t.Fatalf("%s: rules 1-7 pass, the probe must run", tc.name)
		}
		d := DecideTeam(in)
		if d.Route != "team" || d.Reason != "independent_disjoint_slices" || d.ReadyUnits == nil || *d.ReadyUnits != len(tc.team) {
			t.Fatalf("%s: %+v", tc.name, d)
		}
		if !slices.Equal(d.TeamUnits, tc.team) || d.FollowupUnits == nil || !slices.Equal(*d.FollowupUnits, tc.followup) {
			t.Fatalf("%s: team %v followup %v", tc.name, d.TeamUnits, d.FollowupUnits)
		}
		if !reflect.DeepEqual(d, DecideTeam(in)) {
			t.Fatalf("%s: decision is not deterministic", tc.name)
		}
	}
}

func TestDecideTeamCapabilityRulesInOrder(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Capability)
		reason string
		probe  bool
	}{
		{"not claude code", func(c *Capability) { c.ClaudeCode = "" }, "not_claude_code", false},
		{"claude code zero", func(c *Capability) { c.ClaudeCode = "0" }, "not_claude_code", false},
		{"rule 5 before 6", func(c *Capability) { c.ClaudeCode, c.AgentTeams = "", "" }, "not_claude_code", false},
		{"teams env missing", func(c *Capability) { c.AgentTeams = "" }, "teams_env_missing", false},
		{"teams env true word", func(c *Capability) { c.AgentTeams = "true" }, "teams_env_missing", false},
		{"sdk entrypoint", func(c *Capability) { c.Entrypoint = "sdk-cli" }, "not_interactive", false},
		{"unattended", func(c *Capability) { c.SessionAttended = "" }, "not_interactive", false},
		{"injected entrypoint", func(c *Capability) { c.Entrypoint = "cli; rm -rf /" }, "not_interactive", false},
		{"probe not run", func(c *Capability) { c.Version = "" }, "claude_version_unknown", true},
		{"node version", func(c *Capability) { c.Version = "v24.20.0" }, "claude_version_unknown", true},
		{"missing suffix", func(c *Capability) { c.Version = "2.1.288" }, "claude_version_unknown", true},
		{"trailing space", func(c *Capability) { c.Version = "2.1.288 (Claude Code) " }, "claude_version_unknown", true},
		{"overflow", func(c *Capability) { c.Version = "99999999999999999999.1.1 (Claude Code)" }, "claude_version_unknown", true},
		{"old patch", func(c *Capability) { c.Version = "2.1.245 (Claude Code)" }, "claude_version_old", true},
		{"old major", func(c *Capability) { c.Version = "1.9.999 (Claude Code)" }, "claude_version_old", true},
		{"minimum", func(c *Capability) { c.Version = "2.1.246 (Claude Code)" }, "independent_disjoint_slices", true},
		{"integer minor", func(c *Capability) { c.Version = "2.10.0 (Claude Code)" }, "independent_disjoint_slices", true},
		{"newer major", func(c *Capability) { c.Version = "3.0.0 (Claude Code)" }, "independent_disjoint_slices", true},
	}
	for _, tc := range cases {
		capability := capOK()
		tc.mutate(&capability)
		in := TeamInput{Units: shapeU3(), Capability: capability}
		if NeedsVersionProbe(in) != tc.probe {
			t.Fatalf("%s: NeedsVersionProbe = %v", tc.name, !tc.probe)
		}
		d := DecideTeam(in)
		if d.Reason != tc.reason || d.ReadyUnits == nil || *d.ReadyUnits != 3 {
			t.Fatalf("%s: %+v", tc.name, d)
		}
		wantRoute := "subagent"
		if tc.reason == "independent_disjoint_slices" {
			wantRoute = "team"
		}
		if d.Route != wantRoute || (wantRoute == "subagent") != (d.TeamUnits == nil && d.FollowupUnits == nil) {
			t.Fatalf("%s: %+v", tc.name, d)
		}
	}
}

func TestDecideTeamPlanPrecedesCapability(t *testing.T) {
	noEnv := Capability{}
	if d := DecideTeam(TeamInput{Units: shapeU1(), Capability: noEnv}); d.Reason != "fewer_than_two_ready" {
		t.Fatalf("%+v", d)
	}
	overlap := []TeamUnit{teamWorker("T1", "pkg/alpha"), teamWorker("T2", "pkg/alpha/b.go")}
	if d := DecideTeam(TeamInput{Units: overlap, Capability: noEnv}); d.Reason != "ownership_overlap" {
		t.Fatalf("%+v", d)
	}
}

// Spellings that differ only in case or Unicode normalization name one file on
// the default macOS and Windows file systems, so they must overlap.
func TestFirstOverlapPairFoldsCaseAndNormalization(t *testing.T) {
	overlapping := [][2]string{
		{"pkg/Foo.go", "pkg/foo.go"},
		{"internal/CLI/**", "internal/cli/x.go"},
		{"pkg/caf\u00e9", "pkg/cafe\u0301/x.go"},
		{"PKG/A", "pkg/a/**"},
		// Full case folding, which ToLower lacks: final sigma and sharp s.
		{"pkg/\u03bf\u03b4\u03bf\u03c2", "pkg/\u039f\u0394\u039f\u03a3"},
		{"pkg/stra\u00dfe", "pkg/STRASSE/x.go"},
		{"pkg/\u212a", "pkg/k/x.go"},
	}
	for _, pair := range overlapping {
		units := []TeamUnit{teamWorker("T1", pair[0]), teamWorker("T2", pair[1])}
		if a, b, ok := firstOverlapPair(units); !ok || a != "T1" || b != "T2" {
			t.Fatalf("%q %q: %s %s %v", pair[0], pair[1], a, b, ok)
		}
		if d := DecideTeam(TeamInput{Units: units, Capability: capOK()}); d.Reason != "ownership_overlap" {
			t.Fatalf("%q %q: %+v", pair[0], pair[1], d)
		}
	}
	// A path decoding would refuse cannot be compared, so it overlaps everything.
	if _, _, ok := firstOverlapPair([]TeamUnit{teamWorker("T1", "../escape"), teamWorker("T2", "pkg/b")}); !ok {
		t.Fatal("an uncomparable path must not look disjoint")
	}
	for _, pair := range [][2]string{{"pkg/A", "pkg/ab"}, {"pkg/Foo", "pkg/foobar"}, {"pkg/caf\u00e9", "pkg/cafe"}} {
		units := []TeamUnit{teamWorker("T1", pair[0]), teamWorker("T2", pair[1])}
		if _, _, ok := firstOverlapPair(units); ok {
			t.Fatalf("%q %q must stay disjoint", pair[0], pair[1])
		}
	}
}
