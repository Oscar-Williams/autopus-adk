package claude

// The ledger checks per platform, and a claimed ledger that is gone. Windows
// runs only here, through the injected platform: Go reports 0666 or 0444 for
// every Windows file, so the owner-only mode cannot mark provenance there.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

func TestPermissionLedgerModeTrusted_FollowsThePlatform(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		goos string
		perm os.FileMode
		want bool
	}{
		{"linux", 0o600, true},
		{"darwin", 0o600, true},
		{"linux", 0o644, false},
		{"darwin", 0o640, false},
		{"linux", 0o400, false},
		{"windows", 0o666, true},
		{"windows", 0o444, true},
	} {
		assert.Equal(t, tc.want, permissionLedgerModeTrusted(tc.goos, tc.perm), "%s %04o", tc.goos, tc.perm)
	}
}

// generatedLedger returns a project whose ledger an Autopus write claimed, and
// the ledger's path.
func generatedLedger(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	_, err := NewWithRoot(root).Generate(context.Background(), config.DefaultFullConfig("ledger-platform"))
	require.NoError(t, err)
	return root, filepath.Join(root, filepath.FromSlash(claudePermissionLedgerPath))
}

func loadLedgerOn(t *testing.T, goos, root string) (bool, error) {
	t.Helper()
	manifest, err := loadClaudeManifestSafely(root)
	require.NoError(t, err)
	_, owned, err := loadClaudePermissionLedgerOn(goos, root, manifest)
	return owned, err
}

// On Windows the mode check gives way and every other check stays: the
// manifest checksum, the committed transaction, and the regular-file and
// no-symlink rules. Unix keeps refusing any mode but 0600.
func TestLoadPermissionLedgerOn_WindowsDropsOnlyTheModeCheck(t *testing.T) {
	t.Parallel()
	root, ledger := generatedLedger(t)
	require.NoError(t, os.Chmod(ledger, 0o644))

	owned, err := loadLedgerOn(t, "windows", root)
	require.NoError(t, err)
	assert.True(t, owned)
	_, err = loadLedgerOn(t, "linux", root)
	assert.ErrorContains(t, err, "has mode 0644, not 0600")

	data, err := os.ReadFile(ledger)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(ledger, append(data, ' '), 0o644))
	_, err = loadLedgerOn(t, "windows", root)
	assert.ErrorContains(t, err, "permission ledger checksum mismatch")
	require.NoError(t, os.WriteFile(ledger, data, 0o644))

	require.NoError(t, os.RemoveAll(filepath.Join(root, ".autopus", "txns")))
	_, err = loadLedgerOn(t, "windows", root)
	assert.ErrorContains(t, err, "no committed transaction ownership")

	dirRoot, dirLedger := generatedLedger(t)
	require.NoError(t, os.Remove(dirLedger))
	require.NoError(t, os.Mkdir(dirLedger, 0o700))
	_, err = loadLedgerOn(t, "windows", dirRoot)
	assert.Error(t, err, "a directory is never a ledger")

	linkRoot, linkLedger := generatedLedger(t)
	victim := filepath.Join(t.TempDir(), "victim.json")
	require.NoError(t, os.Rename(linkLedger, victim))
	mustSymlink(t, victim, linkLedger)
	_, err = loadLedgerOn(t, "windows", linkRoot)
	assert.ErrorContains(t, err, "crosses symlink")
}

// A claimed ledger that is gone, as git deletes it from a clone that pulls the
// commit untracking it, no longer stops update or clean: update starts a fresh
// owner-only ledger, which records none of the entries .claude/settings.json
// already held, and clean leaves every permission entry.
func TestPermissionLedger_ClaimedButGoneReadsAsNone(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultFullConfig("ledger-gone")
	root, ledger := generatedLedger(t)
	require.NoError(t, os.Remove(ledger))
	manifest, err := adapter.LoadManifest(root, adapterName)
	require.NoError(t, err)
	require.Contains(t, manifest.Files, claudePermissionLedgerPath, "the manifest still claims the ledger")

	_, err = NewWithRoot(root).Update(context.Background(), cfg)

	require.NoError(t, err)
	info, err := os.Stat(ledger)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	fresh := readJSONObject(t, ledger)
	assert.NotContains(t, fresh, "allow", "the entries settings.json already held stay unclaimed")
	assert.NotContains(t, fresh, "deny")

	cleanRoot, cleanLedger := generatedLedger(t)
	settingsPath := filepath.Join(cleanRoot, ".claude", "settings.json")
	before := readJSONObject(t, settingsPath)
	require.NoError(t, os.Remove(cleanLedger))
	require.NoError(t, NewWithRoot(cleanRoot).Clean(context.Background()))
	after := readJSONObject(t, settingsPath)
	assert.Equal(t, permissionValues(before, "allow"), permissionValues(after, "allow"))
	assert.Equal(t, permissionValues(before, "deny"), permissionValues(after, "deny"))
}

// A deleted ledger the manifest claims is written again; any other deleted
// managed file stays deleted, and a present file resolves as before.
func TestResolveUpdateAction_WritesADeletedLedgerAgain(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	manifest := adapter.NewManifest(adapterName)
	manifest.Files[claudePermissionLedgerPath] = adapter.ManifestFile{Checksum: "x", Policy: adapter.OverwriteMerge}
	manifest.Files["CLAUDE.md"] = adapter.ManifestFile{Checksum: "y", Policy: adapter.OverwriteMerge}
	ledger := adapter.FileMapping{TargetPath: claudePermissionLedgerPath, OverwritePolicy: adapter.OverwriteMerge}
	other := adapter.FileMapping{TargetPath: "CLAUDE.md", OverwritePolicy: adapter.OverwriteMerge}

	assert.Equal(t, adapter.ActionCreate, ResolveUpdateAction(root, ledger, manifest))
	assert.Equal(t, adapter.ActionSkip, ResolveUpdateAction(root, other, manifest))
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".autopus"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, filepath.FromSlash(claudePermissionLedgerPath)), []byte("{}"), 0o600))
	assert.Equal(t, adapter.ResolveAction(root, ledger.TargetPath, ledger.OverwritePolicy, manifest),
		ResolveUpdateAction(root, ledger, manifest))
}
