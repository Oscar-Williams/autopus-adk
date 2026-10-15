package cli

import (
	"fmt"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/config"
)

// ompProfileStatusSessionModel marks a plan row that runs on the OMP session
// model because the catalog cannot serve the built-in balanced analysis rung
// (SPEC-REA-001 Gate 1). Such a row carries no effective selector and no
// blocker; generation drops the same route.
const ompProfileStatusSessionModel = "session_model"

// applyOMPReaFallbackToRow turns an unavailable row into a session-model
// fallback when config.OMPReaAnalysisFallback allows it for the route's single
// candidate verdict, and reports whether it did. Every other unavailable row
// stays a blocker.
func applyOMPReaFallbackToRow(
	profile config.RoleModelProfileConf,
	row *ompProfileAgentPreviewPayload,
	resolution omp.OMPModelRouteResolution,
) bool {
	if row.Availability != ompProfileAvailabilityUnavailable || len(resolution.FallbackAttempts) != 1 {
		return false
	}
	if !config.OMPReaAnalysisFallback(profile, row.Agent, resolution.FallbackAttempts[0].Reason) {
		return false
	}
	row.Status = ompProfileStatusSessionModel
	return true
}

// ompReaFallbackWarnings is the Form Text fallback line of every row that runs
// on the session model, in row order. A plan another route blocks writes
// nothing, so it prints no fallback line.
func ompReaFallbackWarnings(rows []ompProfileAgentPreviewPayload, blockers []string) []string {
	if len(blockers) > 0 {
		return nil
	}
	var warnings []string
	for _, row := range rows {
		if row.Status == ompProfileStatusSessionModel {
			warnings = append(warnings, fmt.Sprintf("%s: %s not in catalog → session model",
				row.Agent, row.RequestedSelector))
		}
	}
	return warnings
}
