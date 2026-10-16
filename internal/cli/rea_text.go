package cli

import "github.com/insajin/autopus-adk/pkg/reamcp"

// reaPackage is the rea-agents version every REA line names.
const reaPackage = "rea-agents@" + reamcp.Pin

// reaText is the SPEC-REA-001 Form Text of the init, update, and doctor REA
// lines in one locale. The stored decision never depends on it (REQ-REA-18).
type reaText struct {
	Question      string // the update question, answered on the next line
	Registered    string // %s is the comma-separated list of target files
	Hint          string
	NodeOutside   string // %s is the judged node version
	NodeAbsent    string
	NodeUnknown   string
	OldBinary     string
	Guidance      string
	FlagBoth      string
	FlagWorkspace string
	DoctorOnEntry string
	DoctorOnKey   string
	DoctorOffKey  string
	DoctorUnset   string
}

// reaTextFor returns the REA lines for lang: Korean for ko and English for
// every other locale, as the composition text is.
func reaTextFor(lang string) reaText {
	if lang == "ko" {
		return reaTextKo
	}
	return reaTextEn
}

// reaCurrentText returns the REA lines in the locale the composeLanguage seam
// reports.
func reaCurrentText() reaText {
	return reaTextFor(composeLanguage())
}

var reaTextKo = reaText{
	Question:      "REA 역공학 MCP(" + reaPackage + ")를 이 프로젝트의 Claude Code·Codex·Antigravity에 등록할까요? [Y/n]",
	Registered:    "REA 역공학 MCP를 등록했습니다: %s (" + reaPackage + ")",
	Hint:          "REA 역공학 MCP를 아직 정하지 않았습니다. 켜려면 auto update --rea, 끄려면 auto update --no-rea",
	NodeOutside:   "REA MCP를 등록하지 않았습니다: node %s은 " + reamcp.NodeRange + " 범위가 아닙니다",
	NodeAbsent:    "REA MCP를 등록하지 않았습니다: node를 찾지 못했습니다 (" + reamcp.NodeRange + " 필요)",
	NodeUnknown:   "REA MCP를 등록하지 않았습니다: node 버전을 확인하지 못했습니다",
	OldBinary:     "features.rea는 v0.50.126보다 오래된 auto가 읽지 못합니다",
	Guidance:      "OMP·OpenCode에는 Autopus가 REA MCP를 등록하지 않습니다. /auto reverse 분석은 Claude Code·Codex·Antigravity에서 실행하세요. npx -y " + reaPackage + " setup은 REA를 사용자 범위의 메인 세션에만 등록합니다",
	FlagBoth:      "--rea와 --no-rea는 함께 쓸 수 없습니다",
	FlagWorkspace: "--rea·--no-rea는 --workspace와 함께 쓸 수 없습니다. repo마다 --local로 실행하세요",
	DoctorOnEntry: "REA: 켜짐 (managed 항목)",
	DoctorOnKey:   "REA: 켜짐 (autopus.yaml)",
	DoctorOffKey:  "REA: 꺼짐 (autopus.yaml)",
	DoctorUnset:   "REA: 미결정",
}

var reaTextEn = reaText{
	Question:      "Register the REA reverse-engineering MCP (" + reaPackage + ") for Claude Code, Codex, and Antigravity in this project? [Y/n]",
	Registered:    "Registered the REA reverse-engineering MCP: %s (" + reaPackage + ")",
	Hint:          "REA reverse-engineering MCP is undecided; run auto update --rea to enable or auto update --no-rea to decline",
	NodeOutside:   "REA MCP not registered: node %s is outside " + reamcp.NodeRange,
	NodeAbsent:    "REA MCP not registered: node not found (needs " + reamcp.NodeRange + ")",
	NodeUnknown:   "REA MCP not registered: node version unreadable",
	OldBinary:     "auto older than v0.50.126 cannot read features.rea",
	Guidance:      "Autopus does not register REA MCP for OMP or OpenCode; run /auto reverse analysis in Claude Code, Codex, or Antigravity. npx -y " + reaPackage + " setup registers REA for the main session only (user scope)",
	FlagBoth:      "--rea and --no-rea cannot be used together",
	FlagWorkspace: "--rea and --no-rea cannot be combined with --workspace; run them per repository with --local",
	DoctorOnEntry: "REA: on (managed entry)",
	DoctorOnKey:   "REA: on (autopus.yaml)",
	DoctorOffKey:  "REA: off (autopus.yaml)",
	DoctorUnset:   "REA: undecided",
}
