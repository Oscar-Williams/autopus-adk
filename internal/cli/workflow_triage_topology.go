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

// workflowTriageLong is the `auto workflow triage --help` text. It carries the
// auto-team cost and compatibility notice REQ-TA-14 asks every surface to give.
const workflowTriageLong = `Inspect an advisory workflow route from caller-declared facts.

With --topology, decide whether a Claude Code /auto go run without --team,
--workflow, or --solo runs its first ready Phase 2 wave as a native Agent Team.
The facts file lists the Phase 1 units (id, execution, owned_paths, depends_on,
done). The decision prints as JSON, or with --format human as its one reason
line. Auto-selection is on by default: it picks a team for an interactive
Claude Code session that has CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 and runs
Claude Code 2.1.246 or newer, when at least two ready worker units own disjoint
paths. The environment cannot tell a subagent from the top-level session, so a
session that cannot load the team tools prints tools_fallback_line instead and
stays on subagents.

Each teammate is a separate Claude session, so a team run uses more tokens, and
the feature is experimental. Turn it off with workflow.auto_team: false in
autopus.yaml, or pass --solo to /auto go for one run. auto 0.50.113 through
0.50.125 refuse an autopus.yaml that contains workflow.auto_team with any
value; auto 0.50.112 and older ignore the key and delete it on auto update.

The version is read only from the running binary CLAUDE_CODE_EXECPATH names,
which must be called claude or claude.exe; other installs stay on subagents
with the claude_version_unknown line. Facts that fail validation come back as
an invalid_facts decision whose reason line names the rejected field. Neither
mode launches an agent, writes a file, or changes a gate.`

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
// session environment, and, only when rules 1-7 pass, the version of the
// running Claude Code binary. That probe is its only child process. Facts the
// contract rejects are a subagent decision, not an error, so the reason line
// names the rejected field where the generated snippet would hide stderr.
func runWorkflowTriageTopology(out io.Writer, factsPath, format string) error {
	data, err := readWorkflowTriageFile(factsPath)
	if err != nil {
		return err
	}
	if len(data) > workflowTriageMaxBytes {
		return fmt.Errorf("facts file changed or exceeds limit")
	}
	autoTeam, locale, source, err := readTopologyConfig()
	if err != nil {
		return err
	}
	var decision taskroute.TeamDecision
	facts, err := taskroute.DecodeTeamFacts(data)
	if factsErr := (*taskroute.TeamFactsError)(nil); errors.As(err, &factsErr) {
		decision = taskroute.InvalidFactsDecision(locale, factsErr.Code)
	} else if err != nil {
		return err
	} else {
		snapshot, execPath := observeTeamCapability()
		input := taskroute.TeamInput{Flags: facts.Flags, AutoTeam: autoTeam, Units: facts.Units, Capability: snapshot, Locale: locale}
		if taskroute.NeedsVersionProbe(input) {
			input.Capability.Version = probeClaudeVersion(execPath)
		}
		decision = taskroute.DecideTeam(input)
	}
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
