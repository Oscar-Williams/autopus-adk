package cli

// SPEC-REVMODEL-001 T8 (REQ-RVM-01, 14; acceptance S6): without its
// interactive compose gate, and under a truthy CI, no entry point lists the
// review tab, asks a review select, reads the OMP catalog for it, makes a
// live check request, or writes autopus.yaml through it.

import (
	"strings"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
)

// countReviewForms keeps the real forms probe, which picks the line driver
// for a reader that is not the process stdin, and counts every form run.
func countReviewForms(t *testing.T) *int {
	t.Helper()
	originalProbe, originalRun := composeFormsProbe, composeRunForm
	t.Cleanup(func() { composeFormsProbe, composeRunForm = originalProbe, originalRun })
	composeFormsProbe = composeFormsTerminal
	runs := new(int)
	composeRunForm = func(*cobra.Command, *huh.Form) error {
		*runs++
		return nil
	}
	return runs
}

func assertNoReviewTab(t *testing.T, out string, update reviewUpdate) {
	t.Helper()
	assert.NotContains(t, out, "멀티프로바이더 리뷰")
	assert.NotContains(t, out, "리뷰 참여")
	assert.Empty(t, update.smoke.recorded())
	assert.Empty(t, update.runner.calls)
}

func TestReviewTab_S6NonInteractiveRunsHaveNoTab(t *testing.T) {
	t.Run("(a) update --yes", func(t *testing.T) {
		dir := writeReviewR3(t)
		update := newReviewUpdate(t, dir, nil)
		b0 := reviewFileSum(t, dir)
		forms := countReviewForms(t)

		out, err := runComposeUpdate(t, strings.NewReader(""), "--dir", dir, "--yes")

		require.NoError(t, err, out)
		assertNoReviewTab(t, out, update)
		assert.Zero(t, *forms)
		assert.Equal(t, b0, reviewFileSum(t, dir))
	})
	t.Run("(b) update with scripted input", func(t *testing.T) {
		dir := writeReviewR3(t)
		update := newReviewUpdate(t, dir, nil)
		b0 := reviewFileSum(t, dir)
		forms := countReviewForms(t)

		out, err := runComposeUpdate(t, strings.NewReader("\n\n\n"), "--dir", dir)

		require.NoError(t, err, out)
		assert.Contains(t, out, "구성을 바꿀까요? [Enter] 유지")
		assertNoReviewTab(t, out, update)
		assert.Zero(t, *forms)
		assert.Equal(t, b0, reviewFileSum(t, dir))
	})
	t.Run("(c) quality --apply with scripted input", func(t *testing.T) {
		dir := writeReviewR3(t)
		update := newReviewUpdate(t, dir, nil)
		b0 := reviewFileSum(t, dir)
		forms := countReviewForms(t)
		deps := ompPlatformDependencies{newRunner: func() omp.OMPModelCatalogRunner { return update.runner }}

		out, err := runQualityForms(t, dir, &deps, strings.NewReader("\n\n\n"), "--apply")

		require.NoError(t, err, out)
		assertNoReviewTab(t, out, update)
		assert.Zero(t, *forms)
		assert.Equal(t, b0, reviewFileSum(t, dir))
	})
}

// S6(d): forms and a terminal, but CI is truthy. update skips its compose
// step as before; quality and init open their forms without the review tab.
func TestReviewTab_S6TruthyCIHidesTheTab(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		dir := writeReviewR3(t)
		update := newReviewUpdate(t, dir, nil)
		stubDriverProbes(t, true, "true")
		b0 := reviewFileSum(t, dir)

		out, err := runComposeUpdate(t, formInput("2\n", "\n"), "--dir", dir)

		require.NoError(t, err, out)
		assert.NotContains(t, out, "구성을 바꿀까요?")
		assertNoReviewTab(t, out, update)
		assert.Equal(t, b0, reviewFileSum(t, dir))
	})
	t.Run("quality", func(t *testing.T) {
		dir := writeReviewR3(t)
		update := newReviewUpdate(t, dir, nil)
		stubDriverProbes(t, true, "true")
		b0 := reviewFileSum(t, dir)
		deps := ompPlatformDependencies{newRunner: func() omp.OMPModelCatalogRunner { return update.runner }}

		out, err := runQualityForms(t, dir, &deps, formInput("2\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"))

		require.NoError(t, err, out)
		assert.Contains(t, out, "어느 도구를 바꿀까요? \n1. 모든 도구 같게\n2. Claude Code\n3. Codex\n4. OMP\nEnter a number between 1 and 4: ")
		assertNoReviewTab(t, out, update)
		assert.Equal(t, b0, reviewFileSum(t, dir))
	})
	t.Run("init", func(t *testing.T) {
		newReviewInitRun(t, true)
		stubDriverProbes(t, true, "true")
		smoke := useReviewSmokeBackend(t)

		_, out, err := runComposeInit(t, formInput("2\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"))

		require.NoError(t, err, out)
		assert.Contains(t, out, reviewInitTab+"Enter a number between 1 and 3: ")
		assert.NotContains(t, out, "멀티프로바이더 리뷰")
		assert.Empty(t, smoke.recorded())
	})
}
