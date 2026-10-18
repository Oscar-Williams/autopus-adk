package claude

import (
	"fmt"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// claudeReaServerName is the .mcp.json server owned by the run's REA action.
// The merge's managed-server sweep leaves it alone, so only
// applyClaudeReaMCP ever adds, replaces, or removes it.
const claudeReaServerName = "rea"

// claudeReaAction returns the run's REA action. Generation that never
// resolved the REA decision carries the zero value and keeps the entry.
func claudeReaAction(cfg *config.HarnessConfig) config.ReaMCPAction {
	if cfg == nil {
		return config.ReaMCPKeep
	}
	return cfg.Runtime.ReaMCP
}

// applyClaudeReaMCP applies action to the mcpServers object of a decoded
// .mcp.json document and reports whether the rea entry changed. A write that
// changes REA content must stay inside the project, so a target path that
// crosses a symlink is refused before anything is written.
func (a *Adapter) applyClaudeReaMCP(doc map[string]any, action config.ReaMCPAction) (bool, error) {
	servers, _ := doc["mcpServers"].(map[string]any)
	out, changed := reamcp.ApplyJSON(reamcp.Claude, servers, action)
	if !changed {
		return false, nil
	}
	rel := reamcp.TargetRelPath(reamcp.Claude)
	if _, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(rel)); err != nil {
		return false, fmt.Errorf("REA MCP 쓰기 거부 %s: %w", rel, err)
	}
	doc["mcpServers"] = out
	return true, nil
}
