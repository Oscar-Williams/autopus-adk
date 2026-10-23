package companionmanifest

import (
	"regexp"
	"testing"
)

// The release Go toolchain is pinned in six places that cannot be derived from
// one another at build time: go.mod (which every setup-go step reads through
// go-version-file), the GoReleaser build env, the two release scripts' shell
// literals (run under `env -i` with GOENV=off), and the two tests that arm the
// GoReleaser and release-prep literals.
//
// The GoReleaser post-hook requires the release binary to share a digest with
// the candidate build, so every site has to name the same toolchain. Of the
// four production sites only the GoReleaser env and prepare-release.sh were
// armed by a test. go.mod and the candidate builder could be left behind by a
// partial bump, and nothing failed until the release lane was already running,
// which is far too late.
//
// go.mod is the reference: it is the only site CI reads implicitly, so a
// disagreement means a release would be built by a different Go than the one
// the pull request was verified with.
type goToolchainPinSite struct {
	name    string
	file    string
	pattern string
}

func goToolchainPinSites() []goToolchainPinSite {
	return []goToolchainPinSite{
		{
			name:    "go.mod toolchain line",
			file:    "go.mod",
			pattern: `(?m)^toolchain (go[0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			name:    "GoReleaser auto build env",
			file:    ".goreleaser.yaml",
			pattern: `(?m)^\s*- GOTOOLCHAIN=(go[0-9]+\.[0-9]+\.[0-9]+)$`,
		},
		{
			name:    "release-prep expected toolchain",
			file:    "scripts/companion-release/prepare-release.sh",
			pattern: `(?m)^readonly expected_go_toolchain='(go[0-9]+\.[0-9]+\.[0-9]+)'$`,
		},
		{
			name:    "candidate build expected toolchain",
			file:    "scripts/companion-release/build-omp-context-candidate.sh",
			pattern: `(?m)^readonly expected_go_toolchain='(go[0-9]+\.[0-9]+\.[0-9]+)'$`,
		},
		{
			name:    "GoReleaser oracle pin",
			file:    "internal/companionmanifest/release_goreleaser_config_oracle_test.go",
			pattern: `"GOTOOLCHAIN=(go[0-9]+\.[0-9]+\.[0-9]+)"`,
		},
		{
			name:    "release-prep hardening pin",
			file:    "scripts/companion-release/tests/release-prep-hardening-test.sh",
			pattern: `expected_go_toolchain='(go[0-9]+\.[0-9]+\.[0-9]+)'`,
		},
	}
}

func TestGoToolchainPinAgreesEverywhere(t *testing.T) {
	t.Parallel()

	declared := ""
	for _, site := range goToolchainPinSites() {
		body := readReleaseFile(t, site.file)
		matches := regexp.MustCompile(site.pattern).FindAllStringSubmatch(body, -1)
		if len(matches) != 1 {
			t.Fatalf("%s (%s) states a Go toolchain %d time(s) matching %s, want exactly 1",
				site.name, site.file, len(matches), site.pattern)
		}
		if declared == "" {
			declared = matches[0][1]
			continue
		}
		if matches[0][1] != declared {
			t.Fatalf("%s pins %s while go.mod pins %s — every Go toolchain pin must move "+
				"together: go.mod, .goreleaser.yaml, prepare-release.sh, "+
				"build-omp-context-candidate.sh and the two tests that arm them",
				site.name, matches[0][1], declared)
		}
	}
	if declared == "" {
		t.Fatal("no Go toolchain pin site was readable")
	}
}
