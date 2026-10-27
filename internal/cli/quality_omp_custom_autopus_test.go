package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/config"
)

// In autopus mode the custom editor asks the 17 Autopus agents and no bundled
// name. A task pin would fold onto the fifteen roles that collapsed onto task
// before v0.50.127 (Route Contract rule 2), so pinning "task" used to hand
// annotator, frontend-specialist, ux-validator, and validator a model without
// their capability and the plan refused it. task follows the planner pin.
func TestQualityOMPWizardCustomAutopusPinsOneAutopusAgent(t *testing.T) {
	dir, runner := qualityOMPWizardFixture(t)
	runner.catalog = ompCLIBalancedCatalogJSON()

	// planner is the tenth agent: model 4 (openai-codex/gpt-6-astra) at max.
	input := "omp\ncustom\ngpt\n" + strings.Repeat("\n", 9) + "4\nmax\n" + strings.Repeat("\n", 7) + "y\n"
	out, err := runQualityOMPWizard(t, dir, input, ompPlatformDependencies{
		newRunner: func() omp.OMPModelCatalogRunner { return runner },
		activate:  func(context.Context, string, *config.HarnessConfig) error { return nil },
	})
	require.NoError(t, err, out)
	assert.Contains(t, out, "planner=openai-codex/gpt-6-astra:max")
	for _, bundled := range []string{"task (", "scout (", "sonic ("} {
		assert.NotContains(t, out, bundled, "autopus mode must not offer a bundled name")
	}

	cfg, err := config.LoadPreview(dir)
	require.NoError(t, err)
	assert.Equal(t, "balanced", cfg.RoleModelPolicy.Profile)
	pinned := cfg.RoleModelPolicy.Agents["planner"].Candidates
	require.Len(t, pinned, 1)
	assert.Equal(t, "openai-codex/gpt-6-astra", pinned[0].Selector)
	assert.Len(t, cfg.RoleModelPolicy.Agents, 1, "only the chosen agent is pinned")
}
