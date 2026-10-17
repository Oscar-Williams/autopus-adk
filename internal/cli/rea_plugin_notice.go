package cli

import (
	"bytes"
	"fmt"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaPluginState is the Antigravity plugin MCP file as a run finds it. agy
// reads comments and trailing commas in that file, which strict JSON refuses,
// so Autopus treats such a file as user-owned and never rewrites it. invalid
// marks that state, and namesRea whether its bytes still name rea-agents.
type reaPluginState struct {
	invalid  bool
	namesRea bool
}

// readReaPluginState reads the plugin MCP file through the guarded read. A
// missing, unreadable, or strict-JSON file is not invalid: the first two have
// nothing to say about REA and the last is the adapter's to change.
func readReaPluginState(dir string) reaPluginState {
	data, err := readReaTarget(dir, reamcp.Antigravity)
	if err != nil || data == nil {
		return reaPluginState{}
	}
	if _, err := reamcp.ClassifyFile(reamcp.Antigravity, data); err == nil {
		return reaPluginState{}
	}
	return reaPluginState{invalid: true, namesRea: bytes.Contains(data, []byte("rea-agents@"))}
}

// reaPluginNotice is the one line for a plugin file a REA action could not
// change: add could not register in it, or remove could not take out the rea
// entry it still names. Every other state and action says nothing.
func reaPluginNotice(text reaText, state reaPluginState, action config.ReaMCPAction) string {
	rel := reamcp.TargetRelPath(reamcp.Antigravity)
	switch {
	case !state.invalid:
		return ""
	case action == config.ReaMCPAdd:
		return fmt.Sprintf(text.PluginNotRegistered, rel)
	case action == config.ReaMCPRemove && state.namesRea:
		return fmt.Sprintf(text.PluginNotRemoved, rel)
	}
	return ""
}

// reaPluginSkipItems is the preview line for the same state: the run would
// leave the plugin file alone instead of changing it.
func reaPluginSkipItems(data []byte, action config.ReaMCPAction) []previewItem {
	var reason string
	switch {
	case action == config.ReaMCPAdd:
		reason = "REA would not register here: not strict JSON (comments or trailing commas); remove them or add the rea entry by hand"
	case action == config.ReaMCPRemove && bytes.Contains(data, []byte("rea-agents@")):
		reason = "REA off cannot remove the rea entry here: not strict JSON (comments or trailing commas); remove them or delete the entry by hand"
	default:
		return nil
	}
	return []previewItem{{Path: reamcp.TargetRelPath(reamcp.Antigravity), Kind: "skip",
		Scope: string(reamcp.Antigravity), Reason: reason}}
}
