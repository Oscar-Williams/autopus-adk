package adapter_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/templates"
)

// reverseContractNeedles are the S13 strings (SPEC-REA-001) every platform's
// auto-reverse detail must carry: the --runtime entry, the consent loop in
// which the main session itself calls the runtime tool (security audit M2), the
// confirmation gate that --auto never skips, the slug rule, the local
// behaviour-spec artifact, the clean-room handoff, and the dispatch ban.
var reverseContractNeedles = []string{
	"--runtime",
	"Runtime capture wanted: yes",
	"## Runtime Capture Request",
	"Runtime capture evidence:",
	"메인 세션이 그 REA 도구를 요청의 arguments로 직접 부르고",
	"허용 / 거절",
	"확인 / 수정 / 중단",
	"--auto도 이 확인 gate를 건너뛰지 않습니다",
	"`로그인 화면` → `feature`",
	".autopus/reverse/<slug>/behaviour-spec.md",
	"SPEC: N/A",
	"Task: reverse/<slug>",
	"Behaviour Spec:",
	"Phase: Test Scaffold",
	"Phase: Testing",
	"dispatch prompt에 REA 결과, evidence bundle, 디컴파일 텍스트를 넣지 않습니다",
}

// reverseDetails maps each platform to its auto-reverse detail. verbatim marks
// the platforms that render the shared contract without an invocation rewrite;
// only Codex rewrites `/auto` and `$auto` to its own entrypoints.
var reverseDetails = []struct {
	platform string
	path     string
	verbatim bool
}{
	{platform: "claude-code", path: ".claude/skills/auto-reverse/SKILL.md", verbatim: true},
	{platform: "antigravity-cli", path: ".gemini/skills/autopus/auto-reverse/SKILL.md", verbatim: true},
	{platform: "antigravity-cli", path: ".agents/plugins/autopus/skills/auto-reverse/SKILL.md", verbatim: true},
	{platform: "codex", path: ".codex/skills/codex-auto-reverse/SKILL.md"},
	{platform: "omp", path: ".omp/skills/auto-reverse/SKILL.md", verbatim: true},
	{platform: "opencode", path: ".agents/skills/auto-reverse/SKILL.md", verbatim: true},
}

func generateReverseRouteProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	cfg := config.DefaultFullConfig("reverse-route")
	cfg.Platforms = allPlatforms
	for _, name := range allPlatforms {
		_, err := platformGenerators[name](root, context.Background(), cfg)
		require.NoError(t, err, name)
	}
	return root
}

func readReverseSurface(t *testing.T, root, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	require.NoError(t, err, rel)
	return string(body)
}

// S13: the root routers resolve reverse to exactly one detail and the Claude
// router keeps its byte budget.
func TestReverseRoute_RoutersResolveOneDetail(t *testing.T) {
	t.Parallel()
	root := generateReverseRouteProject(t)

	claudeRouter := readReverseSurface(t, root, ".claude/skills/auto/SKILL.md")
	assert.Equal(t, 1, strings.Count(claudeRouter, ".claude/skills/auto-reverse/SKILL.md"))
	assert.LessOrEqual(t, len([]byte(claudeRouter)), 8192, "Claude root router byte budget")

	for _, surface := range []struct{ router, command, detail string }{
		{
			router:  ".gemini/skills/auto/SKILL.md",
			command: ".gemini/commands/auto/reverse.toml",
			detail:  ".gemini/skills/autopus/auto-reverse/SKILL.md",
		},
		{
			router:  ".agents/plugins/autopus/skills/auto/SKILL.md",
			command: ".agents/plugins/autopus/commands/auto/reverse.toml",
			detail:  ".agents/plugins/autopus/skills/auto-reverse/SKILL.md",
		},
	} {
		router := readReverseSurface(t, root, surface.router)
		assert.Equal(t, 1, strings.Count(router, surface.detail), surface.router)
		assert.Contains(t, readReverseSurface(t, root, surface.command), surface.detail, surface.command)
	}

	assert.Contains(t, readReverseSurface(t, root, ".codex/skills/codex-auto/SKILL.md"),
		"`$codex-auto-doctor`, `$codex-auto-reverse`) before answering or acting.")
	for _, rel := range []string{".omp/skills/auto/SKILL.md", ".omp/commands/auto.md"} {
		assert.Contains(t, readReverseSurface(t, root, rel), "`auto-doctor`, `auto-reverse`.", rel)
	}
	assert.Contains(t, readReverseSurface(t, root, ".omp/commands/auto-reverse.md"),
		"Load exact detail skill `auto-reverse`; this is the same target selected by `/auto reverse ...`.")
	assert.Contains(t, readReverseSurface(t, root, ".agents/skills/auto/SKILL.md"),
		"`auto-doctor`, `auto-reverse`) 중 하나를 로드합니다.")
	assert.Contains(t, readReverseSurface(t, root, ".opencode/commands/auto-reverse.md"),
		"Reinterpret it as `/auto reverse ...` payload")
}

// S13: every auto-reverse detail carries the whole contract, and the platforms
// without an invocation rewrite carry the one shared body byte for byte.
func TestReverseRoute_EveryDetailCarriesTheSharedContract(t *testing.T) {
	t.Parallel()
	root := generateReverseRouteProject(t)
	shared := templates.ReverseRouteContract()

	for _, detail := range reverseDetails {
		body := readReverseSurface(t, root, detail.path)
		assert.Contains(t, body, "name: "+filepath.Base(filepath.Dir(detail.path)), detail.path)
		for _, needle := range reverseContractNeedles {
			assert.Contains(t, body, needle, "%s (%s)", detail.path, detail.platform)
		}
		if detail.verbatim {
			assert.Contains(t, body, shared, "%s must render the shared contract verbatim", detail.path)
		}
		assert.NotContains(t, body, "<!-- autopus:", detail.path)
	}
}

// The Codex detail keeps the main-session gate on Codex syntax: invocations are
// rewritten to Codex entrypoints and no Claude Agent call survives.
func TestReverseRoute_CodexDetailUsesCodexEntrypoints(t *testing.T) {
	t.Parallel()
	root := generateReverseRouteProject(t)
	body := readReverseSurface(t, root, ".codex/skills/codex-auto-reverse/SKILL.md")

	assert.Contains(t, body, "`@auto reverse <대상> <기능> [--runtime]`")
	assert.Contains(t, body, "`$codex-auto-reverse ...`")
	assert.Contains(t, body, "## Codex Multi-Agent V2 Contract")
	assert.NotContains(t, body, "Agent(")
	assert.NotContains(t, body, "`/auto reverse")
}
