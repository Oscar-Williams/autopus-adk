package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/config"
)

// reviewArgsRowFor is the CLI Model Contract row of a provider binary, matched
// by its file name.
func reviewArgsRowFor(binary string) reviewArgsRow {
	switch strings.TrimSuffix(filepath.Base(binary), ".exe") {
	case "claude":
		return reviewClaudeRow
	case "codex":
		return reviewCodexRow
	case "agy":
		return reviewAgyRow
	}
	return reviewArgsRow{kind: reviewRowOther}
}

// reviewCLIModel is one model a CLI row offers: the value written, the
// efforts it takes (none for haiku and for agy ids, which carry theirs), and
// for agy the group it belongs to and its display name.
type reviewCLIModel struct {
	ID      string
	Efforts []string
	Base    string
	Suffix  string
	Display string
}

// reviewClaudeModels is the claude row: the Claude tab's model of every tier,
// with the low–max efforts for every model but haiku (REQ-RVM-21).
func reviewClaudeModels() []reviewCLIModel {
	models := make([]reviewCLIModel, 0, 4)
	for _, tier := range []string{"fable", "opus", "sonnet", "haiku"} {
		model := reviewCLIModel{ID: config.ClaudeModelForTier(tier)}
		if tier != "haiku" {
			model.Efforts = config.QualityEfforts()
		}
		models = append(models, model)
	}
	return models
}

// reviewCodexModels is the codex row: the run's Codex catalog snapshot models
// that report at least one of low–max, by name, each with those levels in
// ladder order (REQ-RVM-22). A model whose levels miss the ladder entirely
// is left out. It reports false for a snapshot it cannot read.
func reviewCodexModels(snapshot driftGenerationSnapshot) ([]reviewCLIModel, bool) {
	if !snapshot.codexCatalogFixed || len(snapshot.codexCatalog) == 0 {
		return nil, false
	}
	catalog, err := config.ParseCodexModelCatalog(snapshot.codexCatalog)
	if err != nil {
		return nil, false
	}
	var models []reviewCLIModel
	for _, entry := range catalog.Models {
		model := reviewCLIModel{ID: entry.Slug}
		for _, level := range config.QualityEfforts() {
			if slices.ContainsFunc(entry.SupportedReasoningLevels, func(reported config.CodexCatalogReasoningLevel) bool {
				return reported.Effort == level
			}) {
				model.Efforts = append(model.Efforts, level)
			}
		}
		if len(model.Efforts) > 0 && !slices.ContainsFunc(models, func(m reviewCLIModel) bool { return m.ID == model.ID }) {
			models = append(models, model)
		}
	}
	slices.SortFunc(models, func(a, b reviewCLIModel) int { return strings.Compare(a.ID, b.ID) })
	return models, true
}

// reviewRawProvider is the raw document's view of one provider, without
// environment expansion: its args values (nil when the key is absent or not a
// list) and its model, model_policy, and backend scalars. Previews read CLI
// values only from here (REQ-RVM-26).
type reviewRawProvider struct {
	Args    []string
	Model   string
	Policy  string
	Backend string
}

// reviewRawOf is the view of a config no file holds, such as init's.
func reviewRawOf(cfg *config.HarnessConfig) map[string]reviewRawProvider {
	raw := make(map[string]reviewRawProvider, len(cfg.Orchestra.Providers))
	for name, entry := range cfg.Orchestra.Providers {
		raw[name] = reviewRawProvider{Args: entry.Args, Model: entry.Model, Policy: entry.ModelPolicy, Backend: entry.Backend}
	}
	return raw
}

// readReviewRaw reads the view of every provider from dir's autopus.yaml.
// A file that cannot be read or parsed, or that holds no providers mapping,
// gives no provider a view, never the expanded config's: the tab then keeps
// every model rather than show or decide on a value it cannot read as
// written (REQ-RVM-26).
func readReviewRaw(dir string) map[string]reviewRawProvider {
	raw := map[string]reviewRawProvider{}
	data, err := os.ReadFile(filepath.Join(dir, "autopus.yaml"))
	if err != nil {
		return raw
	}
	document, _, err := parseAutopusConfigDocument(data)
	if err != nil {
		return raw
	}
	providers := yamlPathValue(document, []string{"orchestra", "providers"})
	if providers == nil || providers.Kind != yaml.MappingNode {
		return raw
	}
	for i := 0; i+1 < len(providers.Content); i += 2 {
		entry, view := providers.Content[i+1], reviewRawProvider{}
		if _, args := yamlMappingPair(entry, "args"); args != nil && args.Kind == yaml.SequenceNode {
			view.Args = reviewArgsValues(args)
		}
		view.Model = reviewRawScalar(entry, "model")
		view.Policy = reviewRawScalar(entry, "model_policy")
		view.Backend = reviewRawScalar(entry, "backend")
		raw[providers.Content[i].Value] = view
	}
	return raw
}

func reviewRawScalar(mapping *yaml.Node, key string) string {
	if _, value := yamlMappingPair(mapping, key); value != nil && value.Kind == yaml.ScalarNode {
		return value.Value
	}
	return ""
}

// reviewMigratedConfig is a detached copy of cfg after
// config.MigrateOrchestraConfig: the args a CLI provider runs with, which the
// tab decides current models on (REQ-RVM-25). `auto quality` loads without
// that migration, so its own config can hold args the runtime never runs.
func reviewMigratedConfig(cfg *config.HarnessConfig) *config.HarnessConfig {
	migrated, err := cloneHarnessConfig(cfg)
	if err != nil || migrated == nil {
		return cfg
	}
	if _, err := config.MigrateOrchestraConfig(migrated); err != nil {
		return cfg
	}
	return migrated
}

// reviewCLIKeepText names a CLI value in a keep label: `<model> · <effort>`,
// or the binary when no model flag is set.
func reviewCLIKeepText(binary string, value reviewCLIValue) string {
	if value.Model == "" {
		return binary
	}
	if value.Effort == "" {
		return value.Model
	}
	return value.Model + " · " + value.Effort
}

// reviewCLICellText is the after table's model cell of a CLI provider:
// `<binary> <model> · <effort>`, or the binary alone.
func reviewCLICellText(binary string, value reviewCLIValue) string {
	if value.Model == "" {
		return binary
	}
	return binary + " " + reviewCLIKeepText(binary, value)
}
