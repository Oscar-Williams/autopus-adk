package cli

// The fail-closed edges of saveReaKey: the errors around the splice, and the
// whole-document check that stands behind it.

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// A missing file, a retired key the prune refuses, and a failed replace each
// end the write with their error and leave autopus.yaml and cfg as they were.
func TestSaveReaKey_FailuresWriteNothing(t *testing.T) {
	missing := t.TempDir()
	cfg := config.DefaultFullConfig("p")
	require.ErrorContains(t, saveReaKey(missing, cfg, false), "read config")
	assert.NoFileExists(t, filepath.Join(missing, "autopus.yaml"))
	assert.Nil(t, cfg.Features.Rea.Enabled)

	anchored := reaSaveHead + "features:\n  cc21:\n    monitor_pattern_timeout_ms: &m 5000\n"
	dir, cfg := reaSaveProject(t, anchored)
	got, err := reaSaved(t, dir, cfg, false)
	require.ErrorIs(t, err, config.ErrRetiredKeyAnchor)
	assert.Equal(t, anchored, got)
	assert.Nil(t, cfg.Features.Rea.Enabled)

	original := renameQualityConfig
	t.Cleanup(func() { renameQualityConfig = original })
	renameQualityConfig = func(string, string) error { return errors.New("rename failed") }
	dir, cfg = reaSaveProject(t, reaSaveHead)
	got, err = reaSaved(t, dir, cfg, true)
	require.EqualError(t, err, "write config: rename failed")
	assert.Equal(t, reaSaveHead, got)
	assert.Nil(t, cfg.Features.Rea.Enabled)
}

// A document whose root is not one block mapping with unique keys is never
// spliced.
func TestSpliceReaKey_NeedsOneRootMapping(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"mode: full\nmode: full\n", "- mode\n", "{mode: full}\n", ""} {
		_, ok := spliceReaKey([]byte(source), false)
		assert.False(t, ok, "%q", source)
	}
}

// verifyReaKeyEdit passes only an edit that adds or changes the key alone. It
// refuses an edit that does not parse, that changes another value, that
// stores a string, that does not load, or that loads to other values, as a
// root features entry does when it overrides the one a merge key supplied. A
// source that does not load is an error, not a refusal.
func TestVerifyReaKeyEdit_RefusesEveryOtherChange(t *testing.T) {
	t.Parallel()
	const source = "mode: full\nproject_name: p\nplatforms: [claude-code]\n"
	const block = "features:\n  rea:\n    enabled: false\n"
	const merged = source + "future_extension: &b\n  features:\n    cc21:\n      enabled: true\n<<: *b\n"
	require.NoError(t, verifyReaKeyEdit([]byte(source), []byte(source+block), false))
	refused := map[string][2]string{
		"edited does not parse":        {source, source + block + "]\n"},
		"another value changed":        {source, "mode: full\nproject_name: q\nplatforms: [claude-code]\n" + block},
		"the key holds a string":       {source, source + "features:\n  rea:\n    enabled: \"false\"\n"},
		"edited does not load":         {source, source + block + "  bogus: 1\n"},
		"a merged entry is overridden": {merged, merged + block},
	}
	for name, pair := range refused {
		var refusal reaKeyRefusal
		err := verifyReaKeyEdit([]byte(pair[0]), []byte(pair[1]), false)
		assert.True(t, errors.As(err, &refusal), "%s: %v", name, err)
	}
	err := verifyReaKeyEdit([]byte(source+"bogus: 1\n"), []byte(source+"bogus: 1\n"+block), false)
	assert.ErrorContains(t, err, "load config")
}
