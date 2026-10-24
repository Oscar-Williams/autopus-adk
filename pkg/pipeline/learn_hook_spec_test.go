package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Hook entries used to carry no SPEC ID (except review issues) and no files, so
// neither `auto learn query --spec` nor `--files` could find them. Every hook
// now records the run's SPEC ID, and a review finding records the files it
// cites as path:line.
func TestLearnHooks_RecordSpecIDAndFindingFiles(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	learnHookGateFail(store, "SPEC-X", PhaseID("implement"), GateType("validation"), "", 1)
	learnHookCoverageGap(store, "SPEC-X", "coverage: 40.0% of statements", 85)
	learnHookExecutorError(store, "SPEC-X", PhaseID("implement"), errors.New("boom"))
	learnHookReviewIssue(store, "FINDING: [HIGH] pkg/a/b.go:12 and pkg/a/b.go:30 drop the error; see c_test.go:4", "SPEC-X")

	entries, err := store.Read()
	require.NoError(t, err)
	require.Len(t, entries, 4)
	for _, e := range entries {
		assert.Equal(t, "SPEC-X", e.SpecID, "entry %s (%s)", e.ID, e.Type)
	}
	assert.Equal(t, []string{"pkg/a/b.go", "c_test.go"}, entries[3].Files)
	assert.Empty(t, findingFiles("FINDING text with no location, version 1.2"))
}
