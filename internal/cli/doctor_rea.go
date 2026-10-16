package cli

import (
	"context"
	"io"
	"strings"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// doctorReaPayload is data.rea of `auto doctor --json` (SPEC-REA-001 Doctor
// Contract). Platforms holds only the selected REA target platforms, and
// Guidance is empty, never null, when the decision is off.
type doctorReaPayload struct {
	Decision       string            `json:"decision"`
	DecisionSource string            `json:"decision_source"`
	Node           doctorReaNode     `json:"node"`
	Platforms      map[string]string `json:"platforms"`
	Guidance       []string          `json:"guidance"`
}

// doctorReaNode is the Node gate verdict; Version is set for pass and
// unsupported only.
type doctorReaNode struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

// reaDoctorCheck is one doctor.rea.* row before rendering.
type reaDoctorCheck struct {
	id      string
	detail  string
	warning bool
}

// diagnoseRea reads the REA decision and its source, runs the Node probe
// once, and classifies the target file of each selected REA target platform.
// It writes nothing, starts no REA server, and makes no network call.
func diagnoseRea(ctx context.Context, dir string, cfg *config.HarnessConfig) doctorReaPayload {
	decision := resolveReaDecision(dir, cfg)
	gate := newReaNodeProbe().Gate(ctx)
	payload := doctorReaPayload{
		Decision:       string(decision.State),
		DecisionSource: string(decision.Source),
		Node:           doctorReaNode{Status: string(gate.Status), Version: gate.Version},
		Platforms:      map[string]string{},
		Guidance:       []string{},
	}
	for _, p := range reaSelectedTargets(cfg.Platforms) {
		payload.Platforms[string(p)] = string(reaFileClass(dir, p))
	}
	if decision.State != reaOff {
		payload.Guidance = reaGuidancePlatforms(cfg.Platforms)
	}
	return payload
}

// reaDoctorChecks returns the Doctor Contract rows in order: the decision,
// the Node gate, each selected platform, and each guided platform. The Node
// row warns when the decision is on and the gate does not pass; a platform
// row warns when the decision is on, the gate passes, and its entry is
// absent.
func reaDoctorChecks(payload doctorReaPayload) []reaDoctorCheck {
	on := payload.Decision == string(reaOn)
	pass := payload.Node.Status == string(reamcp.NodePass)
	decision := "undecided"
	if payload.Decision != string(reaUndecided) {
		decision = payload.Decision + " (" + payload.DecisionSource + ")"
	}
	checks := []reaDoctorCheck{
		{id: "doctor.rea.decision", detail: decision},
		{id: "doctor.rea.node", detail: strings.TrimSpace(payload.Node.Status + " " + payload.Node.Version), warning: on && !pass},
	}
	for _, p := range reamcp.Platforms() {
		if class, selected := payload.Platforms[string(p)]; selected {
			checks = append(checks, reaDoctorCheck{
				id: "doctor.rea.platform." + string(p), detail: class,
				warning: on && pass && class == string(reamcp.ClassAbsent),
			})
		}
	}
	for _, p := range payload.Guidance {
		checks = append(checks, reaDoctorCheck{id: "doctor.rea.guidance." + p, detail: "npx -y " + reaPackage + " setup"})
	}
	return checks
}

// collectReaChecks adds data.rea and the doctor.rea.* checks. A warning row
// marks the report warn, as the edit-guard rows do.
func (r *doctorJSONReport) collectReaChecks(ctx context.Context, dir string, cfg *config.HarnessConfig) {
	payload := diagnoseRea(ctx, dir, cfg)
	r.data.Rea = &payload
	for _, check := range reaDoctorChecks(payload) {
		severity, status := "info", "pass"
		if check.warning {
			severity, status = "warning", "warn"
			r.status = jsonStatusWarn
			r.warnings = append(r.warnings, jsonMessage{Code: "rea_not_registered", Message: check.id + ": " + check.detail})
		}
		r.checks = append(r.checks, jsonCheck{ID: check.id, Severity: severity, Status: status, Detail: check.detail})
	}
}

// checkReaText prints the REA section of `auto doctor`. It is advisory and
// never fails the run.
func checkReaText(ctx context.Context, out io.Writer, dir string, cfg *config.HarnessConfig) {
	payload := diagnoseRea(ctx, dir, cfg)
	text := reaCurrentText()
	tui.SectionHeader(out, "REA")
	tui.Info(out, reaDoctorDecisionLine(text, payload))
	for _, check := range reaDoctorChecks(payload)[1:] {
		label := strings.TrimPrefix(strings.TrimPrefix(check.id, "doctor.rea."), "platform.")
		switch {
		case strings.HasPrefix(check.id, "doctor.rea.guidance."):
		case check.warning:
			tui.SKIP(out, label+": "+check.detail)
		default:
			tui.OK(out, label+": "+check.detail)
		}
	}
	if len(payload.Guidance) > 0 {
		tui.Info(out, text.Guidance)
	}
}

// reaDoctorDecisionLine is the Form Text decision line for payload.
func reaDoctorDecisionLine(text reaText, payload doctorReaPayload) string {
	switch {
	case payload.Decision == string(reaOff):
		return text.DoctorOffKey
	case payload.Decision != string(reaOn):
		return text.DoctorUnset
	case payload.DecisionSource == string(reaSourceManaged):
		return text.DoctorOnEntry
	}
	return text.DoctorOnKey
}
