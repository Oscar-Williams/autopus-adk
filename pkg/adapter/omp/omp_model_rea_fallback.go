package omp

import (
	"fmt"
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

// dropOMPReaAnalysisFallbackRoutes removes reverse-engineer's route when the
// catalog cannot serve the built-in balanced analysis rung
// (config.OMPReaAnalysisFallback, SPEC-REA-001 Gate 1). Without a route the
// agent gets no task.agentModelOverrides key and no projection entry, so OMP
// runs it on the session model and its definition keeps its own thinking
// level. Every other route is left for the existing fail-closed checks. The
// selectors given up are returned for the notice.
func dropOMPReaAnalysisFallbackRoutes(
	profile config.RoleModelProfileConf,
	catalog OMPModelCatalog,
	routes map[string]OMPModelRouteRequest,
) []string {
	request, ok := routes[config.ReverseEngineerAgent]
	if !ok || len(request.Candidates) != 1 {
		return nil
	}
	_, reason := matchOMPModelCandidate(catalog.Models, request.Capability, request.Candidates[0])
	if !config.OMPReaAnalysisFallback(profile, config.ReverseEngineerAgent, reason) {
		return nil
	}
	delete(routes, config.ReverseEngineerAgent)
	return []string{request.Candidates[0].Selector}
}

// ompReaFallbackNotice is the Form Text fallback line for each selector the
// analysis rung gave up, or "" when none was.
func ompReaFallbackNotice(selectors []string) string {
	lines := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		lines = append(lines, fmt.Sprintf("  [omp] %s: %s not in catalog → session model",
			config.ReverseEngineerAgent, selector))
	}
	return strings.Join(lines, "\n")
}

// joinOMPNotices joins the non-empty notices one per line.
func joinOMPNotices(notices ...string) string {
	kept := make([]string, 0, len(notices))
	for _, notice := range notices {
		if notice != "" {
			kept = append(kept, notice)
		}
	}
	return strings.Join(kept, "\n")
}
