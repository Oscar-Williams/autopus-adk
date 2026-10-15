package taskroute

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Routes and reason codes of the Auto-Team Decision Contract.
const (
	routeTeam     = "team"
	routeSubagent = "subagent"

	reasonExplicitFlag              = "explicit_flag"
	reasonConfigOff                 = "config_off"
	reasonFewerThanTwoReady         = "fewer_than_two_ready"
	reasonOwnershipOverlap          = "ownership_overlap"
	reasonNotClaudeCode             = "not_claude_code"
	reasonTeamsEnvMissing           = "teams_env_missing"
	reasonNotInteractive            = "not_interactive"
	reasonNotTopLevel               = "not_top_level"
	reasonClaudeVersionUnknown      = "claude_version_unknown"
	reasonClaudeVersionOld          = "claude_version_old"
	reasonIndependentDisjointSlices = "independent_disjoint_slices"
)

// minimumTeamVersion is the oldest Claude Code release Route B accepts.
var minimumTeamVersion = []int{2, 1, 246}

var claudeVersionLine = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+) \(Claude Code\)$`)

// TopologyFlags are the explicit /auto go topology flags a caller declared.
type TopologyFlags struct {
	Team     bool `json:"team"`
	Workflow bool `json:"workflow"`
	Solo     bool `json:"solo"`
}

// TeamUnit is one row of the Phase 1 assignment table.
type TeamUnit struct {
	ID         string   `json:"id"`
	Execution  string   `json:"execution"`
	OwnedPaths []string `json:"owned_paths"`
	DependsOn  []string `json:"depends_on"`
	Done       bool     `json:"done"`
}

// Capability is the session snapshot the CLI observed. The decision only reads
// it; reading the environment and running the version probe are the caller's.
type Capability struct {
	ClaudeCode      string // CLAUDECODE
	AgentTeams      string // CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS
	Entrypoint      string // CLAUDE_CODE_ENTRYPOINT
	SessionAttended string // CLAUDE_CODE_SESSION_ATTENDED
	ChildSession    string // CLAUDE_CODE_CHILD_SESSION
	ExecPathBase    string // basename of CLAUDE_CODE_EXECPATH, empty when unset
	Version         string // first probe output line, empty when the probe did not run or failed
}

// TeamInput is the one input value of the Auto-Team Decision.
type TeamInput struct {
	Flags      TopologyFlags
	AutoTeam   *bool // workflow.auto_team, nil when the key is absent
	Units      []TeamUnit
	Capability Capability
	Locale     string // language.ai_responses
}

// TeamDecision is the Auto-Team Decision. ReadyUnits is absent for an explicit
// flag. TeamUnits, FollowupUnits, and ToolsFallbackLine belong to an automatic
// team only, and its FollowupUnits is present even when the list is empty.
type TeamDecision struct {
	Version           int       `json:"version"`
	Route             string    `json:"route"`
	Reason            string    `json:"reason"`
	ReadyUnits        *int      `json:"ready_units,omitempty"`
	TeamUnits         []string  `json:"team_units,omitempty"`
	FollowupUnits     *[]string `json:"followup_units,omitempty"`
	Line              string    `json:"line"`
	ToolsFallbackLine string    `json:"tools_fallback_line,omitempty"`
}

// NeedsVersionProbe reports whether rules 1-8 of the Auto-Team Decision all
// pass. Only then may the caller run the version probe and record its output
// in the snapshot, so a run that cannot use a team never executes anything.
func NeedsVersionProbe(in TeamInput) bool {
	_, decided := decideBeforeVersion(in, readyUnits(in.Units))
	return !decided
}

// DecideTeam applies the Auto-Team Decision rules in order and returns the
// first match, so identical inputs always yield an identical decision.
func DecideTeam(in TeamInput) TeamDecision {
	ready := readyUnits(in.Units)
	if d, decided := decideBeforeVersion(in, ready); decided {
		return d
	}
	n := len(ready)
	version, ok := parseClaudeVersion(in.Capability.Version)
	if !ok {
		return TeamDecision{Version: 1, Route: routeSubagent, Reason: reasonClaudeVersionUnknown, ReadyUnits: &n}
	}
	if slices.Compare(version, minimumTeamVersion) < 0 {
		return TeamDecision{Version: 1, Route: routeSubagent, Reason: reasonClaudeVersionOld, ReadyUnits: &n}
	}
	team := make([]string, 0, n)
	for _, unit := range ready {
		team = append(team, unit.ID)
	}
	followup := []string{}
	for _, unit := range in.Units {
		if !unit.Done && !slices.Contains(team, unit.ID) {
			followup = append(followup, unit.ID)
		}
	}
	return TeamDecision{Version: 1, Route: routeTeam, Reason: reasonIndependentDisjointSlices, ReadyUnits: &n, TeamUnits: team, FollowupUnits: &followup}
}

// decideBeforeVersion applies rules 1-8: flags, then config, then the plan,
// then the environment, so a run the plan already keeps on subagents is not
// told to set experimental variables.
func decideBeforeVersion(in TeamInput, ready []TeamUnit) (TeamDecision, bool) {
	if route, _, ok := explicitRoute(in.Flags); ok {
		return TeamDecision{Version: 1, Route: route, Reason: reasonExplicitFlag}, true
	}
	n := len(ready)
	stay := func(reason string) (TeamDecision, bool) {
		return TeamDecision{Version: 1, Route: routeSubagent, Reason: reason, ReadyUnits: &n}, true
	}
	if in.AutoTeam != nil && !*in.AutoTeam {
		return stay(reasonConfigOff)
	}
	if n < 2 {
		return stay(reasonFewerThanTwoReady)
	}
	if _, _, overlap := firstOverlapPair(ready); overlap {
		return stay(reasonOwnershipOverlap)
	}
	c := in.Capability
	switch {
	case c.ClaudeCode != "1":
		return stay(reasonNotClaudeCode)
	case c.AgentTeams != "1":
		return stay(reasonTeamsEnvMissing)
	case c.Entrypoint != "cli" || c.SessionAttended != "1":
		return stay(reasonNotInteractive)
	case c.ChildSession != "":
		return stay(reasonNotTopLevel)
	}
	return TeamDecision{}, false
}

// explicitRoute maps the explicit topology flag to its route. DecodeTeamFacts
// rejects conflicting flags; for a direct conflicting input the declaration
// order decides, so the function stays total and never auto-selects a team.
func explicitRoute(flags TopologyFlags) (route, flag string, ok bool) {
	switch {
	case flags.Team:
		return "team", "--team", true
	case flags.Workflow:
		return "workflow", "--workflow", true
	case flags.Solo:
		return "solo", "--solo", true
	}
	return "", "", false
}

// readyUnits returns, in facts order, the worker units that are not done and
// whose every dependency is done.
func readyUnits(units []TeamUnit) []TeamUnit {
	finished := map[string]bool{}
	for _, unit := range units {
		if unit.Done {
			finished[unit.ID] = true
		}
	}
	var ready []TeamUnit
	for _, unit := range units {
		if unit.Execution != "worker" || unit.Done {
			continue
		}
		waiting := false
		for _, id := range unit.DependsOn {
			if !finished[id] {
				waiting = true
				break
			}
		}
		if !waiting {
			ready = append(ready, unit)
		}
	}
	return ready
}

// firstOverlapPair returns the first pair of units, in facts order, whose
// owned paths overlap. Each pair is judged by ownershipOverlap itself, so the
// predicate stays the one Decide uses; only the pair identity is new.
func firstOverlapPair(units []TeamUnit) (string, string, bool) {
	for i, a := range units {
		for _, b := range units[i+1:] {
			if ownershipOverlap([]Worker{{OwnedPaths: ownedRoots(a.OwnedPaths)}, {OwnedPaths: ownedRoots(b.OwnedPaths)}}) {
				return a.ID, b.ID, true
			}
		}
	}
	return "", "", false
}

// ownedRoots drops the one trailing "/**" an owned path may carry, so
// "pkg/a/**" and "pkg/a" name the same root before cleanPath sees them.
func ownedRoots(paths []string) []string {
	roots := make([]string, len(paths))
	for i, value := range paths {
		roots[i] = strings.TrimSuffix(value, "/**")
	}
	return roots
}

// parseClaudeVersion reads "X.Y.Z (Claude Code)" into integers so 2.10.0
// compares above 2.1.246.
func parseClaudeVersion(line string) ([]int, bool) {
	match := claudeVersionLine.FindStringSubmatch(line)
	if match == nil {
		return nil, false
	}
	version := make([]int, 3)
	for i := range version {
		value, err := strconv.Atoi(match[i+1])
		if err != nil {
			return nil, false
		}
		version[i] = value
	}
	return version, true
}
