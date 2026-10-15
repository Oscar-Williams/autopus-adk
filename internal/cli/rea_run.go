package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaRun carries one init or update run's REA decision from its resolution,
// through the action every platform write of the run reads, to the
// autopus.yaml write after those writes (Decision State Contract).
type reaRun struct {
	text    reaText
	probe   *reaNodeProbe
	state   reaState
	action  config.ReaMCPAction
	active  bool // a REA target platform is selected; otherwise only the OMP·OpenCode line remains
	heldKey bool // autopus.yaml held features.rea before the run
	// before holds each REA target file as the run found it (nil when
	// missing), so finish can name the files this run registered.
	before map[reamcp.Platform][]byte
}

// newReaRun starts a run's REA state over cfg's selected platforms. It reads
// the three target files before any platform write.
func newReaRun(dir string, cfg *config.HarnessConfig, heldKey bool) *reaRun {
	run := &reaRun{
		text:    reaCurrentText(),
		probe:   newReaNodeProbe(),
		state:   reaUndecided,
		active:  len(reaSelectedTargets(cfg.Platforms)) > 0,
		heldKey: heldKey,
		before:  map[reamcp.Platform][]byte{},
	}
	for _, p := range reamcp.Platforms() {
		if data, err := readReaTarget(dir, p); err == nil {
			run.before[p] = data
		}
	}
	return run
}

// checkReaFlags rejects --rea with --no-rea, and either with a workspace
// run, before anything is written (REQ-REA-02, REQ-REA-07).
func checkReaFlags(on, off, workspace bool) error {
	text := reaCurrentText()
	switch {
	case on && off:
		return errors.New(text.FlagBoth)
	case (on || off) && workspace:
		return errors.New(text.FlagWorkspace)
	}
	return nil
}

// registerReaFlags adds --rea and --no-rea to init or update.
func registerReaFlags(cmd *cobra.Command, on, off *bool) {
	cmd.Flags().BoolVar(on, "rea", false, "REA 역공학 MCP를 Claude Code·Codex·Antigravity 프로젝트 범위에 등록 (Node "+reamcp.NodeRange+")")
	cmd.Flags().BoolVar(off, "no-rea", false, "REA 역공학 MCP를 등록하지 않고 managed 항목을 제거")
}

// reaContext returns the command's context, or Background before one is set.
func reaContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

// decide records the run's decision and sets the action the platform writes
// read: on with a passing Node gate adds, on with a failing gate keeps and
// prints one Node notice, off removes, and undecided keeps. Only an on
// decision of an active run asks the probe, which runs once per run.
func (r *reaRun) decide(ctx context.Context, out io.Writer, cfg *config.HarnessConfig, state reaState) {
	r.state = state
	r.action = config.ReaMCPKeep
	switch {
	case !r.active:
	case state == reaOff:
		r.action = config.ReaMCPRemove
	case state == reaOn:
		if gate := r.probe.Gate(ctx); gate.Status == reamcp.NodePass {
			r.action = config.ReaMCPAdd
		} else {
			r.line(out, r.nodeNotice(gate))
		}
	}
	cfg.Runtime.ReaMCP = r.action
}

// nodeNotice is the Node line for a gate that did not pass. Only the judged
// version reaches the terminal, never other probe output.
func (r *reaRun) nodeNotice(gate reamcp.NodeGate) string {
	switch gate.Status {
	case reamcp.NodeUnsupported:
		return fmt.Sprintf(r.text.NodeOutside, terminalSafe(gate.Version))
	case reamcp.NodeAbsent:
		return r.text.NodeAbsent
	}
	return r.text.NodeUnknown
}

// finish runs after the run's platform writes. An off decision first sweeps
// the managed entry out of the unselected target files; a refused sweep is
// reported but still records the decision. M is then read back from the
// selected target files, autopus.yaml is written by the write table, and the
// registration, old-binary, and OMP·OpenCode lines are printed once.
func (r *reaRun) finish(out io.Writer, dir string, cfg *config.HarnessConfig) error {
	var sweepErr error
	if r.active {
		if r.state == reaOff {
			sweepErr = sweepReaUnselected(dir, cfg.Platforms)
		}
		if files := r.registered(dir, cfg.Platforms); len(files) > 0 {
			r.line(out, fmt.Sprintf(r.text.Registered, strings.Join(files, ", ")))
		}
		next := reaKeyAfterRun(r.state, cfg.Features.Rea.Enabled, reaManaged(dir, cfg.Platforms))
		if !reaSameKey(cfg.Features.Rea.Enabled, next) {
			cfg.Features.Rea.Enabled = next
			if err := config.Save(dir, cfg); err != nil {
				return errors.Join(sweepErr, fmt.Errorf("features.rea 저장 실패: %w", err))
			}
		}
		if !r.heldKey && cfg.Features.Rea.Enabled != nil {
			r.line(out, r.text.OldBinary)
		}
	}
	r.guidance(out, cfg.Platforms)
	return sweepErr
}

// registered lists the selected target files this run gave the current
// managed entry: managed now, and the add action would have changed the file
// as the run found it. A rerun over current entries registers nothing.
func (r *reaRun) registered(dir string, platforms []string) []string {
	if r.action != config.ReaMCPAdd {
		return nil
	}
	var files []string
	for _, p := range reaSelectedTargets(platforms) {
		before, known := r.before[p]
		if !known || reaFileClass(dir, p) != reamcp.ClassManaged {
			continue
		}
		if _, changed, err := reamcp.ApplyFile(p, before, config.ReaMCPAdd); err == nil && changed {
			files = append(files, reamcp.TargetRelPath(p))
		}
	}
	return files
}

// guidance prints the OMP·OpenCode line when the project includes either and
// the decision is not off (REQ-REA-09).
func (r *reaRun) guidance(out io.Writer, platforms []string) {
	if r.state != reaOff && len(reaGuidancePlatforms(platforms)) > 0 {
		r.line(out, r.text.Guidance)
	}
}

func (r *reaRun) line(out io.Writer, text string) {
	fmt.Fprintf(out, "  %s\n", text)
}
