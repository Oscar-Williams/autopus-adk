package learn

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MarkReused counts each distinct listed entry once in one rewrite, and an
// unknown ID fails the call without touching the file.
func TestStore_MarkReused(t *testing.T) {
	t.Parallel()

	store, err := NewStore(t.TempDir())
	require.NoError(t, err)
	for _, id := range []string{"L-001", "L-002", "L-003"} {
		require.NoError(t, store.Append(LearningEntry{ID: id, Type: EntryTypeFixPattern}))
	}

	require.NoError(t, store.MarkReused([]string{"L-001", "L-003", "L-001"}))
	before, err := os.ReadFile(store.path)
	require.NoError(t, err)

	assert.Error(t, store.MarkReused([]string{"L-002", "L-999"}))
	after, err := os.ReadFile(store.path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after))

	entries, err := store.Read()
	require.NoError(t, err)
	got := map[string]int{}
	for _, e := range entries {
		got[e.ID] = e.ReuseCount
	}
	assert.Equal(t, map[string]int{"L-001": 1, "L-002": 0, "L-003": 1}, got)
	require.NoError(t, store.MarkReused(nil))
}
