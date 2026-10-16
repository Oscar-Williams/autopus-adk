package cli

// SPEC-TUIUX-001 acceptance environment: every fixture is a project named p
// under a temp parent, HOME and XDG live in temp dirs, PATH holds no provider
// CLI, the REA Node probe passes (TestMain), and the run's Codex catalog
// probe answers an acceptance catalog. NPD is the normalized project digest
// of acceptance.md.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/internal/cli/tui"
	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

// newTUIUXEnv isolates one run: the TTY and CI seams, Korean text, the
// environment of acceptance.md, and the run's catalog probe on CAT-A.
func newTUIUXEnv(t *testing.T, tty bool, ci string) *[]composeProbeCall {
	t.Helper()
	stubDriverProbes(t, tty, ci)
	originalLanguage := composeLanguage
	t.Cleanup(func() { composeLanguage = originalLanguage })
	composeLanguage = func() string { return "ko" }
	for _, key := range []string{"HOME", "CODEX_HOME", "XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_STATE_HOME", "PATH"} {
		t.Setenv(key, t.TempDir())
	}
	t.Setenv("LANG", "ko_KR.UTF-8")
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANGUAGE"} {
		t.Setenv(key, "")
	}
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("CI", ci)
	return composeSessionCatalog(t, "CAT-A")
}

// tuiuxProject makes the empty project directory p under a temp parent.
func tuiuxProject(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "p")
	require.NoError(t, os.Mkdir(dir, 0o755))
	return dir
}

// tuiuxFixtureD writes D into a fresh project p: the saved code defaults on
// Claude Code and Codex with supervisor inherit.
func tuiuxFixtureD(t *testing.T) string {
	t.Helper()
	dir := tuiuxProject(t)
	cfg := config.DefaultFullConfig("p")
	cfg.Platforms = []string{"claude-code", "codex"}
	cfg.Quality.SupervisorModelPolicy = config.SupervisorModelPolicyInherit
	require.NoError(t, config.Save(dir, cfg))
	return dir
}

// tuiuxW is the wizard answer W of acceptance.md.
func tuiuxW() tui.InitWizardResult {
	rea := true
	return tui.InitWizardResult{
		UsageProfile: "developer", CommentsLang: "en", CommitsLang: "ko", AILang: "ko",
		Quality: "balanced", ReviewGate: false, Rea: &rea, Methodology: "tdd",
	}
}

// stubTUIUXWizard replaces the wizard with a seam that records every call
// and answers the results in turn, the last one again once they run out.
func stubTUIUXWizard(t *testing.T, results ...tui.InitWizardResult) *[]tui.InitWizardOpts {
	t.Helper()
	original := runInitWizard
	t.Cleanup(func() { runInitWizard = original })
	calls := &[]tui.InitWizardOpts{}
	runInitWizard = func(opts tui.InitWizardOpts) (*tui.InitWizardResult, error) {
		*calls = append(*calls, opts)
		result := results[min(len(*calls), len(results))-1]
		if result.Rea != nil {
			rea := *result.Rea
			result.Rea = &rea
		}
		return &result, nil
	}
	return calls
}

// stubTUIUXFormsProbe fixes the forms decision without touching the runner.
func stubTUIUXFormsProbe(t *testing.T, forms bool) {
	t.Helper()
	original := composeFormsProbe
	t.Cleanup(func() { composeFormsProbe = original })
	composeFormsProbe = func(*cobra.Command) bool { return forms }
}

// runTUIUXCommand runs the real root command with args and in, both output
// streams in one buffer, and returns it with dir written <DIR>. A nil in
// leaves the process stdin as the command input.
func runTUIUXCommand(t *testing.T, dir string, in io.Reader, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	if in != nil {
		root.SetIn(in)
	}
	root.SetArgs(args)
	err := root.Execute()
	return strings.ReplaceAll(out.String(), dir, "<DIR>"), err
}

// runTUIUXInit runs `auto init --dir <DIR> --platforms claude-code,codex`.
func runTUIUXInit(t *testing.T, dir string, in io.Reader, args ...string) (string, error) {
	t.Helper()
	return runTUIUXCommand(t, dir, in, append([]string{"init", "--dir", dir, "--platforms", "claude-code,codex"}, args...)...)
}

// The NPD normalization (research.md ### NPD 정규화): the manifest
// generated_at value, the backup timestamp directory with or without its
// transaction fraction, and the adapter transaction journal directory.
var (
	tuiuxGeneratedAt = regexp.MustCompile(`"generated_at": "[^"]*"`)
	tuiuxBackupStamp = regexp.MustCompile(`\.autopus/backup/[0-9]{8}T[0-9]{6}(\.[0-9]+)?/`)
	tuiuxTxnStamp    = regexp.MustCompile(`^\.autopus/txns/[0-9]{8}T[0-9]{6}\.[0-9]+-`)
)

// tuiuxJournal compares an adapter transaction journal by its decoded
// values without the transaction time: the id, created_at, the timestamp in
// each backup path, and the checksum of a manifest entry, whose bytes hold
// generated_at.
func tuiuxJournal(t *testing.T, data []byte) []byte {
	t.Helper()
	var journal adapter.TransactionJournal
	require.NoError(t, json.Unmarshal(data, &journal))
	journal.ID, journal.CreatedAt = "<TX>", "<T>"
	for i, entry := range journal.Entries {
		entry.BackupPath = tuiuxBackupStamp.ReplaceAllString(entry.BackupPath, ".autopus/backup/<TS>/")
		if strings.HasSuffix(entry.Path, "-manifest.json") && entry.AfterChecksum != "" {
			entry.AfterChecksum = "<T>"
		}
		journal.Entries[i] = entry
	}
	encoded, err := json.Marshal(journal)
	require.NoError(t, err)
	return encoded
}

// tuiuxNPD is the normalized project digest: every path under dir with the
// SHA-256 of its bytes, a symlink with its target, after the normalization
// above turns each time value into <T>, <TS>, or <TX>.
func tuiuxNPD(t *testing.T, dir string) string {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || path == dir {
			return walkErr
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			rel += "/"
		}
		rel = tuiuxBackupStamp.ReplaceAllString(rel, ".autopus/backup/<TS>/")
		rel = tuiuxTxnStamp.ReplaceAllString(rel, ".autopus/txns/<TX>-")
		switch {
		case entry.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			lines = append(lines, rel+" -> "+target)
		case entry.IsDir():
			lines = append(lines, rel)
		default:
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			data = tuiuxGeneratedAt.ReplaceAll(data, []byte(`"generated_at": "<T>"`))
			if strings.HasPrefix(rel, ".autopus/txns/") && strings.HasSuffix(rel, "/journal.json") {
				data = tuiuxJournal(t, data)
			}
			sum := sha256.Sum256(data)
			lines = append(lines, rel+" "+hex.EncodeToString(sum[:]))
		}
		return nil
	})
	require.NoError(t, err)
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}
