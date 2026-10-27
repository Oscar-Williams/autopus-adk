package cli

import (
	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/adapter/omp"
)

// ompProfileStatusSessionModel marks a plan row that runs on the OMP session
// model because the catalog cannot serve the built-in balanced analysis rung
// (SPEC-REA-001 Gate 1). Such a row carries no effective selector and no
// blocker; generation drops the same route.
const ompProfileStatusSessionModel = "session_model"

// applyOMPReaFallbackToRow turns the row of a route the doctor compilation
// marked as the analysis-rung fallback (omp.MarkOMPReaAnalysisFallbacks, from
// config.OMPReaAnalysisFallback) into a session-model row, and reports whether
// it did. Every other unavailable row stays a blocker.
func applyOMPReaFallbackToRow(row *ompProfileAgentPreviewPayload, resolution omp.OMPModelRouteResolution) bool {
	if row.Availability != ompProfileAvailabilityUnavailable || resolution.Status != omp.OMPReaSessionModelStatus {
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
			warnings = append(warnings, reaFallbackLine(adapter.ReaFallback{
				Catalog: "OMP", Agent: row.Agent, Requested: row.RequestedSelector,
			}, composeLanguage()))
		}
	}
	return warnings
}
