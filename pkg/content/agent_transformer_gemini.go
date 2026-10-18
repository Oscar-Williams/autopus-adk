package content

import (
	"fmt"
	"strings"
)

// TransformAgentForGemini produces a Gemini MD template from an agent source.
// Its frontmatter follows the Antigravity CLI Markdown agent schema, because
// the plugin mirror of this file is the copy `agy` loads.
func TransformAgentForGemini(src AgentSource) string {
	var sb strings.Builder

	body := NormalizeAgentReferences(src.Body, "gemini")
	tools := geminiNativeTools(src.Meta.Tools)

	sb.WriteString("---\n")
	fmt.Fprintf(&sb, "name: auto-agent-%s\n", src.Meta.Name)
	fmt.Fprintf(&sb, "description: %s\n", src.Meta.Description)
	if len(tools) > 0 {
		sb.WriteString("tools:\n")
		for _, tool := range tools {
			fmt.Fprintf(&sb, "  - %s\n", tool)
		}
	}
	if grantsReaMCP(src.Meta.Tools) {
		sb.WriteString("inheritMcp: true\n")
	}
	if len(src.Meta.Skills) > 0 {
		sb.WriteString("skills:\n")
		for _, s := range src.Meta.Skills {
			fmt.Fprintf(&sb, "  - %s\n", s)
		}
	}
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	sb.WriteString("\n")

	return sb.String()
}

// geminiNativeTools maps source tools onto the names Antigravity CLI resolves
// for a Markdown agent. agy 1.4 refuses to construct an agent whose tools list
// names anything outside its registry ("unknown component: tool ... not found
// in registry"), which Gemini CLI's glob, replace, and run_shell_command are:
// Edit is edit_file (an exact Target and Replacement, unique unless
// AllowMultiple), Glob is find_by_name (glob patterns through fd), and Bash is
// run_command. grep_search and find_by_name left agy's default toolset in
// 1.2.7 and stay available to an agent that lists them.
// @AX:NOTE [AUTO]: This provider vocabulary allowlist intentionally omits unknown and advisory-only source tools.
func geminiNativeTools(source string) []string {
	native := map[string]string{
		"Read":  "read_file",
		"Write": "write_file",
		"Edit":  "edit_file",
		"Grep":  "grep_search",
		"Glob":  "find_by_name",
		"Bash":  "run_command",
	}
	seen := make(map[string]bool, len(native))
	tools := make([]string, 0, len(native))
	for _, candidate := range strings.Split(source, ",") {
		mapped, ok := native[strings.TrimSpace(candidate)]
		if !ok || seen[mapped] {
			continue
		}
		seen[mapped] = true
		tools = append(tools, mapped)
	}
	return tools
}

// grantsReaMCP reports whether a source grants REA MCP tools (SPEC-REA-001
// M2: reverse-engineer holds REA's static analysis tools one by one). An agy
// Markdown agent has no per-tool MCP grant: it reaches the session's MCP
// servers through its call_mcp_tool dispatcher, and each call still asks for
// agy's MCP permission. So the Antigravity projection lists no REA tool, which
// agy would reject as unknown, and declares inheritMcp, agy's Markdown field
// for the session's MCP servers (agy 1.4 hands a plugin subagent those servers
// either way). The OMP and OpenCode projections drop every REA reference
// (isReaMCPReference).
func grantsReaMCP(source string) bool {
	for _, candidate := range strings.Split(source, ",") {
		if isReaMCPReference(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}
