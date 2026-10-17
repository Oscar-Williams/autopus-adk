package claude

// The permission ledger's owner-only mode is its provenance: the ownership
// check trusts only a ledger an Autopus write left 0600. A clone checks the
// committed ledger out at 0644, so the update that claims it must tighten it
// (v0.50.124–v0.50.125 kept the 0644 and the next update refused the ledger
// the run itself had written).

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
)

// A fresh clone: no manifest and no transactions (both gitignored) and the
// committed ledger checked out at 0644 with whatever content the repository
// holds. The claiming update writes its own ledger at 0600, the next update
// trusts it, and clean keeps the user's own entry the checked-out ledger named.
func TestUpdate_ClaimsACheckedOutLedgerOwnerOnly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cfg := config.DefaultFullConfig("ledger-clone")
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)
	require.NoError(t, os.Remove(filepath.Join(root, ".autopus", adapterName+"-manifest.json")))
	require.NoError(t, os.RemoveAll(filepath.Join(root, ".autopus", "txns")))
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	writeJSONFile(t, settingsPath, map[string]any{"permissions": map[string]any{"deny": []string{"UserDeny"}}})
	ledgerPath := filepath.Join(root, filepath.FromSlash(claudePermissionLedgerPath))
	require.NoError(t, os.WriteFile(ledgerPath, []byte(`{"version":1,"deny":["UserDeny"]}`+"\n"), 0o644))
	require.NoError(t, os.Chmod(ledgerPath, 0o644))

	_, err = a.Update(context.Background(), cfg)

	require.NoError(t, err)
	info, err := os.Stat(ledgerPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "the claiming write tightens the checked-out ledger")
	data, err := os.ReadFile(ledgerPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "UserDeny", "the claim writes its own ledger, never the checked-out one")
	_, err = a.Update(context.Background(), cfg)
	require.NoError(t, err, "the next update trusts the ledger the claim wrote")
	require.NoError(t, a.Clean(context.Background()))
	assert.Equal(t, []string{"UserDeny"}, permissionValues(readJSONObject(t, settingsPath), "deny"))
}

// A ledger the manifest already claims but that lost its owner-only mode, as
// v0.50.124–v0.50.125 left a claimed checkout, stays refused before anything
// is written, and the error names the file and the recovery; chmod 600
// restores trust.
func TestUpdate_RefusesAClaimedLedgerThatLostItsMode(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cfg := config.DefaultFullConfig("ledger-claimed")
	a := NewWithRoot(root)
	_, err := a.Generate(context.Background(), cfg)
	require.NoError(t, err)
	ledgerPath := filepath.Join(root, filepath.FromSlash(claudePermissionLedgerPath))
	require.NoError(t, os.Chmod(ledgerPath, 0o644))
	before, err := os.ReadFile(ledgerPath)
	require.NoError(t, err)

	_, err = a.Update(context.Background(), cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission ledger must be an owner-only regular file: "+
		".autopus/claude-code-permissions.json has mode 0644, not 0600; "+
		"if your own auto update wrote it, review its allow and deny lists and chmod 600 it")
	assertFileContent(t, ledgerPath, before)
	info, err := os.Stat(ledgerPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

	require.NoError(t, os.Chmod(ledgerPath, 0o600))
	_, err = a.Update(context.Background(), cfg)
	require.NoError(t, err)
}
