package content_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/content"
)

// The Form Text unavailable sentences (SPEC-REA-001): one for the platforms
// Autopus registers REA on, one for OMP and OpenCode.
const (
	reaUnavailableRegistered = "REA MCP 도구를 찾지 못했습니다. auto update --rea로 이 프로젝트에 등록하거나 npx -y rea-agents@6.5.0 setup으로 사용자 범위에 등록하세요."
	reaUnavailableOther      = "이 플랫폼의 reverse-engineer에는 REA 도구가 없습니다. /auto reverse는 Claude Code·Codex·Antigravity에서 실행하세요."
)

// S9: the source frontmatter follows the Reverse Engineer Agent Contract.
func TestReverseEngineerSource_FrontmatterFollowsTheContract(t *testing.T) {
	t.Parallel()
	meta := ompAgentSource(t, "reverse-engineer").Meta
	assert.Equal(t, content.AgentSourceMeta{
		Name: "reverse-engineer", Description: meta.Description, Model: "opus", Effort: "high",
		Tools: "Read, Grep, Glob, mcp__rea", PermissionMode: "default", MaxTurns: 120,
	}, meta)
	assert.NotEmpty(t, meta.Description)
	for _, denied := range []string{"Write", "Edit", "Bash"} {
		assert.NotContains(t, meta.Tools, denied)
	}
}

// S9: Antigravity resolves the REA server through the mcp_<server>_* wildcard,
// after the read tools and in source order.
func TestReverseEngineerGemini_ToolsAreReadToolsAndTheREAServer(t *testing.T) {
	t.Parallel()
	frontmatter, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(ompAgentSource(t, "reverse-engineer")))
	assert.Equal(t, []string{"read_file", "grep_search", "glob", "mcp_rea_*"}, frontmatter.Tools)
	assert.Empty(t, frontmatter.Skills)
}

// S9: Autopus registers no REA server for OMP, so the agent keeps only the
// sorted read tools and no mcp_ value.
func TestReverseEngineerOMP_DropsTheREAServer(t *testing.T) {
	t.Parallel()
	rendered, err := content.RenderOMPAgentDefinition(ompAgentSource(t, "reverse-engineer"),
		content.OMPAgentDefinitionOptions{ThinkingLevel: "high"})
	require.NoError(t, err)
	fields := ompDefinitionFrontmatter(t, rendered)
	assert.Equal(t, []any{"glob", "grep", "read"}, fields["tools"])
	assert.Equal(t, "high", fields["thinkingLevel"])
	assert.NotContains(t, fields, "model")
}

// S9: OpenCode denies everything but the read tools and its standing
// question, skill, and task permissions; no mcp__ key is allowed.
func TestReverseEngineerOpenCode_AllowsNoMCPKey(t *testing.T) {
	t.Parallel()
	rendered := content.TransformAgentForOpenCode(ompAgentSource(t, "reverse-engineer"))
	parts := strings.SplitN(rendered, "---\n", 3)
	require.Len(t, parts, 3)
	var frontmatter struct {
		Permission map[string]string `yaml:"permission"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &frontmatter))
	allowed := make([]string, 0, len(frontmatter.Permission))
	for key, value := range frontmatter.Permission {
		if key == "*" {
			assert.Equal(t, "deny", value)
			continue
		}
		assert.Equal(t, "allow", value, key)
		allowed = append(allowed, key)
	}
	sort.Strings(allowed)
	assert.Equal(t, []string{"glob", "grep", "question", "read", "skill", "task"}, allowed)
	assert.NotContains(t, parts[1], "mcp__")
}

// The other agents keep their MCP tools exactly as before: only the REA
// server is platform-gated.
func TestREAToolGate_LeavesOtherMCPToolsUntouched(t *testing.T) {
	t.Parallel()
	source := content.AgentSource{Meta: content.AgentSourceMeta{
		Name: "mcp-fixture", Tools: "Read, mcp__context7, mcp__reader, mcp__rea__inspect",
	}}
	rendered, err := content.RenderOMPAgentDefinition(source, content.OMPAgentDefinitionOptions{})
	require.NoError(t, err)
	assert.Equal(t, []any{"mcp__context7", "mcp__reader", "read"}, ompDefinitionFrontmatter(t, rendered)["tools"])
	openCode := content.TransformAgentForOpenCode(source)
	assert.Contains(t, openCode, `"mcp__reader": allow`)
	assert.NotContains(t, openCode, "mcp__rea__inspect")
	frontmatter, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(source))
	assert.Equal(t, []string{"read_file"}, frontmatter.Tools)
}

// S9: every platform's body demands the five-section behaviour specification,
// the Runtime Capture Request, and both unavailable sentences, and never
// reaches for a user-question tool.
func TestReverseEngineerBody_EveryPlatformCarriesTheAnalysisContract(t *testing.T) {
	t.Parallel()
	source := ompAgentSource(t, "reverse-engineer")
	omp, err := content.RenderOMPAgentDefinition(source, content.OMPAgentDefinitionOptions{})
	require.NoError(t, err)
	rendered := map[string]string{
		"claude":   source.Body,
		"codex":    content.TransformAgentForCodex(source),
		"gemini":   content.TransformAgentForGemini(source),
		"omp":      omp,
		"opencode": content.TransformAgentForOpenCode(source),
	}
	for platform, body := range rendered {
		for _, want := range []string{
			"## Behaviour", "## Inputs and Outputs", "## Evidence", "## Unknowns", "## Acceptance Items",
			"B-1", "A-1", "evidence_id", "## Runtime Capture Request",
			"- target:", "- tool:", "- will run:", "- why:",
			"Runtime capture wanted: yes", "Runtime capture consent: <tool> on <target>",
			reaUnavailableRegistered, reaUnavailableOther,
		} {
			assert.Contains(t, body, want, "%s: %s", platform, want)
		}
		assert.NotContains(t, body, "AskUserQuestion", platform)
	}
}
