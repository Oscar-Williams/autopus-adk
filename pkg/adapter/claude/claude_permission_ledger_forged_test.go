package claude

// A ledger the checks trust can still be forged: Windows has no mode mark, and
// a checkout under umask 077 writes the repository's files 0600. Update then
// drops the denies such a ledger names, and clean retracts only its allows, so
// a forged ledger cannot weaken the user's denies (v0.50.126 audit M1).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

var forgedLedgerUserDenies = []string{"Read(./.env)", "Bash(curl:*)", "Bash(rm -rf:*)"}

// plantForgedLedger writes a ledger, a manifest claiming it, and a committed
// journal as a checkout of a hostile repository would, all at mode, over a
// settings.json holding one user allow and the user's denies.
func plantForgedLedger(t *testing.T, root string, mode os.FileMode) {
	t.Helper()
	writeJSONFile(t, filepath.Join(root, ".claude", "settings.json"), map[string]any{
		"permissions": map[string]any{"allow": []string{"UserAllow"}, "deny": forgedLedgerUserDenies},
	})
	ledger, err := json.MarshalIndent(claudePermissionLedger{
		Version: claudePermissionLedgerVersion, Allow: []string{"UserAllow"}, Deny: forgedLedgerUserDenies[:2],
	}, "", "  ")
	require.NoError(t, err)
	ledger = append(ledger, '\n')
	sum := adapter.Checksum(string(ledger))
	manifest := adapter.NewManifest(adapterName)
	manifest.Files[claudePermissionLedgerPath] = adapter.ManifestFile{Checksum: sum, Policy: adapter.OverwriteMerge}
	manifestData, err := json.Marshal(manifest)
	require.NoError(t, err)
	journal, err := json.Marshal(adapter.TransactionJournal{
		ID: "forged", Platform: adapterName, Status: adapter.TransactionStatusCommitted, CreatedAt: "2020-01-01T00:00:00Z",
		Entries: []adapter.TransactionJournalEntry{{Path: claudePermissionLedgerPath, Operation: "write", AfterChecksum: sum}},
	})
	require.NoError(t, err)
	txn := filepath.Join(root, ".autopus", "txns", "20200101T000000.000000000-claude-code")
	require.NoError(t, os.MkdirAll(txn, 0o755))
	for path, data := range map[string][]byte{
		filepath.Join(root, filepath.FromSlash(claudePermissionLedgerPath)): ledger,
		filepath.Join(root, ".autopus", adapterName+"-manifest.json"):       manifestData,
		filepath.Join(txn, "journal.json"):                                  journal,
	} {
		require.NoError(t, os.WriteFile(path, data, mode))
		require.NoError(t, os.Chmod(path, mode))
	}
}

func TestPermissionLedger_ForgedTrustedLedgerNeverRemovesADeny(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cleanRoot := t.TempDir()
	plantForgedLedger(t, cleanRoot, 0o600)
	require.NoError(t, NewWithRoot(cleanRoot).Clean(ctx))
	cleaned := readJSONObject(t, filepath.Join(cleanRoot, ".claude", "settings.json"))
	assert.Equal(t, forgedLedgerUserDenies, permissionValues(cleaned, "deny"))
	assert.Empty(t, permissionValues(cleaned, "allow"), "an allow the ledger names may go: that fails safe")

	updateRoot := t.TempDir()
	plantForgedLedger(t, updateRoot, 0o600)
	_, err := NewWithRoot(updateRoot).Update(ctx, config.DefaultFullConfig("ledger-forged"))
	require.NoError(t, err)
	assert.NotContains(t, readJSONObject(t, filepath.Join(updateRoot, filepath.FromSlash(claudePermissionLedgerPath))),
		"deny", "update does not carry the forged denies into the ledger it writes")
	assert.Equal(t, forgedLedgerUserDenies,
		permissionValues(readJSONObject(t, filepath.Join(updateRoot, ".claude", "settings.json")), "deny"))

	windowsRoot := t.TempDir()
	plantForgedLedger(t, windowsRoot, 0o644)
	manifest, err := loadClaudeManifestSafely(windowsRoot)
	require.NoError(t, err)
	ledger, owned, err := loadClaudePermissionLedgerOn("windows", windowsRoot, manifest)
	require.NoError(t, err)
	require.True(t, owned, "Windows trusts the 0644 checkout")
	write, err := NewWithRoot(windowsRoot).prepareCleanSettings(ledger)
	require.NoError(t, err)
	require.NotNil(t, write)
	var settings map[string]any
	require.NoError(t, json.Unmarshal(write.Content, &settings))
	assert.Equal(t, forgedLedgerUserDenies, permissionValues(settings, "deny"))
}

// The retraction itself never touches a deny, whatever the ledger it is given
// names.
func TestRemoveManagedPermissions_RetractsAllowsOnly(t *testing.T) {
	t.Parallel()
	settings := map[string]any{"permissions": map[string]any{
		"allow": []any{"Bash(git:*)", "UserAllow"}, "deny": []any{"Bash(curl:*)", "UserDeny"},
	}}

	changed := removeManagedPermissions(settings, &claudePermissionLedger{
		Version: claudePermissionLedgerVersion, Allow: []string{"Bash(git:*)"}, Deny: []string{"Bash(curl:*)", "UserDeny"},
	})

	assert.True(t, changed)
	assert.Equal(t, []string{"UserAllow"}, permissionValues(settings, "allow"))
	assert.Equal(t, []string{"Bash(curl:*)", "UserDeny"}, permissionValues(settings, "deny"))
}
