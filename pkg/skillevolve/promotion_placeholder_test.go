package skillevolve

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The generator writes a stub, not a skill edit, and the stub passes the static
// safety gate. Promoting it unchanged used to overwrite the real skill with the
// stub, so promotion refuses it, even in a dry run, and leaves the skill as is.
func TestPromoteCandidate_RefusesThePlaceholderTheGeneratorWrites(t *testing.T) {
	t.Parallel()

	sourceRel := "autopus-adk/content/skills/testing-strategy.md"
	stub := proposedFilesFor("fp-1", []string{sourceRel}, []string{"AC-SEVOLVE-003"})
	require.Len(t, stub, 1)

	safety, err := EvaluateSafety(context.Background(), promotionReadyCandidate(stub[0].Content), SafetyOptions{})
	require.NoError(t, err)
	require.True(t, safety.PromotionAllowed, "the stub must reach the placeholder check, not stop at safety")

	projectDir := t.TempDir()
	original := validSkillContent("testing-strategy") + "real guidance\n"
	writeWorkspaceFile(t, projectDir, sourceRel, original)

	for _, apply := range []bool{false, true} {
		result, err := PromoteCandidate(context.Background(), PromotionOptions{
			ProjectDir: projectDir,
			Candidate:  promotionReadyCandidate(stub[0].Content),
			Approval:   HumanApproval{ApprovedBy: "human-reviewer"},
			Apply:      apply,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "placeholder content")
		assert.False(t, result.Applied)
	}
	assert.Equal(t, original, readWorkspaceFiles(t, projectDir, []string{sourceRel})[sourceRel])
}
