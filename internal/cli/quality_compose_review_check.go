package cli

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/insajin/autopus-adk/pkg/config"
)

// composeReviewCheckTimeout is the budget of each live check request; the
// doctor smoke keeps its own 30-second default.
const composeReviewCheckTimeout = 60 * time.Second

// composeReviewCheckDetailRunes bounds the failure detail of a result line.
const composeReviewCheckDetailRunes = 200

// reviewLiveCheck asks the live check select, whose default makes no call
// (REQ-RVM-10). Only on 확인 호출 does it run runProviderTransportSmoke, the
// doctor smoke unchanged, on the candidate: a copy of the run's config with
// the working value, never the file. It sends one request per provider in
// turn, and an OMP provider retries a transient error within the request,
// so the label names the most model calls. A provider name with terminal
// control characters stops the check before any request.
func (f *composeFormRun) reviewLiveCheck(tab *reviewTabState) error {
	text := f.reviewText()
	candidate, err := reviewCandidate(f.shared.cfg, tab.working)
	if err != nil {
		f.warn(err)
		return nil
	}
	names := resolveSpecReviewProviderNames(candidate, false)
	run, err := composeSelect(f.cmd, text.Check, "", []composeOption[bool]{
		{text.CheckSkip, false},
		{fmt.Sprintf(text.CheckRun, len(names), len(names)*ompReviewMaxAttempts), true},
	}, false)
	if err != nil || !run {
		return err
	}
	for _, name := range names {
		if terminalSafe(name) != name {
			f.reviewLine(fmt.Sprintf(text.CheckRefused, name))
			return nil
		}
	}
	failed := false
	for _, result := range runProviderTransportSmoke(commandContext(f.cmd), candidate, composeReviewCheckTimeout) {
		f.reviewLine(reviewCheckLine(candidate, result, text))
		failed = failed || result.Status != "pass"
	}
	if failed {
		f.reviewLine(text.CheckFailed)
	}
	return nil
}

// reviewCandidate is the live check's config: a detached copy of cfg
// carrying the working value. Nothing writes it.
func reviewCandidate(cfg *config.HarnessConfig, working reviewComposition) (*config.HarnessConfig, error) {
	candidate, err := cloneHarnessConfig(cfg)
	if err != nil {
		return nil, err
	}
	candidate.Spec.ReviewGate.Providers = slices.Clone(working.Providers)
	candidate.Spec.ReviewGate.Judge = working.Judge
	for name, selector := range working.Models {
		entry := candidate.Orchestra.Providers[name]
		entry.Model = selector
		candidate.Orchestra.Providers[name] = entry
	}
	return candidate, nil
}

// reviewCheckLine is one result line (Live Check Contract): the provider,
// its candidate selector or CLI backend, and the verdict. The review_gate
// result of a failed assembly names no model.
func reviewCheckLine(candidate *config.HarnessConfig, result providerSmokeResult, text composeReviewText) string {
	subject := result.Provider
	if entry, ok := candidate.Orchestra.Providers[result.Provider]; ok || result.Provider != "review_gate" {
		model := text.CLIBackend
		if entry.Backend == config.ProviderBackendOMP {
			model = entry.Model
		}
		subject += " (" + model + ")"
	}
	verdict := text.CheckPass
	switch result.Status {
	case "pass":
	case "warn":
		verdict = text.CheckWarn
	default:
		verdict = fmt.Sprintf(text.CheckFail, reviewCheckDetail(result.Detail))
	}
	return fmt.Sprintf(text.CheckLine, subject, verdict)
}

// reviewCheckDetail is the first line of a failure detail, terminal-safe and
// cut to composeReviewCheckDetailRunes runes.
func reviewCheckDetail(detail string) string {
	line, _, _ := strings.Cut(detail, "\n")
	runes := []rune(terminalSafe(line))
	if len(runes) > composeReviewCheckDetailRunes {
		runes = runes[:composeReviewCheckDetailRunes]
	}
	return string(runes)
}
