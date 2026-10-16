package content_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/content"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaStaticToolSource is the source tools value M2 requires: the read tools
// and REA's audited static analysis tools, one by one.
func reaStaticToolSource() string {
	tools := []string{"Read", "Grep", "Glob"}
	for _, tool := range reamcp.StaticAnalysisTools() {
		tools = append(tools, "mcp__rea__"+tool)
	}
	return strings.Join(tools, ", ")
}

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
		Tools: reaStaticToolSource(), PermissionMode: "default", MaxTurns: 120,
	}, meta)
	assert.NotEmpty(t, meta.Description)
	for _, tool := range strings.Split(meta.Tools, ", ") {
		assert.NotContains(t, []string{"Write", "Edit", "Bash", "mcp__rea"}, tool, "no write tool and no whole-server grant")
	}
	for _, runtime := range append(reamcp.RuntimeCaptureTools(), reamcp.CallerPathWriteTools()...) {
		assert.NotContains(t, strings.Split(meta.Tools, ", "), "mcp__rea__"+runtime, "M2: %s is main-session only", runtime)
	}
}

// S9 and M2: Antigravity grants the read tools and each static REA tool as
// mcp_rea_<tool>, in source order, and never the mcp_rea_* wildcard.
func TestReverseEngineerGemini_ToolsAreReadToolsAndStaticREATools(t *testing.T) {
	t.Parallel()
	frontmatter, _ := decodeGeminiAgent(t, content.TransformAgentForGemini(ompAgentSource(t, "reverse-engineer")))
	want := []string{"read_file", "grep_search", "glob"}
	for _, tool := range reamcp.StaticAnalysisTools() {
		want = append(want, "mcp_rea_"+tool)
	}
	assert.Equal(t, want, frontmatter.Tools)
	assert.NotContains(t, frontmatter.Tools, "mcp_rea_*")
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

// S9 and M1: OpenCode denies everything but the read tools and the standing
// question and skill permissions: no mcp__ key, and no task, so the analysis
// agent cannot dispatch another agent.
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
	assert.Equal(t, []string{"glob", "grep", "question", "read", "skill"}, allowed)
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
	assert.Equal(t, []string{"read_file", "mcp_rea_inspect"}, frontmatter.Tools)
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
			"- target:", "- tool:", "- arguments:", "- will run:", "- why:",
			"Runtime capture wanted: yes", "Runtime capture evidence: <tool> on <target>: <evidence_id>",
			reaUnavailableRegistered, reaUnavailableOther,
		} {
			assert.Contains(t, body, want, "%s: %s", platform, want)
		}
		assert.NotContains(t, body, "AskUserQuestion", platform)
	}
}

// M1: Codex agent roles take no tool list, so only reverse-engineer's role file
// carries the read-only sandbox the role schema accepts.
func TestReverseEngineerCodex_RoleRunsInTheReadOnlySandbox(t *testing.T) {
	t.Parallel()
	assert.Contains(t, content.TransformAgentForCodex(ompAgentSource(t, "reverse-engineer")), "\nsandbox_mode = \"read-only\"\n")
	assert.NotContains(t, content.TransformAgentForCodex(ompAgentSource(t, "executor")), "sandbox_mode")
}
