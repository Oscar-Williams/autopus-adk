package antigravity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// agy splits `globs` on commas, so the join must refuse what that split would
// break or leave without a pattern rather than emit a glob rule that never
// applies.
func TestAgyGlobs_JoinsOnlyWhatAgyCanSplitBack(t *testing.T) {
	t.Parallel()

	joined, err := agyGlobs([]string{"**/*.go", "**/*.ts"})
	require.NoError(t, err)
	assert.Equal(t, "**/*.go,**/*.ts", joined)

	_, err = agyGlobs(nil)
	assert.Error(t, err, "a glob rule without globs never applies")

	_, err = agyGlobs([]string{"**/*.{go,ts}"})
	assert.ErrorContains(t, err, "**/*.{go,ts}", "agy would split the brace set into two broken globs")
}

func TestFileSizeLimitGlobs_CarriesTheCanonicalSourceGlobs(t *testing.T) {
	t.Parallel()

	globs, err := fileSizeLimitGlobs()
	require.NoError(t, err)
	assert.Equal(t, "**/*.go,**/*.ts,**/*.tsx,**/*.js,**/*.jsx,**/*.py,**/*.rs", globs)
}
