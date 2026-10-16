package cli

// SPEC-REVMODEL-001 T3 (REQ-RVM-06, 07, 08; acceptance S2, S4): the model
// select offers exactly the attested installed models of the provider's
// family, the thinking select follows the OMP Autopus tab rule, and an
// unreadable catalog, catalog_invalid included, offers nothing.

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func reviewSelectors(choices []ompCustomModelChoice) []string {
	selectors := make([]string, 0, len(choices))
	for _, choice := range choices {
		selectors = append(selectors, choice.selector())
	}
	return selectors
}

// S2: the family of each provider name picks its models; example/foo-1
// belongs to none of them, and an unknown family lists every attested model.
func TestReviewModelChoices_FilterByProviderFamily(t *testing.T) {
	choices := reviewCATGChoices(t)

	assert.Equal(t, []string{"anthropic/claude-fable-5-1", "anthropic/claude-opus-5-5"}, reviewSelectors(reviewModelChoices("claude", choices)))
	assert.Equal(t, []string{"openai-codex/gpt-6-astra", "openai-codex/gpt-6-luna"}, reviewSelectors(reviewModelChoices("codex", choices)))
	assert.Equal(t, []string{"google-antigravity/gemini-3.1-pro", "google-antigravity/gemini-3.8-flash"},
		reviewSelectors(reviewModelChoices("gemini", choices)))
	assert.Equal(t, reviewSelectors(choices), reviewSelectors(reviewModelChoices("zeta", choices)))
	assert.Contains(t, reviewSelectors(choices), "example/foo-1")
}

// S2: levels in the normalized catalog order; the working level when the
// model reports it, otherwise the first.
func TestReviewThinking_FollowsTheOMPAutopusRule(t *testing.T) {
	choices := reviewCATGChoices(t)
	luna := reviewModelChoices("codex", choices)[1]
	opus := reviewModelChoices("claude", choices)[1]

	assert.Equal(t, []string{"high", "low", "medium"}, luna.Thinking)
	assert.Equal(t, "high", reviewThinkingDefault(luna, "openai-codex/gpt-6-astra:max"))
	assert.Equal(t, "low", reviewThinkingDefault(luna, "openai-codex/gpt-6-astra:low"))
	assert.Equal(t, []string{"high", "max"}, opus.Thinking)
	assert.Equal(t, "max", reviewThinkingDefault(opus, "anthropic/claude-fable-5-1:max"))
}

// S4(a): 그대로 names the working selector and is marked when the attested
// list lacks its model.
func TestReviewModelOptions_MarkAModelOffTheCatalog(t *testing.T) {
	gemini := reviewModelChoices("gemini", reviewChoicesOf(t, reviewCatalogCATN))

	options := reviewModelOptions("그대로", "google-antigravity/gemini-3.1-pro:high", "google-antigravity/gemini-3.1-pro:high", gemini, composeReviewTextKO)

	require.Len(t, options, 2)
	assert.Equal(t, composeOption[string]{"그대로 — google-antigravity/gemini-3.1-pro:high (catalog에 없음)", ""}, options[0])
	assert.Equal(t, composeOption[string]{"google-antigravity/gemini-3.8-flash", "google-antigravity/gemini-3.8-flash"}, options[1])

	codex := reviewModelChoices("codex", reviewCATGChoices(t))
	options = reviewModelOptions("그대로", "openai-codex/gpt-6-astra:max", "openai-codex/gpt-6-astra:max", codex, composeReviewTextKO)
	assert.Equal(t, "그대로 — openai-codex/gpt-6-astra:max", options[0].label)
	options = reviewModelOptions("그대로", "openai-codex/gpt-6-astra:max", `"openai-codex/gpt-6-astra:max"`, codex, composeReviewTextKO)
	assert.Equal(t, `그대로 — "openai-codex/gpt-6-astra:max"`, options[0].label, "the label shows the text it is given")
}

// S4(c), REQ-RVM-14: the catalog is read once per tab entry, and a failed
// read or a catalog_invalid catalog is unreadable.
func TestReviewCatalog_ReadsOnceAndReportsUnreadable(t *testing.T) {
	cases := []struct {
		name   string
		runner *ompCLIFakeRunner
		want   bool
	}{
		{"CATG", &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATG)}, true},
		{"CATE", &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATE)}, false},
		{"CATX", &ompCLIFakeRunner{fail: true}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			catalog := &reviewCatalog{runner: tc.runner}
			first, ok := catalog.load(context.Background(), &out, "note")
			assert.Equal(t, tc.want, ok)
			calls := len(tc.runner.calls)
			second, again := catalog.load(context.Background(), &out, "note")
			assert.Equal(t, ok, again)
			assert.Equal(t, first, second)
			assert.Len(t, tc.runner.calls, calls, "one read per tab entry")
			if tc.want {
				assert.Len(t, first, 7)
			}
		})
	}
	catalog := &reviewCatalog{}
	_, ok := catalog.load(context.Background(), &bytes.Buffer{}, "note")
	assert.False(t, ok, "a run without an OMP catalog runner has no catalog")
}
