package codex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S13 (SPEC-REA-001): the Codex router description ends with /reverse and the
// route registry renders a custom reverse detail instead of a template path.
func TestCodexReverseRoute_RouterDescriptionAndRegistry(t *testing.T) {
	t.Parallel()
	require.Equal(t, "auto", workflowSpecs[0].Name)
	assert.Equal(t,
		"Autopus 명령 라우터 — setup/status/goal/update/plan/go/fix/review/sync/idea/map/why/verify/secure/test/qa/dev/canary/doctor/reverse 서브커맨드를 해석합니다",
		workflowSpecs[0].Description)

	spec, ok := codexWorkflowSpecByName("auto-reverse")
	require.True(t, ok, "auto-reverse must be a registered Codex workflow")
	assert.Empty(t, spec.SkillPath)
	rendered, ok := renderCustomWorkflowSkill(spec)
	require.True(t, ok, "auto-reverse must render a custom workflow body")
	assert.Contains(t, rendered, "name: auto-reverse\n")
	assert.Contains(t, rendered, "Runtime capture evidence: <tool> on <target>: <evidence_id>")
}
