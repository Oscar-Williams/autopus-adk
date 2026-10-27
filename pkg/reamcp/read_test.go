package reamcp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTargetFile(t *testing.T, dir, rel string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))
	return path
}

// A missing file is nil data and a regular file is read up to the cap.
func TestReadTarget_ReadsMissingAndRegularFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rel := TargetRelPath(Antigravity)
	data, err := ReadTarget(dir, rel)
	require.NoError(t, err)
	assert.Nil(t, data)

	full := bytes.Repeat([]byte(" "), TargetLimit)
	writeTargetFile(t, dir, rel, full)
	data, err = ReadTarget(dir, rel)
	require.NoError(t, err)
	assert.Len(t, data, TargetLimit)
}

// An oversized file, a directory, a symlinked file, and a symlinked parent
// directory are refused with their reason and never read.
func TestReadTarget_RefusesUnsafeTargets(t *testing.T) {
	t.Parallel()
	rel := TargetRelPath(Antigravity)

	large := t.TempDir()
	writeTargetFile(t, large, rel, bytes.Repeat([]byte(" "), TargetLimit+1))
	_, err := ReadTarget(large, rel)
	assert.Equal(t, &ReadError{Rel: rel, Reason: ReadTooLarge}, err)

	directory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(directory, filepath.FromSlash(rel)), 0o755))
	_, err = ReadTarget(directory, rel)
	assert.Equal(t, &ReadError{Rel: rel, Reason: ReadIrregular}, err)

	outside := writeTargetFile(t, t.TempDir(), "o.json", []byte(`{"mcpServers": {}}`))
	linked := t.TempDir()
	path := filepath.Join(linked, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	if err := os.Symlink(outside, path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	data, err := ReadTarget(linked, rel)
	assert.Nil(t, data)
	assert.Equal(t, &ReadError{Rel: rel, Reason: ReadSymlink}, err)
	assert.Equal(t, rel+": crosses a symlink", err.Error())

	parent := t.TempDir()
	require.NoError(t, os.Symlink(filepath.Dir(outside), filepath.Join(parent, ".agents")))
	_, err = ReadTarget(parent, rel)
	assert.Equal(t, &ReadError{Rel: rel, Reason: ReadSymlink}, err)
}
