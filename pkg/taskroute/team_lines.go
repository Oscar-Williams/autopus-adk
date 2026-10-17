package taskroute

import (
	"regexp"
	"strconv"
	"strings"
)

// reasonTeamToolsUnavailable names tools_fallback_line, the line the lead
// prints instead of line when the team tool schemas fail to load.
const reasonTeamToolsUnavailable = "team_tools_unavailable"

// reasonInvalidFacts is the reason of a decision for rejected facts.
const reasonInvalidFacts = "invalid_facts"

var (
	sessionTokenPattern  = regexp.MustCompile(`^[a-z0-9-]{1,32}$`)
	execPathTokenPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)
)

// lineValues holds the placeholders of a reason line. entrypoint, attended,
// and execBase are raw observations; renderTeamLine normalizes them so no
// unvetted environment text reaches the line.
type lineValues struct {
	n          int
	a, b       string
	entrypoint string
	attended   string
	execBase   string
	version    string
	flag       string
	code       string
}

// teamLineText holds the Korean and English line of every reason, verbatim
// from the Reason Line Contract.
var teamLineText = map[string][2]string{
	reasonIndependentDisjointSlices: {
		"Agent Team: 전환 — 준비된 slice {n}개의 owned_paths가 겹치지 않아 native Agent Team으로 실행한다",
		"Agent Team: on — {n} ready slices have disjoint owned_paths; running them as a native Agent Team"},
	reasonConfigOff: {
		"Agent Team: 유지 — autopus.yaml의 workflow.auto_team이 false라 subagent로 실행한다",
		"Agent Team: off — workflow.auto_team is false in autopus.yaml; running subagents"},
	reasonFewerThanTwoReady: {
		"Agent Team: 유지 — 함께 실행할 준비된 slice가 {n}개라 subagent로 실행한다",
		"Agent Team: off — {n} ready slice(s) can run together; running subagents"},
	reasonOwnershipOverlap: {
		"Agent Team: 유지 — {a}, {b}의 owned_paths가 겹쳐 subagent로 실행한다",
		"Agent Team: off — {a} and {b} have overlapping owned_paths; running subagents"},
	reasonNotClaudeCode: {
		"Agent Team: 유지 — Claude Code 세션이 아니라 subagent로 실행한다",
		"Agent Team: off — not a Claude Code session; running subagents"},
	reasonTeamsEnvMissing: {
		"Agent Team: 유지 — 세션 환경에 CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1이 없어 subagent로 실행한다",
		"Agent Team: off — CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is not in the session environment; running subagents"},
	reasonNotInteractive: {
		"Agent Team: 유지 — 대화형 CLI 세션이 아니라(entrypoint={e}, attended={t}) subagent로 실행한다",
		"Agent Team: off — not an interactive CLI session (entrypoint={e}, attended={t}); running subagents"},
	reasonClaudeVersionUnknown: {
		"Agent Team: 유지 — 실행 중인 Claude Code 버전을 읽지 못해(execpath={x}) subagent로 실행한다",
		"Agent Team: off — the running Claude Code version could not be read (execpath={x}); running subagents"},
	reasonClaudeVersionOld: {
		"Agent Team: 유지 — Claude Code {v} 버전은 2.1.246보다 낮아 subagent로 실행한다",
		"Agent Team: off — Claude Code {v} is older than 2.1.246; running subagents"},
	reasonTeamToolsUnavailable: {
		"Agent Team: 유지 — SendMessage·task 도구 schema를 불러오지 못해 subagent로 실행한다",
		"Agent Team: off — SendMessage/task tool schemas could not be loaded; running subagents"},
	reasonExplicitFlag: {
		"Agent Team: 자동 선택 안 함 — {f} 지정됨",
		"Agent Team: auto-selection skipped — {f} was given"},
	reasonInvalidFacts: {
		"Agent Team: 유지 — topology facts가 잘못되어({c}) subagent로 실행한다",
		"Agent Team: off — the topology facts are invalid ({c}); running subagents"},
}

// renderTeamLine renders the line of reason, in Korean only when locale is
// exactly "ko" and in English for every other value, including none.
func renderTeamLine(locale, reason string, values lineValues) string {
	text := teamLineText[reason]
	line := text[1]
	if locale == "ko" {
		line = text[0]
	}
	return strings.NewReplacer(
		"{n}", strconv.Itoa(values.n),
		"{a}", values.a,
		"{b}", values.b,
		"{e}", sessionToken(values.entrypoint),
		"{t}", sessionToken(values.attended),
		"{x}", execPathToken(values.execBase),
		"{v}", values.version,
		"{f}", values.flag,
		"{c}", values.code,
	).Replace(line)
}

// sessionToken passes an entrypoint or attended value through only when it is
// a short lowercase token, so the line never carries arbitrary session text.
func sessionToken(value string) string {
	if value == "" {
		return "unset"
	}
	if sessionTokenPattern.MatchString(value) {
		return value
	}
	return "unknown"
}

// execPathToken passes the CLAUDE_CODE_EXECPATH basename through only when it
// is a short file-name token; the full path never reaches the line.
func execPathToken(base string) string {
	if base == "" {
		return "unset"
	}
	if execPathTokenPattern.MatchString(base) {
		return base
	}
	return "unknown"
}
