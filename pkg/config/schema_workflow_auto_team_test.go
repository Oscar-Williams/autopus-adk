package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// autoTeamKeyLine is the line of auto_team in writeWorkflowConfig(t,
// "workflow:\n  auto_team: <v>\n"): workflowTestBase opens with a blank line
// and spans five lines.
const autoTeamKeyLine = 7

func TestWorkflowAutoTeamAcceptsOnlyBoolNodes(t *testing.T) {
	t.Parallel()
	for value, want := range map[string]bool{
		"false": false, "False": false, "FALSE": false, `!!bool "false"`: false,
		"true": true, "True": true, "TRUE": true, `!!bool "true"`: true,
	} {
		for _, load := range []func(string) (*HarnessConfig, error){Load, LoadPreview} {
			cfg, err := load(writeWorkflowConfig(t, "workflow:\n  auto_team: "+value+"\n"))
			require.NoError(t, err, value)
			require.NotNil(t, cfg.Workflow.AutoTeam, value)
			assert.Equal(t, want, *cfg.Workflow.AutoTeam, value)
			assert.Equal(t, want, cfg.Workflow.AutoTeamEnabled(), value)
			assert.Equal(t, DefaultCoverageThreshold, cfg.Workflow.CoverageThreshold, value)
		}
	}
}

func TestWorkflowAutoTeamAbsentKeyIsEnabled(t *testing.T) {
	t.Parallel()
	for _, body := range []string{"", "workflow:\n  coverage_threshold: 70\n"} {
		cfg, err := LoadPreview(writeWorkflowConfig(t, body))
		require.NoError(t, err)
		assert.Nil(t, cfg.Workflow.AutoTeam)
		assert.True(t, cfg.Workflow.AutoTeamEnabled())
	}
	assert.True(t, WorkflowConf{AutoTeam: boolPtr(true)}.AutoTeamEnabled())
	assert.False(t, WorkflowConf{AutoTeam: boolPtr(false)}.AutoTeamEnabled())
}

func TestWorkflowAutoTeamRejectsEveryNonBoolNode(t *testing.T) {
	t.Parallel()
	values := []string{`"false"`, `"true"`, `'false'`, "null", "~", "", "yes", `"yes"`, "y", "no", "n", "off", `"off"`, "on",
		"1", "0", "[true]", "{enabled: true}", "!!str true"}
	for _, value := range values {
		for _, load := range []func(string) (*HarnessConfig, error){Load, LoadPreview} {
			dir := writeWorkflowConfig(t, "workflow:\n  auto_team: "+value+"\n")
			before, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
			require.NoError(t, err)
			_, err = load(dir)
			require.Error(t, err, value)
			assert.Contains(t, err.Error(), fmt.Sprintf("workflow.auto_team: must be true or false (line %d)", autoTeamKeyLine), value)
			after, readErr := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
			require.NoError(t, readErr)
			assert.Equal(t, before, after, "a rejected load must not rewrite the file")
		}
	}
}

func TestWorkflowAutoTeamCheckFollowsAliasesAndMergeKeys(t *testing.T) {
	t.Parallel()
	// future_extension is reserved and pruned before decoding, yet the nodes
	// its anchors name still reach the decoder through the aliases below.
	merged := "future_extension:\n  base: &base\n    auto_team: yes\nworkflow:\n  <<: *base\n"
	_, err := LoadPreview(writeWorkflowConfig(t, merged))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow.auto_team: must be true or false (line 8)")

	aliased := "future_extension:\n  flag: &flag on\nworkflow:\n  auto_team: *flag\n"
	_, err = LoadPreview(writeWorkflowConfig(t, aliased))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow.auto_team: must be true or false (line 9)")

	accepted := "future_extension:\n  flag: &flag false\nworkflow: &wf\n  auto_team: *flag\n"
	cfg, err := LoadPreview(writeWorkflowConfig(t, accepted))
	require.NoError(t, err)
	require.NotNil(t, cfg.Workflow.AutoTeam)
	assert.False(t, *cfg.Workflow.AutoTeam)

	sequence := "future_extension:\n  a: &a {coverage_threshold: 70}\n  b: &b {auto_team: \"no\"}\nworkflow:\n  <<: [*a, *b, *a]\n"
	_, err = LoadPreview(writeWorkflowConfig(t, sequence))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow.auto_team: must be true or false (line 8)")

	var empty yaml.Node
	assert.NoError(t, checkWorkflowAutoTeamNode(&empty))
}

func TestWorkflowAutoTeamIsNeverWrittenUnlessSet(t *testing.T) {
	t.Parallel()
	encoded, err := yaml.Marshal(DefaultFullConfig("fresh"))
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "auto_team")

	for _, value := range []bool{false, true} {
		dir := writeWorkflowConfig(t, fmt.Sprintf("workflow:\n  auto_team: %v\n", value))
		cfg, err := Load(dir)
		require.NoError(t, err)
		require.NoError(t, Save(dir, cfg))
		saved, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
		require.NoError(t, err)
		assert.Equal(t, 1, strings.Count(string(saved), fmt.Sprintf("auto_team: %v", value)), string(saved))
		reloaded, err := LoadPreview(dir)
		require.NoError(t, err)
		require.NotNil(t, reloaded.Workflow.AutoTeam)
		assert.Equal(t, value, *reloaded.Workflow.AutoTeam)
	}

	dir := writeWorkflowConfig(t, "workflow:\n  coverage_threshold: 70\n")
	cfg, err := Load(dir)
	require.NoError(t, err)
	require.NoError(t, Save(dir, cfg))
	saved, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	require.NoError(t, err)
	assert.NotContains(t, string(saved), "auto_team")
}

func TestWorkflowAutoTeamIsAStrictSchemaKey(t *testing.T) {
	t.Parallel()
	assert.False(t, slices.Contains(removedConfigKeys, "workflow.auto_team"))
	assert.False(t, slices.Contains(reservedConfigKeys, "auto_team"))
	_, err := LoadPreview(writeWorkflowConfig(t, "workflow:\n  auto_teams: false\n"))
	require.Error(t, err, "a misspelt key still fails strict decoding")
}

func boolPtr(value bool) *bool { return &value }
