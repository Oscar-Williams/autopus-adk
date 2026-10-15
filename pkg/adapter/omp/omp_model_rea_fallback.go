package omp

import (
	"github.com/insajin/autopus-adk/pkg/adapter"
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

// ompReaFallbackFacts is the structured fallback fact of each selector the
// analysis rung gave up; the CLI prints them in the run's language.
func ompReaFallbackFacts(selectors []string) []adapter.ReaFallback {
	facts := make([]adapter.ReaFallback, 0, len(selectors))
	for _, selector := range selectors {
		facts = append(facts, adapter.ReaFallback{
			Catalog: "OMP", Agent: config.ReverseEngineerAgent, Requested: selector,
		})
	}
	return facts
}

// OMPReaSessionModelStatus is the routing status of the route generation
// drops for the analysis-rung fallback, in a compilation that keeps it.
const OMPReaSessionModelStatus = "session_model"

// MarkOMPReaAnalysisFallbacks rewrites, in a compilation that kept every
// route, each unresolved route config.OMPReaAnalysisFallback lets generation
// drop, so doctor and plan rows report the session-model fallback instead of
// a blocked route. Every other resolution is left as compiled.
func MarkOMPReaAnalysisFallbacks(profile config.RoleModelProfileConf, compilation *OMPModelRoutingCompilation) {
	for index := range compilation.Resolutions {
		resolution := &compilation.Resolutions[index]
		if resolution.Status == "selected" || len(resolution.FallbackAttempts) != 1 {
			continue
		}
		if config.OMPReaAnalysisFallback(profile, resolution.Agent, resolution.FallbackAttempts[0].Reason) {
			resolution.Status = OMPReaSessionModelStatus
			resolution.Reason = "rea_analysis_fallback"
		}
	}
}
