package cli

// SPEC-REVMODEL-001 acceptance fixtures shared by the review tab tests: the
// strict catalogs CATG, CATN, and CATE, R3 and its variants on disk, and the
// recording live check backend FB.

import (
	"context"
	"crypto/sha256"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// reviewCatalogModel is one CATG line: an available strict-catalog model.
func reviewCatalogModel(provider, id, family, thinking string) string {
	return `{"provider":"` + provider + `","id":"` + id + `","family":"` + family +
		`","capabilities":["deep_reasoning"],"thinking":[` + thinking +
		`],"auth_enabled":true,"keyless":false,"disabled":false}`
}

var reviewCatalogCATGModels = []string{
	reviewCatalogModel("anthropic", "claude-fable-5-1", "anthropic", `"low","medium","high","max"`),
	reviewCatalogModel("anthropic", "claude-opus-5-5", "anthropic", `"high","max"`),
	reviewCatalogModel("openai-codex", "gpt-6-astra", "openai", `"low","medium","high","xhigh","max"`),
	reviewCatalogModel("openai-codex", "gpt-6-luna", "openai", `"low","medium","high"`),
	reviewCatalogModel("google-antigravity", "gemini-3.1-pro", "google", `"low","high"`),
	reviewCatalogModel("google-antigravity", "gemini-3.8-flash", "google", `"low","medium","high"`),
	reviewCatalogModel("example", "foo-1", "example", `"high"`),
}

func reviewCatalogOf(models ...string) string {
	return `{"models":[` + strings.Join(models, ",") + `]}`
}

var (
	// CATG is the strict catalog every review tab scenario starts from.
	reviewCatalogCATG = reviewCatalogOf(reviewCatalogCATGModels...)
	// CATN is CATG without google-antigravity/gemini-3.1-pro.
	reviewCatalogCATN = reviewCatalogOf(append(append([]string(nil), reviewCatalogCATGModels[:4]...), reviewCatalogCATGModels[5:]...)...)
	// CATE is CATG with example/bar-1 reporting an empty thinking list,
	// which makes the whole catalog catalog_invalid.
	reviewCatalogCATE = reviewCatalogOf(append(append([]string(nil), reviewCatalogCATGModels...),
		`{"provider":"example","id":"bar-1","family":"example","capabilities":["deep_reasoning"],"thinking":[],"auth_enabled":true,"keyless":false,"disabled":false}`)...)
)

// reviewChoicesOf is what installedOMPAttestedModelChoices returns for a
// strict catalog.
func reviewChoicesOf(t *testing.T, catalog string) []ompCustomModelChoice {
	t.Helper()
	return sortedOMPCustomModelChoices(strictOMPCustomModelChoices(mustOMPCatalog(t, []byte(catalog)).Models))
}

// writeReviewProject saves cfg into a fresh temp project.
func writeReviewProject(t *testing.T, cfg *config.HarnessConfig) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, config.Save(dir, cfg))
	return dir
}

// writeReviewR3 saves R3, edited by edits first.
func writeReviewR3(t *testing.T, edits ...func(*config.HarnessConfig)) string {
	t.Helper()
	cfg := reviewR3Config()
	for _, edit := range edits {
		edit(cfg)
	}
	return writeReviewProject(t, cfg)
}

func readReviewFile(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	require.NoError(t, err)
	return string(data)
}

func reviewFileSum(t *testing.T, dir string) [32]byte {
	t.Helper()
	return sha256.Sum256([]byte(readReviewFile(t, dir)))
}

// reviewFileDiff returns the lines of before and after that differ, in
// pairs, for files that keep their line count, and fails otherwise.
func reviewFileDiff(t *testing.T, before, after string) [][2]string {
	t.Helper()
	old, edited := strings.Split(before, "\n"), strings.Split(after, "\n")
	require.Len(t, edited, len(old), "line count changed:\n%s", after)
	var diff [][2]string
	for i := range old {
		if old[i] != edited[i] {
			diff = append(diff, [2]string{old[i], edited[i]})
		}
	}
	return diff
}

// reviewSmokeRequest is one Execute request FB recorded.
type reviewSmokeRequest struct {
	Provider, Prompt, Model string
	Timeout                 time.Duration
}

// reviewSmokeBackend is FB: it records every request and answers the smoke
// marker, or HTTP 404 for gemini on google-antigravity/gemini-3.1-pro:high.
type reviewSmokeBackend struct {
	mu       sync.Mutex
	requests []reviewSmokeRequest
}

func (b *reviewSmokeBackend) Name() string { return "review-smoke-fake" }

func (b *reviewSmokeBackend) Execute(_ context.Context, req orchestra.ProviderRequest) (*orchestra.ProviderResponse, error) {
	b.mu.Lock()
	b.requests = append(b.requests, reviewSmokeRequest{Provider: req.Provider, Prompt: req.Prompt, Model: req.Config.Model, Timeout: req.Timeout})
	b.mu.Unlock()
	if req.Config.Model == "google-antigravity/gemini-3.1-pro:high" {
		return nil, errors.New("HTTP 404 Not Found (Cloud Code Assist)")
	}
	return &orchestra.ProviderResponse{Provider: req.Provider, Output: providerSmokeMarker}, nil
}

func (b *reviewSmokeBackend) recorded() []reviewSmokeRequest {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]reviewSmokeRequest(nil), b.requests...)
}

// useReviewSmokeBackend installs FB as providerSmokeBackendFactory and puts
// empty executables of binaries first on PATH, so the installed check
// passes without running anything. An update run detects platforms from
// PATH, so a test adds only the binaries its providers need.
func useReviewSmokeBackend(t *testing.T, binaries ...string) *reviewSmokeBackend {
	t.Helper()
	backend := &reviewSmokeBackend{}
	original := providerSmokeBackendFactory
	t.Cleanup(func() { providerSmokeBackendFactory = original })
	providerSmokeBackendFactory = func(orchestra.OrchestraConfig) orchestra.ExecutionBackend { return backend }
	bin := t.TempDir()
	for _, binary := range binaries {
		setFakeProviderOnPath(t, bin, binary)
	}
	return backend
}

// reviewUpdate is the environment of a real `auto update` over a review
// fixture: the compose seams, the OMP tab's catalog runner, the OMP
// generation's runner, and FB with an omp stand-in on PATH.
type reviewUpdate struct {
	runner *ompCLIFakeRunner
	smoke  *reviewSmokeBackend
}

// newReviewUpdate prepares it and settles dir with one `auto update --yes`,
// so the file is B0 when it returns; runner nil answers CATG. The project's
// REA decision is stored first: an interactive update asks an undecided
// project before its composition step (SPEC-REA-001), and these answers
// start at the composition menu.
func newReviewUpdate(t *testing.T, dir string, runner *ompCLIFakeRunner) reviewUpdate {
	t.Helper()
	cfg, err := config.Load(dir)
	require.NoError(t, err)
	off := false
	cfg.Features.Rea.Enabled = &off
	require.NoError(t, config.Save(dir, cfg))
	newComposeUpdateRun(t, true, "")
	usePlatformOMPModelRunner(t, &ompModelDoctorE2ERunner{catalog: ompCLIProfileNativeCatalogJSON()})
	out, err := runComposeUpdate(t, strings.NewReader(""), "--dir", dir, "--yes")
	require.NoError(t, err, out)
	useAccessibleComposeForms(t)
	if runner == nil {
		runner = &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATG)}
	}
	useComposeUpdateOMPDeps(t, ompPlatformDependencies{newRunner: func() omp.OMPModelCatalogRunner { return runner }})
	return reviewUpdate{runner: runner, smoke: useReviewSmokeBackend(t, "omp")}
}
