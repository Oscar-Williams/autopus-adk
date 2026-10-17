package content_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contentfs "github.com/insajin/autopus-adk/content"
	"github.com/insajin/autopus-adk/pkg/content"
)

// agyResolvedTools are tool names Antigravity CLI 1.4.0 resolved for a
// Markdown agent in the live probes of 2026-10-17: an agent listing them was
// constructed. agyUnknownTools failed construction with `unknown component:
// tool "<name>" not found in registry`, so a single one of them keeps the
// agent from starting.
var (
	agyResolvedTools = []string{
		"read_file", "view_file", "write_file", "write_to_file", "edit_file", "replace_file_content",
		"multi_replace_file_content", "grep_search", "find_by_name", "list_dir", "run_command",
		"search_web", "read_url_content",
	}
	agyUnknownTools = []string{"glob", "replace", "run_shell_command", "codebase_search", "command_status"}
)

// Every shipped agent's Antigravity projection lists only names agy resolves;
// the Gemini CLI vocabulary it carried before made every agent fail to start.
func TestTransformAgentForGemini_EveryAgentListsOnlyToolsAgyResolves(t *testing.T) {
	t.Parallel()
	sources, err := content.LoadAgentSourcesFromFS(contentfs.FS, "agents")
	require.NoError(t, err)
	require.NotEmpty(t, sources)
	for _, source := range sources {
		frontmatter, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(source))
		require.NotEmpty(t, frontmatter.Tools, source.Meta.Name)
		for _, tool := range frontmatter.Tools {
			assert.Contains(t, agyResolvedTools, tool, "%s lists %s", source.Meta.Name, tool)
			assert.NotContains(t, agyUnknownTools, tool, "%s lists %s", source.Meta.Name, tool)
		}
	}
}

// The executor keeps its write tools under agy's names, and a read-only
// reviewer still has no write tool.
func TestTransformAgentForGemini_WriteToolsFollowTheSource(t *testing.T) {
	t.Parallel()
	executor, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(ompAgentSource(t, "executor")))
	assert.Equal(t, []string{"read_file", "write_file", "edit_file", "grep_search", "find_by_name", "run_command"}, executor.Tools)
	reviewer, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(ompAgentSource(t, "reviewer")))
	assert.Equal(t, []string{"read_file", "grep_search", "find_by_name", "run_command"}, reviewer.Tools)
	assert.False(t, reviewer.InheritMCP)
}
