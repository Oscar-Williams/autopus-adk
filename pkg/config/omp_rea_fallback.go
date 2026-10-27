package config

// ReverseEngineerAgent is the canonical name of the REA analysis agent
// (SPEC-REA-001).
const ReverseEngineerAgent = "reverse-engineer"

// OMPReaAnalysisFallback is the single decision behind the Gate 1 analysis
// rung fallback (SPEC-REA-001 Preset Contract, OMP balanced row): an OMP route
// that the catalog cannot serve leaves the agent on the session model instead
// of blocking only for reverse-engineer on the built-in balanced analysis
// rung, and only when the catalog lacks the rung's model (catalogReason
// model_unknown) or that model's thinking level (thinking_unsupported). Every
// other rung, an operator pin, a custom or ultra profile, and every other
// catalog verdict keep failing closed. catalogReason is the candidate verdict
// the OMP router reports for the route's single candidate.
func OMPReaAnalysisFallback(profile RoleModelProfileConf, agent, catalogReason string) bool {
	if agent != ReverseEngineerAgent || !profile.Builtin || profile.OMPAgentOperatorPinned(agent) {
		return false
	}
	if catalogReason != "model_unknown" && catalogReason != "thinking_unsupported" {
		return false
	}
	candidates, err := profile.AgentCandidates(agent)
	if err != nil || len(candidates) != 1 {
		return false
	}
	for _, rungs := range balancedRungCandidates {
		if candidates[0] == rungs[balancedRungAnalysis] {
			return true
		}
	}
	return false
}
