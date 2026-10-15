package content

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/templates"
)

// S12: only references whose first `__` segment after `mcp__` is exactly
// `rea` survive verbatim; every other mcp__ reference is rewritten as before,
// including mcp__reader, whose name merely starts with rea.
func TestReplacePlatformReferences_KeepsREAReferencesVerbatim(t *testing.T) {
	t.Parallel()
	body := strings.Join([]string{
		"Call mcp__rea__list_unknowns, then mcp__rea for the server.",
		`Use mcp__context7__query-docs("react", topic="hooks").`,
		"Fallback mcp__other__y.",
		"Mixed mcp__rea__inspect_analysis_view and mcp__other__y.",
		"Prefix mcp__reader__x.",
	}, "\n")
	want := strings.Join([]string{
		"Call mcp__rea__list_unknowns, then mcp__rea for the server.",
		`Use Context7 MCP first; fallback: WebSearch "react hooks docs" (topic: hooks).`,
		"Fallback Context7 MCP with WebSearch fallback.",
		"Mixed mcp__rea__inspect_analysis_view and Context7 MCP with WebSearch fallback.",
		"Prefix Context7 MCP with WebSearch fallback.",
	}, "\n")
	for _, platform := range []string{"codex", "gemini", "opencode", "omp"} {
		assert.Equal(t, want, ReplacePlatformReferences(body, platform), platform)
	}
	assert.Equal(t, body, ReplacePlatformReferences(body, "claude"))
}

func TestIsReaMCPReference_ComparesTheFirstSegmentExactly(t *testing.T) {
	t.Parallel()
	for reference, want := range map[string]bool{
		"mcp__rea":                    true,
		"mcp__rea__":                  true,
		"mcp__rea__list_unknowns":     true,
		"mcp__rea__inspect__deep":     true,
		"mcp__reader__x":              false,
		"mcp__rea_agents__x":          false,
		"mcp__reax":                   false,
		"mcp__context7__query-docs":   false,
		"mcp__other__rea":             false,
		"mcp__sequential-thinking__x": false,
	} {
		assert.Equal(t, want, isReaMCPReference(reference), reference)
	}
}

// S12: the generated Codex and Antigravity reverse-engineer templates keep the
// REA tool references of the agent body verbatim.
func TestGeneratedReverseEngineerTemplates_KeepREAReferences(t *testing.T) {
	t.Parallel()
	for _, path := range []string{
		"codex/agents/reverse-engineer.toml.tmpl",
		"gemini/agents/reverse-engineer.md.tmpl",
	} {
		data, err := fs.ReadFile(templates.FS, path)
		require.NoError(t, err, path)
		assert.Contains(t, string(data), "Claude Code `mcp__rea__*`", path)
		assert.NotContains(t, string(data), "Context7 MCP with WebSearch fallback", path)
	}
}
