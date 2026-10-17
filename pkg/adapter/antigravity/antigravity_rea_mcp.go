package antigravity

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// antigravityReaAction returns the run's REA action. Generation that never
// resolved the REA decision carries the zero value and keeps the entry.
func antigravityReaAction(cfg *config.HarnessConfig) config.ReaMCPAction {
	if cfg == nil {
		return config.ReaMCPKeep
	}
	return cfg.Runtime.ReaMCP
}

// sweepLegacyReaMCP removes a managed rea entry from the mcpServers object of
// decoded .gemini/settings.json, whatever the run's action: agy never reads
// MCP servers from that file, and release candidates of v0.50.126 wrote the
// entry there. It runs after mergeSettingsMaps, which can only add or
// overwrite keys. other_pin and user-owned entries stay.
//
// Generate and Clean write settings with os.WriteFile, which follows a
// symlink, so a sweep that changes content first checks the path; a refused
// path ends the platform's run with an error before any settings byte is
// written. Settings without a managed entry come out as before.
func (a *Adapter) sweepLegacyReaMCP(settings map[string]any) error {
	servers, _ := settings["mcpServers"].(map[string]any)
	out, changed := reamcp.ApplyJSON(reamcp.Antigravity, servers, config.ReaMCPRemove)
	if !changed {
		return nil
	}
	rel := reamcp.LegacyTargetRelPath(reamcp.Antigravity)
	if _, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(rel)); err != nil {
		return fmt.Errorf("REA MCP 쓰기 거부 %s: %w", rel, err)
	}
	settings["mcpServers"] = out
	return nil
}

// reaPluginMCP is what a run's REA action does to the plugin's mcp_config.json:
// the bytes to write, or, when the managed entry was all the file held, its
// removal.
type reaPluginMCP struct {
	content []byte
	remove  bool
}

// planReaPluginMCP applies action to .agents/plugins/autopus/mcp_config.json,
// the file agy reads the workspace plugin's MCP servers from. It reports false
// when the file stays as it is, and a keep run never reads it. The file is
// not a manifest mapping: the user may add servers to it, and pruning or
// cleaning the plugin tree must never delete those, so Autopus owns only the
// managed entry, exactly as in the other REA target files.
//
// The file is read through reamcp.ReadTarget, the guarded read the CLI uses:
// a symlink anywhere on the path, a file that is not regular, or one larger
// than reamcp.TargetLimit is never read, and the add or remove that needed it
// ends the run with an error naming the file. A file that does not decode
// cannot hold a managed entry, so remove leaves it alone, and add refuses it
// rather than overwrite it. A change whose path crosses a symlink or leaves
// the project is refused before anything is written.
func (a *Adapter) planReaPluginMCP(action config.ReaMCPAction) (reaPluginMCP, bool, error) {
	if action != config.ReaMCPAdd && action != config.ReaMCPRemove {
		return reaPluginMCP{}, false, nil
	}
	rel := reamcp.TargetRelPath(reamcp.Antigravity)
	data, err := reamcp.ReadTarget(a.root, rel)
	if err != nil {
		// The read error already names the file.
		return reaPluginMCP{}, false, fmt.Errorf("REA MCP 쓰기 거부: %w", err)
	}
	out, changed, err := reamcp.ApplyFile(reamcp.Antigravity, data, action)
	switch {
	case err != nil && action == config.ReaMCPRemove:
		return reaPluginMCP{}, false, nil
	case err != nil:
		// The decode error already names the file.
		return reaPluginMCP{}, false, fmt.Errorf("REA MCP 쓰기 거부: %w", err)
	case !changed:
		return reaPluginMCP{}, false, nil
	}
	if _, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(rel)); err != nil {
		return reaPluginMCP{}, false, fmt.Errorf("REA MCP 쓰기 거부 %s: %w", rel, err)
	}
	return reaPluginMCP{content: out, remove: reamcp.RemovesFile(reamcp.Antigravity, out)}, true, nil
}

// writeReaPluginMCP applies a planned change outside a transaction (Generate
// and Clean).
func (a *Adapter) writeReaPluginMCP(change reaPluginMCP) error {
	rel := reamcp.TargetRelPath(reamcp.Antigravity)
	path, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(rel))
	if err != nil {
		return fmt.Errorf("REA MCP 쓰기 거부 %s: %w", rel, err)
	}
	if change.remove {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("REA MCP 제거 실패 %s: %w", rel, err)
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("REA MCP 디렉터리 생성 실패 %s: %w", rel, err)
	}
	if err := adapter.WriteFileIfChanged(path, change.content, 0o644); err != nil {
		return fmt.Errorf("REA MCP 쓰기 실패 %s: %w", rel, err)
	}
	return nil
}

// addReaPluginMCPToPlan puts a planned change into an update transaction, so
// it commits or rolls back with the platform's other writes. The manifest the
// plan records stays without the file.
func addReaPluginMCPToPlan(plan *adapter.TransactionPlan, change reaPluginMCP) {
	rel := reamcp.TargetRelPath(reamcp.Antigravity)
	if change.remove {
		plan.Removes = append(plan.Removes, adapter.TransactionRemove{Path: rel})
		return
	}
	plan.Writes = append(plan.Writes, adapter.TransactionWrite{Path: rel, Content: change.content, Perm: 0o644})
}
