package pipeline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/insajin/autopus-adk/pkg/learn"
)

var (
	// coverageRe extracts coverage percentage from go test output.
	coverageRe = regexp.MustCompile(`coverage: (\d+\.?\d*)% of statements`)
	// findingRe extracts individual review findings.
	findingRe = regexp.MustCompile(`FINDING: \[(\w+)\] (.+)`)
	// findingFileRe extracts `path/to/file.ext:line` references from a finding.
	findingFileRe = regexp.MustCompile(`([A-Za-z0-9_.\-]+(?:/[A-Za-z0-9_.\-]+)*\.[A-Za-z0-9]+):\d+`)
)

// Every hook records the run's SPEC ID, so `auto learn query --spec` (the
// /auto go learning injection) finds pipeline-recorded entries too.

// learnHookGateFail records a gate failure event. Silently returns when store is nil.
func learnHookGateFail(store *learn.Store, specID string, phaseID PhaseID, gate GateType, output string, attempt int) {
	if store == nil {
		return
	}
	severity := learn.SeverityMedium
	if attempt >= defaultMaxRetries {
		severity = learn.SeverityCritical
	}
	pattern := fmt.Sprintf("gate %s failed at phase %s (attempt %d)", gate, phaseID, attempt)
	_ = learn.RecordGateFail(store, learn.RecordOpts{
		Phase:    string(phaseID),
		SpecID:   specID,
		Pattern:  pattern,
		Severity: severity,
	})
}

// learnHookCoverageGap records a coverage gap when coverage is below threshold.
// Silently returns when store is nil.
func learnHookCoverageGap(store *learn.Store, specID string, output string, threshold float64) {
	if store == nil {
		return
	}
	coverage := parseCoverage(output)
	if coverage < 0 || coverage >= threshold {
		return
	}
	pattern := fmt.Sprintf("coverage %.1f%% below threshold %.1f%%", coverage, threshold)
	_ = learn.RecordCoverageGap(store, learn.RecordOpts{
		SpecID:   specID,
		Pattern:  pattern,
		Severity: learn.SeverityMedium,
	})
}

// learnHookReviewIssue records individual review findings from gate output.
// Silently returns when store is nil.
func learnHookReviewIssue(store *learn.Store, output string, specID string) {
	if store == nil {
		return
	}
	matches := findingRe.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		// Graceful degradation: record entire output as single issue.
		_ = learn.RecordReviewIssue(store, learn.RecordOpts{
			SpecID:   specID,
			Pattern:  output,
			Severity: learn.SeverityMedium,
		})
		return
	}
	for _, m := range matches {
		sev := mapFindingSeverity(m[1])
		_ = learn.RecordReviewIssue(store, learn.RecordOpts{
			SpecID:   specID,
			Files:    findingFiles(m[2]),
			Pattern:  m[2],
			Severity: sev,
		})
	}
}

// learnHookExecutorError records an executor error. Silently returns when store is nil.
func learnHookExecutorError(store *learn.Store, specID string, phaseID PhaseID, err error) {
	if store == nil {
		return
	}
	_ = learn.RecordExecutorError(store, learn.RecordOpts{
		Phase:    string(phaseID),
		SpecID:   specID,
		Pattern:  err.Error(),
		Severity: learn.SeverityHigh,
	})
}

// findingFiles returns the distinct file paths a finding cites as `path:line`,
// so `auto learn query --files` can match review issues.
func findingFiles(finding string) []string {
	var files []string
	seen := map[string]bool{}
	for _, m := range findingFileRe.FindAllStringSubmatch(finding, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			files = append(files, m[1])
		}
	}
	return files
}

// parseCoverage extracts coverage percentage from output. Returns -1 if not found.
func parseCoverage(output string) float64 {
	m := coverageRe.FindStringSubmatch(output)
	if len(m) < 2 {
		return -1
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return -1
	}
	return v
}

// mapFindingSeverity maps a finding severity string to learn.Severity.
func mapFindingSeverity(s string) learn.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return learn.SeverityCritical
	case "high":
		return learn.SeverityHigh
	case "medium":
		return learn.SeverityMedium
	default:
		return learn.SeverityLow
	}
}
