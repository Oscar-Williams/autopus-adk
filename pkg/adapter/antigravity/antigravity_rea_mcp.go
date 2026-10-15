package antigravity

import (
	"fmt"
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

// applyAntigravityReaMCP applies action to the mcpServers object of decoded
// .gemini/settings.json. It runs after mergeSettingsMaps, which can only add
// or overwrite keys, so remove is what deletes mcpServers.rea.
//
// Generate and Clean write settings with os.WriteFile, which follows a
// symlink; only a run whose REA action changes content is checked, so a keep
// run behaves exactly as before. A refused path ends the platform's run with
// an error before any settings byte is written.
func (a *Adapter) applyAntigravityReaMCP(settings map[string]any, action config.ReaMCPAction) error {
	servers, _ := settings["mcpServers"].(map[string]any)
	out, changed := reamcp.ApplyJSON(reamcp.Antigravity, servers, action)
	if !changed {
		return nil
	}
	rel := reamcp.TargetRelPath(reamcp.Antigravity)
	if _, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(rel)); err != nil {
		return fmt.Errorf("REA MCP 쓰기 거부 %s: %w", rel, err)
	}
	settings["mcpServers"] = out
	return nil
}
