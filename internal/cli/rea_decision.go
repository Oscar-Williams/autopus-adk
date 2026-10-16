package cli

import (
	"errors"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaState is a project's REA decision (SPEC-REA-001 Decision State Contract).
type reaState string

const (
	reaOn        reaState = "on"
	reaOff       reaState = "off"
	reaUndecided reaState = "undecided"
)

// reaSource names what a decision was read from; doctor reports it.
type reaSource string

const (
	reaSourceKey     reaSource = "features.rea"
	reaSourceManaged reaSource = "managed_entry"
	reaSourceNone    reaSource = "none"
)

// reaDecision is a resolved decision and its source.
type reaDecision struct {
	State  reaState
	Source reaSource
}

// resolveReaDecision applies the Decision State Contract: a stored
// features.rea.enabled wins, a managed rea entry in the target file of a
// selected REA target platform means on, and every other state, including
// files with only user-owned or other-pin entries, is undecided.
func resolveReaDecision(dir string, cfg *config.HarnessConfig) reaDecision {
	if enabled := cfg.Features.Rea.Enabled; enabled != nil {
		if *enabled {
			return reaDecision{State: reaOn, Source: reaSourceKey}
		}
		return reaDecision{State: reaOff, Source: reaSourceKey}
	}
	if reaManaged(dir, cfg.Platforms) {
		return reaDecision{State: reaOn, Source: reaSourceManaged}
	}
	return reaDecision{State: reaUndecided, Source: reaSourceNone}
}

// reaFlagState is the decision --rea or --no-rea names, else undecided.
// checkReaFlags has already rejected the two together.
func reaFlagState(on, off bool) reaState {
	switch {
	case on:
		return reaOn
	case off:
		return reaOff
	}
	return reaUndecided
}

// reaSelectedTargets returns the REA target platforms among platforms, in
// reamcp order.
func reaSelectedTargets(platforms []string) []reamcp.Platform {
	var selected []reamcp.Platform
	for _, p := range reamcp.Platforms() {
		if containsPlatform(platforms, string(p)) {
			selected = append(selected, p)
		}
	}
	return selected
}

// reaGuidancePlatforms returns omp and opencode, in that order, when
// platforms includes them: Autopus writes no REA MCP entry for either.
func reaGuidancePlatforms(platforms []string) []string {
	guided := []string{}
	for _, p := range []string{"omp", "opencode"} {
		if containsPlatform(platforms, p) {
			guided = append(guided, p)
		}
	}
	return guided
}

// reaFileClass classifies the rea entry of p's target file. A file that
// cannot be read or decoded is user_owned: no REA write ever changes it.
func reaFileClass(dir string, p reamcp.Platform) reamcp.Class {
	data, err := readReaTarget(dir, p)
	if err != nil {
		return reamcp.ClassUserOwned
	}
	class, err := reamcp.ClassifyFile(p, data)
	if err != nil {
		return reamcp.ClassUserOwned
	}
	return class
}

// reaManaged is M: whether the target file of any selected REA target
// platform holds a managed entry. A run reads it after its platform writes.
func reaManaged(dir string, platforms []string) bool {
	for _, p := range reaSelectedTargets(platforms) {
		if reaFileClass(dir, p) == reamcp.ClassManaged {
			return true
		}
	}
	return false
}

// reaPlannedManaged is M for a preview, read from the planned mappings: the
// run's action applied to each selected target file as it is now. A file the
// platform manifest tracks but the operator deleted stays deleted
// (adapter.ActionSkip), so it is not in M.
func reaPlannedManaged(dir string, platforms []string, action config.ReaMCPAction) bool {
	for _, p := range reaSelectedTargets(platforms) {
		data, err := readReaTarget(dir, p)
		if err != nil || (data == nil && reaTargetSkipped(dir, p)) {
			continue
		}
		planned, _, err := reamcp.ApplyFile(p, data, action)
		if err != nil {
			continue
		}
		if class, err := reamcp.ClassifyFile(p, planned); err == nil && class == reamcp.ClassManaged {
			return true
		}
	}
	return false
}

// reaTargetSkipped reports a missing target file that p's manifest tracks:
// update skips it instead of recreating it (adapter.ResolveAction).
func reaTargetSkipped(dir string, p reamcp.Platform) bool {
	manifest, err := adapter.LoadManifest(dir, string(p))
	if err != nil || manifest == nil {
		return false
	}
	rel := reamcp.TargetRelPath(p)
	_, slash := manifest.Files[rel]
	_, native := manifest.Files[filepath.FromSlash(rel)]
	return slash || native
}

// reaKeyAfterRun applies the Decision State Contract write table: the
// features.rea.enabled value autopus.yaml holds after a run whose decision is
// state, given the stored value and M. On keeps an absent key while a managed
// entry carries the decision and otherwise stores true; off stores false;
// undecided writes nothing.
func reaKeyAfterRun(state reaState, stored *bool, managed bool) *bool {
	switch state {
	case reaOn:
		if managed && stored == nil {
			return nil
		}
		return reaBool(true)
	case reaOff:
		return reaBool(false)
	}
	return stored
}

func reaBool(value bool) *bool { return &value }

// reaSameKey reports whether two features.rea.enabled values are equal.
func reaSameKey(a, b *bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// reaSweepPreviewItems lists, for a preview whose run turns REA off, each
// unselected target file the off sweep would change, so `auto update --plan`
// shows every file the run writes. Any other run lists nothing.
func reaSweepPreviewItems(dir string, cfg *config.HarnessConfig) []previewItem {
	if cfg == nil || cfg.Runtime.ReaMCP != config.ReaMCPRemove {
		return nil
	}
	var items []previewItem
	for _, p := range reamcp.Platforms() {
		if containsPlatform(cfg.Platforms, string(p)) {
			continue
		}
		data, err := readReaTarget(dir, p)
		if err != nil || data == nil {
			continue
		}
		if _, changed, err := reamcp.ApplyFile(p, data, config.ReaMCPRemove); err == nil && changed {
			items = append(items, previewItem{Path: reamcp.TargetRelPath(p), Kind: "update", Scope: string(p),
				Reason: "REA off would remove the managed rea entry of this unselected platform"})
		}
	}
	return items
}

// sweepReaUnselected removes the managed rea entry from the target file of
// every REA target platform outside platforms (the off sweep). No adapter runs
// for those files, so the CLI applies the same reamcp transform itself.
func sweepReaUnselected(dir string, platforms []string) error {
	var errs []error
	for _, p := range reamcp.Platforms() {
		if !containsPlatform(platforms, string(p)) {
			errs = append(errs, removeReaEntry(dir, p))
		}
	}
	return errors.Join(errs...)
}
