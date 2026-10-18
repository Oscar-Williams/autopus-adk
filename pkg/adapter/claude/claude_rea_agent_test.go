package claude_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// S9 (Reverse Engineer Agent Contract) and security audit M2: the Claude file
// grants the read tools and REA's static analysis tools one by one, never the
// whole server, keeps the default permission mode, and projects the analysis
// rung under balanced and the opus tier under ultra.
func TestClaudeReverseEngineerAgent_FrontmatterPerQualityMode(t *testing.T) {
	t.Parallel()
	for mode, want := range map[string][2]string{
		"balanced": {"model: claude-opus-5-5", "effort: high"},
		"ultra":    {"model: opus", "effort: max"},
	} {
		lines := agentFrontmatterLines(t, generateClaudeAgentsForMode(t, mode), "reverse-engineer.md")
		staticTools := []string{"Read", "Grep", "Glob"}
		for _, tool := range reamcp.StaticAnalysisTools() {
			staticTools = append(staticTools, "mcp__rea__"+tool)
		}
		assert.Contains(t, lines, "tools: "+strings.Join(staticTools, ", "), mode)
		assert.Contains(t, lines, "permissionMode: default", mode)
		assert.Contains(t, lines, want[0], mode)
		assert.Contains(t, lines, want[1], mode)
		var tools string
		for _, line := range lines {
			if strings.HasPrefix(line, "tools:") {
				tools = line
			}
		}
		require.NotEmpty(t, tools, mode)
		granted := strings.Split(strings.TrimPrefix(tools, "tools: "), ", ")
		for _, denied := range []string{"Write", "Edit", "Bash", "mcp__rea", "mcp__rea__capture_process_scenario"} {
			assert.NotContains(t, granted, denied, mode)
		}
	}
}
