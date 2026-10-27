package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter"
	"github.com/insajin/autopus-adk/pkg/config"
)

// codexCatalogEdited rewrites fullCodexCatalogForTest: drop removes models by
// slug, levels replaces one model's reasoning levels.
func codexCatalogEdited(t *testing.T, drop []string, levels map[string][]string) []byte {
	t.Helper()
	var catalog struct {
		Models []map[string]any `json:"models"`
	}
	require.NoError(t, json.Unmarshal([]byte(fullCodexCatalogForTest), &catalog))
	kept := catalog.Models[:0]
	for _, model := range catalog.Models {
		slug := model["slug"].(string)
		if containsString(drop, slug) {
			continue
		}
		if replaced, ok := levels[slug]; ok {
			entries := make([]map[string]string, 0, len(replaced))
			for _, level := range replaced {
				entries = append(entries, map[string]string{"effort": level})
			}
			model["supported_reasoning_levels"] = entries
		}
		kept = append(kept, model)
	}
	catalog.Models = kept
	data, err := json.Marshal(catalog)
	require.NoError(t, err)
	return data
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// generateWithReaSink runs one Generate over catalog and collects the REA
// fallbacks it reports and the adapter's own diagnostics.
func generateWithReaSink(t *testing.T, catalog []byte) (*adapter.PlatformFiles, []adapter.ReaFallback, string, error) {
	t.Helper()
	a := NewWithRoot(t.TempDir(), WithModelCatalog(catalog))
	var diagnostics bytes.Buffer
	a.codexFallbackWriter = &diagnostics
	var facts []adapter.ReaFallback
	ctx := adapter.WithReaFallbackSink(context.Background(), func(fact adapter.ReaFallback) {
		facts = append(facts, fact)
	})
	files, err := a.Generate(ctx, config.DefaultFullConfig("rea-codex"))
	return files, facts, diagnostics.String(), err
}

// Scope expansion of SPEC-REA-001 Gate 1, heterogeneous oracle: Sol present
// pins the analysis rung; Sol missing, or Sol without xhigh, leaves only
// reverse-engineer on the session model and reports one fact; a missing
// existing-rung model keeps today's refusal.
func TestCodexReaFallback_AnalysisRungFollowsTheCatalog(t *testing.T) {
	t.Parallel()
	sol := config.CodexProfile{Model: config.CodexSolModel, Effort: config.CodexEffortXHigh}
	fact := adapter.ReaFallback{Catalog: "Codex", Agent: "reverse-engineer", Requested: "gpt-6.1-sol/xhigh"}

	files, facts, _, err := generateWithReaSink(t, []byte(fullCodexCatalogForTest))
	require.NoError(t, err)
	assertCodexRenderedProfile(t, codexAgentMappingContent(t, files.Files, "reverse-engineer.toml"), sol)
	assert.Empty(t, facts)

	for name, catalog := range map[string][]byte{
		"sol missing":       codexCatalogEdited(t, []string{config.CodexSolModel}, nil),
		"sol without xhigh": codexCatalogEdited(t, nil, map[string][]string{config.CodexSolModel: {"low", "high"}}),
	} {
		files, facts, diagnostics, err := generateWithReaSink(t, catalog)
		require.NoError(t, err, name)
		assertCodexRenderedProfile(t, codexAgentMappingContent(t, files.Files, "reverse-engineer.toml"), config.CodexProfile{})
		assertCodexRenderedProfile(t, codexAgentMappingContent(t, files.Files, "planner.toml"),
			config.CodexProfile{Model: config.CodexAstraModel, Effort: config.CodexEffortMax})
		assert.Equal(t, []adapter.ReaFallback{fact}, facts, name)
		assert.NotContains(t, diagnostics, "gpt-6.1-sol", "%s: the CLI prints the fallback, not the adapter", name)
	}

	_, facts, _, err = generateWithReaSink(t, codexCatalogEdited(t, []string{config.CodexAstraModel}, nil))
	require.ErrorIs(t, err, ErrCodexNativePlacementUnavailable)
	assert.True(t, strings.Contains(err.Error(), config.CodexAstraModel), err.Error())
	assert.Empty(t, facts, "a refused run reports no fallback")
}

// The pure resolver the composition table shares answers the same way.
func TestResolveAgentProfile_AnalysisRungSessionModel(t *testing.T) {
	t.Parallel()
	quality := config.DefaultFullConfig("rea-resolve").Quality
	catalog := codexCatalogEdited(t, []string{config.CodexSolModel}, nil)

	resolution, err := ResolveAgentProfile(quality, catalog, "reverse-engineer", "opus", "high")
	require.NoError(t, err)
	assert.Equal(t, config.CodexSessionModelResolution(
		config.CodexProfile{Model: config.CodexSolModel, Effort: config.CodexEffortXHigh}), resolution)

	_, err = ResolveAgentProfile(quality, codexCatalogEdited(t, []string{config.CodexLunaModel, "gpt-5.6-luna"}, nil),
		"executor", "sonnet", "medium")
	require.ErrorIs(t, err, ErrCodexNativePlacementUnavailable, "an existing rung keeps the refusal")
}
