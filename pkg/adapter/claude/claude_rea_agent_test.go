package claude_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S9 (Reverse Engineer Agent Contract): the Claude file grants the REA server
// and the read tools only, keeps the default permission mode, and projects
// the analysis rung under balanced and the opus tier under ultra.
func TestClaudeReverseEngineerAgent_FrontmatterPerQualityMode(t *testing.T) {
	t.Parallel()
	for mode, want := range map[string][2]string{
		"balanced": {"model: claude-opus-5-5", "effort: high"},
		"ultra":    {"model: opus", "effort: max"},
	} {
		lines := agentFrontmatterLines(t, generateClaudeAgentsForMode(t, mode), "reverse-engineer.md")
		assert.Contains(t, lines, "tools: Read, Grep, Glob, mcp__rea", mode)
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
		for _, denied := range []string{"Write", "Edit", "Bash"} {
			assert.NotContains(t, tools, denied, mode)
		}
	}
}
