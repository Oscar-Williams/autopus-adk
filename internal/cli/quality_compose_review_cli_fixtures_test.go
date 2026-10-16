package cli

// SPEC-REVMODEL-001 CLI fixtures: RL and its variants on disk, and a form run
// whose PATH holds only the executables a scenario names, so no real omp or
// agy on the machine decides what the tab offers.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/config"
)

// writeReviewRL writes RL (B1): R3 with the code default CLI providers.
func writeReviewRL(t *testing.T, edits ...func(*config.HarnessConfig)) string {
	t.Helper()
	return writeReviewR3(t, append([]func(*config.HarnessConfig){reviewRLConfig}, edits...)...)
}

// reviewRLConfig is RL: R3 with the code default CLI providers, none of them
// OMP-backed.
func reviewRLConfig(cfg *config.HarnessConfig) {
	cfg.Orchestra.Providers = config.DefaultFullConfig("p").Orchestra.Providers
}

// replaceReviewFile edits dir's autopus.yaml text, failing when old is absent.
func replaceReviewFile(t *testing.T, dir, old, replacement string) {
	t.Helper()
	text := readReviewFile(t, dir)
	require.Contains(t, text, old)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(strings.Replace(text, old, replacement, 1)), 0o644))
}

// useReviewPath makes PATH hold exactly empty executables named binaries.
func useReviewPath(t *testing.T, binaries ...string) {
	t.Helper()
	bin := t.TempDir()
	for _, binary := range binaries {
		require.NoError(t, os.WriteFile(filepath.Join(bin, binary), []byte("#!/bin/sh\nexit 0\n"), 0o755))
	}
	t.Setenv("PATH", bin)
}

// runReviewCLIForms runs a real Korean file session over dir with the run's
// Codex catalog CATC (CAT-E), the OMP tab on CATG, FB, and PATH holding only
// binaries.
func runReviewCLIForms(t *testing.T, dir string, binaries []string, lines ...string) reviewRun {
	t.Helper()
	composeSessionCatalog(t, "CAT-E")
	useReviewPath(t, binaries...)
	smoke := useReviewSmokeBackend(t)
	runner := &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATG)}
	deps := ompPlatformDependencies{newRunner: func() omp.OMPModelCatalogRunner { return runner }}
	session := newComposeFormSession(t, dir)
	_, out, err := runComposeForms(t, session, composeFormOptions{OMP: &composeOMPTab{deps: deps}}, lines...)
	return reviewRun{out: out, runner: runner, smoke: smoke, cfg: session.Config, err: err}
}

// reviewArgsLine returns the `args:` line of provider in an autopus.yaml text.
func reviewArgsLine(t *testing.T, text, provider string) string {
	t.Helper()
	at := strings.Index(text, "\n        "+provider+":\n")
	require.GreaterOrEqual(t, at, 0, "no provider %s", provider)
	rest := text[at+1:]
	line := rest[strings.Index(rest, "args: "):]
	return line[:strings.Index(line, "\n")]
}
