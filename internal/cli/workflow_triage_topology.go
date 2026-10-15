package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/taskroute"
)

// workflowTriageTopologyReport is the --topology output. The units come from
// the caller and the capability from this process's environment; nothing was
// launched, written, or gated.
type workflowTriageTopologyReport struct {
	Mode               string                 `json:"mode"`
	UnitSource         string                 `json:"unit_source"`
	CapabilitySource   string                 `json:"capability_source"`
	ConfigSource       string                 `json:"config_source"`
	ExecutionPerformed bool                   `json:"execution_performed"`
	Decision           taskroute.TeamDecision `json:"decision"`
}

// runWorkflowTriageTopology decides whether Phase 2 of a Claude Code /auto go
// run uses a native Agent Team. It reads the facts file and autopus.yaml, the
// session environment, and, only when rules 1-8 pass, the version of the
// running Claude Code binary. That probe is its only child process.
func runWorkflowTriageTopology(out io.Writer, factsPath, format string) error {
	data, err := readWorkflowTriageFile(factsPath)
	if err != nil {
		return err
	}
	if len(data) > workflowTriageMaxBytes {
		return fmt.Errorf("facts file changed or exceeds limit")
	}
	facts, err := taskroute.DecodeTeamFacts(data)
	if err != nil {
		return err
	}
	autoTeam, locale, source, err := readTopologyConfig()
	if err != nil {
		return err
	}
	snapshot, execPath := observeTeamCapability()
	input := taskroute.TeamInput{Flags: facts.Flags, AutoTeam: autoTeam, Units: facts.Units, Capability: snapshot, Locale: locale}
	if taskroute.NeedsVersionProbe(input) {
		input.Capability.Version = probeClaudeVersion(execPath)
	}
	decision := taskroute.DecideTeam(input)
	if format == "human" {
		_, err := fmt.Fprintln(out, decision.Line)
		return err
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(workflowTriageTopologyReport{
		Mode: "topology", UnitSource: "caller_declared", CapabilitySource: "observed_process_env",
		ConfigSource: source, ExecutionPerformed: false, Decision: decision,
	})
}

// readTopologyConfig returns workflow.auto_team, language.ai_responses, and the
// config source of the working directory. A missing autopus.yaml is the
// default (enabled, English) and is not loaded, because LoadPreview would
// synthesize one. An existing file is read with LoadPreview, never Load, which
// rewrites autopus.yaml when it normalizes a platform name.
func readTopologyConfig() (*bool, string, string, error) {
	if _, err := os.Lstat("autopus.yaml"); errors.Is(err, fs.ErrNotExist) {
		return nil, "", "default", nil
	}
	cfg, err := config.LoadPreview(".")
	if err != nil {
		return nil, "", "", fmt.Errorf("config load failed: %w", err)
	}
	return cfg.Workflow.AutoTeam, cfg.Language.AIResponses, "autopus.yaml", nil
}
