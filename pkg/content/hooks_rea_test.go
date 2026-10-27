package content_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/adapter/antigravity"
	"github.com/insajin/autopus-adk/pkg/adapter/claude"
	"github.com/insajin/autopus-adk/pkg/adapter/codex"
	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/adapter/opencode"
	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/content"
)

// SPEC-REA-001 REQ-REA-15 / S14: Autopus pre-approves no REA MCP tool and no
// project MCP server. The only pre-approved MCP permission stays the
// sequential-thinking tool, and the agent frontmatter's `mcp__rea` stays a tool
// allow-list entry rather than a permission.
const sequentialThinkingPermission = "mcp__sequential-thinking__sequentialthinking"

// reaPreapprovalKeys are the settings that would approve a project MCP server
// or its tools without the user's own prompt answer.
var reaPreapprovalKeys = []string{
	"enableAllProjectMcpServers",
	"enabledMcpjsonServers",
	"enabled_tools",
	"default_tools_approval_mode",
}

func mcpPermissions(allow []string) []string {
	var mcp []string
	for _, value := range allow {
		if strings.HasPrefix(value, "mcp__") {
			mcp = append(mcp, value)
		}
	}
	return mcp
}

func TestDetectPermissions_PreapprovesNoReaMCP(t *testing.T) {
	t.Parallel()
	for name, marker := range map[string]string{"plain": "", "go": "go.mod", "node": "package.json"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if marker != "" {
				require.NoError(t, os.WriteFile(filepath.Join(root, marker), []byte("{}\n"), 0o644))
			}
			perms := content.DetectPermissions(root, config.PermissionsConf{})
			assert.Equal(t, []string{sequentialThinkingPermission}, mcpPermissions(perms.Allow))
			for _, value := range perms.Allow {
				assert.False(t, strings.HasPrefix(value, "mcp__rea"), value)
			}
		})
	}
}

// S14: a five-platform project generated with the REA add action carries no
// pre-approval in any generated file.
func TestGeneratedSurfaces_PreapproveNoReaMCP(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cfg := config.DefaultFullConfig("rea-preapproval")
	cfg.Platforms = []string{"claude-code", "codex", "antigravity-cli", "omp", "opencode"}
	cfg.Runtime.ReaMCP = config.ReaMCPAdd
	for _, platform := range []adapter.PlatformAdapter{
		claude.NewWithRoot(root),
		codex.NewWithRoot(root),
		antigravity.NewWithRoot(root),
		omp.NewWithRoot(root),
		opencode.NewWithRoot(root, opencode.WithCLIVersion("1.18.7")),
	} {
		_, err := platform.Generate(context.Background(), cfg)
		require.NoError(t, err, platform.Name())
	}

	var settings struct {
		Permissions struct {
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &settings))
	assert.Equal(t, []string{sequentialThinkingPermission}, mcpPermissions(settings.Permissions.Allow))

	// Antigravity registers REA in the workspace plugin's mcp_config.json.
	var plugin struct {
		MCPServers map[string]map[string]any `json:"mcpServers"`
	}
	data, err = os.ReadFile(filepath.Join(root, ".agents", "plugins", "autopus", "mcp_config.json"))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &plugin))
	require.Contains(t, plugin.MCPServers, "rea")
	for _, key := range []string{"trust", "enabledTools", "disabledTools"} {
		assert.NotContains(t, plugin.MCPServers["rea"], key, "the managed entry neither trusts nor filters the server")
	}

	codexConfig, err := os.ReadFile(filepath.Join(root, ".codex", "config.toml"))
	require.NoError(t, err)
	assert.NotContains(t, string(codexConfig), "[mcp_servers.rea.tools")

	files := 0
	require.NoError(t, filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || !entry.Type().IsRegular() {
			return walkErr
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files++
		rel, _ := filepath.Rel(root, path)
		for _, key := range reaPreapprovalKeys {
			assert.NotContains(t, string(body), key, rel)
		}
		return nil
	}))
	assert.Greater(t, files, 100, "the walk must cover the generated surface")
}
