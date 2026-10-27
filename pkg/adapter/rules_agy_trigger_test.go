package adapter_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	contentfs "github.com/insajin/autopus-adk/content"
	pkgcontent "github.com/insajin/autopus-adk/pkg/content"
	"github.com/insajin/autopus-adk/pkg/rulecond"
)

// agyRuleRoots are the two copies of every Antigravity rule: the workspace
// plugin agy loads, and the legacy Gemini-compatible surface it mirrors.
var agyRuleRoots = []string{".agents/plugins/autopus/rules", ".gemini/rules/autopus"}

// agyRuleFrontmatter is the part of a rule frontmatter agy 1.4.2 reads. `globs`
// is a plain string on purpose: agy's loader skips a rule whose `globs` is a
// YAML list ("cannot unmarshal !!seq into string") and splits the string on
// commas, so decoding it the same way is what keeps a list from passing here.
type agyRuleFrontmatter struct {
	Description string `yaml:"description"`
	Trigger     string `yaml:"trigger"`
	Globs       string `yaml:"globs"`
}

// agyTriggerFor maps how claude-code loads a canonical rule onto the agy
// trigger with the same loading condition. agy has no tool-call trigger and no
// skill-reference trigger, so hook-fired and skill-scoped rules stay out of
// baseline context as model_decision rules: agy lists only their name and
// description and loads the body when the model decides it applies.
func agyTriggerFor(class rulecond.Class) string {
	switch class {
	case rulecond.ClassAlways:
		return "always_on"
	case rulecond.ClassPathsScoped:
		return "glob"
	default:
		return "model_decision"
	}
}

func parseAgyRuleFrontmatter(t *testing.T, path string) agyRuleFrontmatter {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	require.True(t, strings.HasPrefix(content, "---\n"),
		"%s must open with frontmatter: agy never applies a rule without a trigger", path)
	end := strings.Index(content[len("---\n"):], "\n---\n")
	require.GreaterOrEqual(t, end, 0, "%s must close its frontmatter", path)

	var fm agyRuleFrontmatter
	require.NoError(t, yaml.Unmarshal([]byte(content[len("---\n"):len("---\n")+end]), &fm),
		"%s frontmatter must decode the way agy's rule loader decodes it", path)
	return fm
}

// TestAgyRuleTriggers_MatchClaudeLoadingClass: agy 1.4.2 applies a plugin rule
// only through a valid `trigger` (`agy plugin validate`: "no valid trigger in
// the frontmatter, so the rule never applies"). Every generated rule carries
// the trigger that loads it under the same condition claude-code loads the
// canonical rule, so the always_on set is exactly claude-code's baseline set.
func TestAgyRuleTriggers_MatchClaudeLoadingClass(t *testing.T) {
	t.Parallel()

	dir := generatePlatform(t, "gemini")
	for _, root := range agyRuleRoots {
		entries, err := os.ReadDir(filepath.Join(dir, filepath.FromSlash(root)))
		require.NoError(t, err)
		require.NotEmpty(t, entries)

		var alwaysOn []string
		for _, entry := range entries {
			name := entry.Name()
			fm := parseAgyRuleFrontmatter(t, filepath.Join(dir, filepath.FromSlash(root), name))

			if platformRuleAdditions["gemini"][name] {
				assert.Equal(t, "always_on", fm.Trigger, "%s/%s", root, name)
				continue
			}

			raw, err := fs.ReadFile(contentfs.FS, pkgcontent.EmbeddedPath("rules", name))
			require.NoError(t, err, "%s/%s has no canonical content rule", root, name)
			source, err := rulecond.ParseRule(name, raw)
			require.NoError(t, err)

			want := agyTriggerFor(rulecond.Classify(source))
			assert.Equal(t, want, fm.Trigger, "%s/%s", root, name)
			switch want {
			case "always_on":
				alwaysOn = append(alwaysOn, name)
				assert.Empty(t, fm.Globs, "%s/%s", root, name)
			case "glob":
				assert.Equal(t, strings.Join(source.Globs, ","), fm.Globs,
					"%s/%s must match the files claude-code loads it for", root, name)
			case "model_decision":
				// The description is all agy shows the model for this rule.
				assert.True(t, strings.HasPrefix(fm.Description, "Apply "),
					"%s/%s description must say when the rule applies: %q", root, name, fm.Description)
			}
		}

		want := make([]string, 0, len(alwaysRuleFiles))
		for name := range alwaysRuleFiles {
			want = append(want, name)
		}
		sort.Strings(want)
		sort.Strings(alwaysOn)
		assert.Equal(t, want, alwaysOn, "%s always_on rules beyond the agy-only additions", root)
	}
}
