package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/learn"
)

// The query used to print 1.00 for every match and never counted a reuse, so
// the summary's top patterns were always 0x. It now prints the score that
// ranked each entry, --top keeps the best N, and --mark-reused counts exactly
// the printed entries.
func TestCLI_LearnQuery_PrintsScoresTopAndMarksReuse(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(cwd) }()
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))

	learningsDir := filepath.Join(tempDir, ".autopus", "learnings")
	require.NoError(t, os.MkdirAll(learningsDir, 0o755))
	// Timestamps are old enough that no entry earns the recency bonus.
	raw := `{"id":"L-001","timestamp":"2026-01-01T00:00:00Z","files":["a.go"],"pattern":"nil map write"}
{"id":"L-002","timestamp":"2026-01-02T00:00:00Z","pattern":"nil map in cache"}
{"id":"L-003","timestamp":"2026-01-03T00:00:00Z","files":["b.go"],"pattern":"unrelated"}
`
	require.NoError(t, os.WriteFile(filepath.Join(learningsDir, "pipeline.jsonl"), []byte(raw), 0o644))

	run := func(args ...string) string {
		cmd := newLearnQueryCmd()
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
		return stdout.String()
	}

	out := run("--files", "a.go", "--keywords", "nil map")
	assert.Regexp(t, `L-001\s+\S*\s+4\.00\s+nil map write`, out, "file match 3 + keyword 1")
	assert.Regexp(t, `L-002\s+\S*\s+1\.00\s+nil map in cache`, out, "keyword only")
	assert.NotContains(t, out, "L-003")
	assert.Less(t, strings.Index(out, "L-001"), strings.Index(out, "L-002"), "highest score first")

	out = run("--files", "a.go", "--keywords", "nil map", "--top", "1", "--mark-reused")
	assert.Contains(t, out, "L-001")
	assert.NotContains(t, out, "L-002")

	store, err := learn.NewStore(tempDir)
	require.NoError(t, err)
	entries, err := store.Read()
	require.NoError(t, err)
	reuse := map[string]int{}
	for _, e := range entries {
		reuse[e.ID] = e.ReuseCount
	}
	assert.Equal(t, map[string]int{"L-001": 1, "L-002": 0, "L-003": 0}, reuse)
}
