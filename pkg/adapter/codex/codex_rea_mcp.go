package codex

import (
	"fmt"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// The rea server is deliberately absent from codexManagedConfigKeys: that
// table is per key, so listing it would overwrite the command and args of a
// user's own [mcp_servers.rea]. The whole managed table changes only through
// the run's REA action, applied after the structural merge.

// codexReaAction returns the run's REA action. Generation that never
// resolved the REA decision carries the zero value and keeps the table.
func codexReaAction(cfg *config.HarnessConfig) config.ReaMCPAction {
	if cfg == nil {
		return config.ReaMCPKeep
	}
	return cfg.Runtime.ReaMCP
}

// applyCodexReaMCP applies action to merged Codex config text. A write that
// changes REA content must stay inside the project, so a target path that
// crosses a symlink is refused before anything is written.
func (a *Adapter) applyCodexReaMCP(text string, action config.ReaMCPAction) (string, error) {
	out, changed := reamcp.ApplyCodexTOML(text, action)
	if !changed {
		return text, nil
	}
	if _, err := adapter.SafeWorkspacePath(a.root, filepath.FromSlash(codexConfigRelPath)); err != nil {
		return "", fmt.Errorf("REA MCP 쓰기 거부 %s: %w", codexConfigRelPath, err)
	}
	return out, nil
}
