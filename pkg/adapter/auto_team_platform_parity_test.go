package adapter_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/adapter/antigravity"
	"github.com/insajin/autopus-adk/pkg/adapter/codex"
	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/adapter/opencode"
	"github.com/insajin/autopus-adk/pkg/config"
)

// autoTeamMarkers are the strings only the Claude auto-team surface carries
// (SPEC-TEAMAUTO-001 REQ-TA-11).
var autoTeamMarkers = []string{"--topology", "Agent Team Auto-Selection", "Automatic Entry from Route A", "auto_team"}

// TestNonClaudePlatformsCarryNoAutoTeamText proves S9: codex, antigravity-cli,
// opencode, and omp generate none of the auto-team text, on disk or in the
// returned file set.
func TestNonClaudePlatformsCarryNoAutoTeamText(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		generate func(root string) (*adapter.PlatformFiles, error)
	}{
		{"codex", func(root string) (*adapter.PlatformFiles, error) {
			return codex.NewWithRoot(root).Generate(context.Background(), config.DefaultFullConfig("teamauto"))
		}},
		{"antigravity-cli", func(root string) (*adapter.PlatformFiles, error) {
			return antigravity.NewWithRoot(root).Generate(context.Background(), config.DefaultFullConfig("teamauto"))
		}},
		{"opencode", func(root string) (*adapter.PlatformFiles, error) {
			return opencode.NewWithRoot(root, opencode.WithCLIVersion("1.18.7")).Generate(context.Background(), config.DefaultFullConfig("teamauto"))
		}},
		{"omp", func(root string) (*adapter.PlatformFiles, error) {
			cfg := config.DefaultFullConfig("teamauto")
			cfg.Platforms = []string{"omp"}
			return omp.NewWithRoot(root).Generate(context.Background(), cfg)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			files, err := tc.generate(root)
			require.NoError(t, err)
			for _, file := range files.Files {
				requireNoAutoTeamText(t, file.TargetPath, string(file.Content))
			}
			walked := 0
			err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil || entry.IsDir() {
					return walkErr
				}
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				walked++
				requireNoAutoTeamText(t, path, string(data))
				return nil
			})
			require.NoError(t, err)
			require.NotZero(t, walked, "Generate wrote nothing to inspect")
		})
	}
}

func requireNoAutoTeamText(t *testing.T, path, content string) {
	t.Helper()
	for _, marker := range autoTeamMarkers {
		require.False(t, strings.Contains(content, marker), "%s carries %q", path, marker)
	}
}
