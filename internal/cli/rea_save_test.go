package cli

// SPEC-REA-001 Decision State Contract, the update write: saveReaKey edits
// only features.rea.enabled in the existing autopus.yaml bytes, the way the
// review tab's line splice writer edits only its spans (SPEC-REVMODEL-001).

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// reaSaveHead is a loadable autopus.yaml head in a layout yaml.v3 never
// writes: a comment line, an inline comment, and three-space indentation.
const reaSaveHead = "# hand-edited\nmode: full  # keep\nproject_name: p\nplatforms:\n   - claude-code\n" +
	"language:\n   comments: en\n"

// reaSaveProject writes text as dir's autopus.yaml with mode 0640 and returns
// dir and the config text loads to.
func reaSaveProject(t *testing.T, text string) (string, *config.HarnessConfig) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(text), 0o640))
	cfg, err := config.ParseConfig([]byte(text))
	require.NoError(t, err)
	return dir, cfg
}

// reaSaved runs saveReaKey and returns the autopus.yaml it leaves.
func reaSaved(t *testing.T, dir string, cfg *config.HarnessConfig, enabled bool) (string, error) {
	t.Helper()
	err := saveReaKey(dir, cfg, enabled)
	data, readErr := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	require.NoError(t, readErr)
	return string(data), err
}

// The RC's own autopus.yaml (no features key, `off` unquoted, providers
// without binary or args, omp_context_policy before language) gains exactly
// the three key lines at its end; config.Save moved blocks and added empty
// values there (smoke 2026-10-17).
func TestSaveReaKey_RCLayoutGainsOnlyTheKeyLines(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join(composeTableTestdata, "R.yaml"))
	require.NoError(t, err)
	dir, cfg := reaSaveProject(t, string(data))

	got, err := reaSaved(t, dir, cfg, false)

	require.NoError(t, err)
	assert.Equal(t, string(data)+"features:\n    rea:\n        enabled: false\n", got)
	assert.Equal(t, reaBool(false), cfg.Features.Rea.Enabled)
	reloaded, err := config.Load(dir)
	require.NoError(t, err)
	assert.Equal(t, reaBool(false), reloaded.Features.Rea.Enabled)
}

func TestSaveReaKey_EditsOnlyTheKey(t *testing.T) {
	t.Parallel()
	const tail = "workflow:\n   coverage_threshold: 85\n"
	const cc21 = "features:\n  cc21:\n    enabled: false\n"
	cases := []struct {
		name, before, after string
		enabled             bool
	}{{
		name:   "no features: the block follows the last root entry, one file step per level",
		before: reaSaveHead + tail,
		after:  reaSaveHead + tail + "features:\n   rea:\n      enabled: false\n",
	}, {
		name:    "features without rea: rea closes the features block, at its own step",
		before:  reaSaveHead + "features:\n  cc21:\n    enabled: false  # cc21\n    task_created_mode: 'warn'\n\n# tail\n" + tail,
		after:   reaSaveHead + "features:\n  cc21:\n    enabled: false  # cc21\n    task_created_mode: 'warn'\n  rea:\n    enabled: true\n\n# tail\n" + tail,
		enabled: true,
	}, {
		name:   "a stored value is replaced in place, its tag and comment kept",
		before: reaSaveHead + "features:\n    rea:\n        enabled: !!bool true # decided by hand\n" + tail,
		after:  reaSaveHead + "features:\n    rea:\n        enabled: !!bool false # decided by hand\n" + tail,
	}, {
		name:   "yes reads as true and is replaced",
		before: reaSaveHead + "features:\n  rea:\n    enabled: yes\n",
		after:  reaSaveHead + "features:\n  rea:\n    enabled: false\n",
	}, {
		name:   "a null value takes the value",
		before: reaSaveHead + "features:\n  rea:\n    enabled:\n" + tail,
		after:  reaSaveHead + "features:\n  rea:\n    enabled: false\n" + tail,
	}, {
		name:   "CRLF lines stay CRLF and the added lines use CRLF",
		before: strings.ReplaceAll(reaSaveHead+cc21+tail, "\n", "\r\n"),
		after:  strings.ReplaceAll(reaSaveHead+cc21+"  rea:\n    enabled: false\n"+tail, "\n", "\r\n"),
	}, {
		name:   "a file without a final newline gets one before the block",
		before: reaSaveHead + strings.TrimSuffix(tail, "\n"),
		after:  reaSaveHead + tail + "features:\n   rea:\n      enabled: false\n",
	}, {
		name:   "a retired key never rides along",
		before: reaSaveHead + cc21 + "    monitor_pattern_timeout_ms: 5000\n" + tail,
		after:  reaSaveHead + cc21 + "  rea:\n    enabled: false\n" + tail,
	}, {
		name:   "a file without a nested mapping indents by two spaces",
		before: "mode: full\nproject_name: p\nplatforms:\n  - claude-code\n",
		after:  "mode: full\nproject_name: p\nplatforms:\n  - claude-code\nfeatures:\n  rea:\n    enabled: false\n",
	}}
	for _, tc := range cases {
		dir, cfg := reaSaveProject(t, tc.before)

		got, err := reaSaved(t, dir, cfg, tc.enabled)

		require.NoError(t, err, tc.name)
		assert.Equal(t, tc.after, got, tc.name)
		assert.Equal(t, reaBool(tc.enabled), cfg.Features.Rea.Enabled, tc.name)
		info, statErr := os.Stat(filepath.Join(dir, "autopus.yaml"))
		require.NoError(t, statErr)
		assert.Equal(t, os.FileMode(0o640), info.Mode().Perm(), tc.name)
	}
}

// What the splice cannot edit safely is refused: nothing is written, cfg
// keeps its value, and no whole-config save stands in for the edit.
func TestSaveReaKey_RefusesWhatItCannotEditSafely(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"flow features":         reaSaveHead + "features: {cc21: {enabled: false}}\n",
		"flow rea":              reaSaveHead + "features:\n  rea: {enabled: true}\n",
		"empty flow rea":        reaSaveHead + "features:\n  rea: {}\n",
		"null features":         reaSaveHead + "features:\nworkflow:\n   coverage_threshold: 85\n",
		"aliased features":      reaSaveHead + "features: &f\n  cc21:\n    enabled: false\nfuture_extension: *f\n",
		"anchored value":        reaSaveHead + "features:\n  rea:\n    enabled: &e true\n",
		"merge key":             reaSaveHead + "future_extension: &base\n  cc21:\n    enabled: false\nfeatures:\n  <<: *base\n",
		"lone CR line break":    strings.Replace(reaSaveHead, "project_name: p\n", "project_name: p\r\r\n", 1),
		"split block scalar":    reaSaveHead + "future_extension: |\n  text\n\n  more\n",
		"alias at the target":   reaSaveHead + "future_extension: &t true\nfeatures:\n  rea:\n    enabled: *t\n",
		"value on its own line": reaSaveHead + "features:\n  rea:\n    enabled:\n      true\n",
	}
	for name, text := range cases {
		dir, cfg := reaSaveProject(t, text)
		held := cfg.Features.Rea.Enabled

		got, err := reaSaved(t, dir, cfg, false)

		var refusal reaKeyRefusal
		require.True(t, errors.As(err, &refusal), "%s: %v", name, err)
		assert.Equal(t, text, got, name)
		assert.Equal(t, held, cfg.Features.Rea.Enabled, name)
	}
	assert.EqualError(t, reaKeyRefusal{enabled: true}, "refuse to edit features.rea.enabled in autopus.yaml: "+
		"the key cannot be changed as plain lines alone; nothing was written, set features.rea.enabled: true by hand")
}

// Init writes autopus.yaml whole from cfg, so its key write saves cfg again
// even where the splice refuses; an update run over the same file refuses and
// fails its finish with the refusal.
func TestReaRunSaveKey_InitSavesWholeAndUpdateSplices(t *testing.T) {
	t.Parallel()
	refused := strings.Replace(reaSaveHead, "project_name: p\n", "project_name: p\r\r\n", 1)

	initDir, initCfg := reaSaveProject(t, refused)
	initRun := prepareInitRea(initDir, initCfg, false, true)
	var out bytes.Buffer
	initRun.decide(context.Background(), &out, initCfg, reaOff)
	require.NoError(t, initRun.finish(&out, initDir, initCfg))
	reloaded, err := config.Load(initDir)
	require.NoError(t, err)
	assert.Equal(t, reaBool(false), reloaded.Features.Rea.Enabled)

	updateDir, updateCfg := reaSaveProject(t, refused)
	updateRun, _ := testReaRun(t, updateDir, updateCfg, "")
	out.Reset()
	updateRun.decide(context.Background(), &out, updateCfg, reaOff)
	err = updateRun.finish(&out, updateDir, updateCfg)
	var refusal reaKeyRefusal
	require.True(t, errors.As(err, &refusal), "%v", err)
	assert.Contains(t, err.Error(), "features.rea 저장 실패: refuse to edit features.rea.enabled")
	data, readErr := os.ReadFile(filepath.Join(updateDir, "autopus.yaml"))
	require.NoError(t, readErr)
	assert.Equal(t, refused, string(data))
	assert.Nil(t, updateCfg.Features.Rea.Enabled)
	assert.NotContains(t, out.String(), reaTextKo.OldBinary, "a refused key prints no old-binary line")
}
