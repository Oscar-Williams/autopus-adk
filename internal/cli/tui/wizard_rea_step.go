package tui

import (
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
)

// reaTargetPlatforms are the platforms that get a project-scope REA MCP entry
// (SPEC-REA-001 MCP Entry Shape Contract). OMP and OpenCode get guidance
// only, so selecting just them never asks the REA question.
var reaTargetPlatforms = []string{"claude-code", "codex", "antigravity-cli"}

// reaStepText is the REA step's Form Text in one locale.
type reaStepText struct {
	Title       string
	Description string
}

// reaStepTextFor returns the REA step text: Korean for ko and English for
// every other locale, as the composition text is.
func reaStepTextFor(lang string) reaStepText {
	if lang == "ko" {
		return reaStepText{
			Title:       "REA 역공학 MCP",
			Description: "Claude Code·Codex·Antigravity에 rea-agents@6.5.0을 프로젝트 범위로 등록합니다 (Node ^22.19.0 || ^24.11.0 || >=26.0.0)",
		}
	}
	return reaStepText{
		Title:       "REA reverse-engineering MCP",
		Description: "Registers rea-agents@6.5.0 for Claude Code, Codex, and Antigravity in this project (Node ^22.19.0 || ^24.11.0 || >=26.0.0)",
	}
}

// reaStepWanted reports whether init asks the REA question: neither a flag
// nor an existing decision preset it, and a REA target platform is selected.
func reaStepWanted(opts InitWizardOpts) bool {
	return opts.Rea == nil && slices.ContainsFunc(opts.Platforms, func(platform string) bool {
		return slices.Contains(reaTargetPlatforms, platform)
	})
}

// reaDefault is the REA answer before the step runs: the preset decision,
// else Yes (REQ-REA-02).
func reaDefault(opts InitWizardOpts) bool {
	if opts.Rea != nil {
		return *opts.Rea
	}
	return true
}

// buildReaStep creates the REA confirmation step in the system locale.
func buildReaStep(num, total int, r *InitWizardResult) *huh.Form {
	return buildReaStepFor(num, total, r, DetectSystemLang())
}

// buildReaStepFor creates the REA confirmation step with the text for lang.
func buildReaStepFor(num, total int, r *InitWizardResult, lang string) *huh.Form {
	text := reaStepTextFor(lang)

	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("[%d/%d] %s", num, total, text.Title)).
				Description(text.Description).
				Affirmative("Yes").
				Negative("No").
				Value(&r.Rea),
		),
	).WithTheme(AutopusTheme()).WithWidth(bannerWidth + 10)
}
