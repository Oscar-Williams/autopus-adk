package cli

// SPEC-REVMODEL-001 T7 (REQ-RVM-12; acceptance S11, S12): update and quality
// write a review edit through the line splice writer only when it changes
// something, a later update keeps it, and an unsafe target refuses through
// the tab with the file and the run's config unchanged.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/adapter/omp"
	"github.com/insajin/autopus-adk/pkg/config"
)

// The answers that turn gemini off and apply, and that keep everything.
var (
	reviewTurnOffGemini = []string{"\n", "\n", "2\n", "\n", "\n", "\n", "\n", "2\n", "\n"}
	reviewKeepAll       = []string{"\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n", "\n"}
)

// S11: an unchanged pass writes nothing, an applied one survives the rest
// of the update, and the next `auto update --yes` leaves the file alone.
func TestReviewTab_S11UpdateWritesEditsOnlyAndKeepsThem(t *testing.T) {
	dir := writeReviewR3(t)
	newReviewUpdate(t, dir, nil)
	b0 := reviewFileSum(t, dir)

	out, err := runComposeUpdate(t, formInput(reviewLines(reviewOpen, reviewKeepAll)...), "--dir", dir)
	require.NoError(t, err, out)
	assert.Contains(t, out, "변경 없음\n"+reviewCheckR3)
	assert.Equal(t, b0, reviewFileSum(t, dir))

	out, err = runComposeUpdate(t, formInput(reviewLines(reviewOpen, reviewTurnOffGemini)...), "--dir", dir)
	require.NoError(t, err, out)
	assert.Contains(t, out, "구성을 적용했습니다.\n")
	assert.Contains(t, out, "Update complete")
	assert.Equal(t, []string{"claude", "codex"}, reviewLoaderValueKeepCwd(t, dir).Spec.ReviewGate.Providers)
	applied := reviewFileSum(t, dir)

	out, err = runComposeUpdate(t, strings.NewReader(""), "--dir", dir, "--yes")
	require.NoError(t, err, out)
	assert.Equal(t, applied, reviewFileSum(t, dir))
}

// reviewLoaderValueKeepCwd reads the loader value without leaving the test
// in dir, for a test that runs more commands afterwards.
func reviewLoaderValueKeepCwd(t *testing.T, dir string) *config.HarnessConfig {
	t.Helper()
	cfg, err := loadHarnessConfigForDir(dir, globalFlags{})
	require.NoError(t, err)
	return cfg
}

// S12 through the tab: the write-refused line names the first refused key;
// the file and the run's orchestra and review gate stay as they were.
func TestReviewTab_S12UnsafeTargetsRefuseThroughTheTab(t *testing.T) {
	t.Setenv("CODEX_REVIEW_MODEL", "openai-codex/gpt-6-astra:max")
	luna := []string{"\n", "\n", "\n", "\n", "3\n", "\n", "\n", "\n", "\n", "2\n", "\n"}
	cases := []struct {
		name  string
		file  func(*testing.T, string) string
		lines []string
		key   string
	}{
		{"(a) placeholder model", reviewReplace("model: openai-codex/gpt-6-astra:max", "model: ${CODEX_REVIEW_MODEL}"), luna,
			"orchestra.providers.codex.model"},
		{"(b) aliased list", func(t *testing.T, text string) string {
			text = reviewReplace("            providers: [claude, codex, gemini]\n        secure:", "            providers: &rp [claude, codex, gemini]\n        secure:")(t, text)
			return reviewSpecLast(t, reviewReplace("        providers: [claude, codex, gemini]\n        judge", "        providers: *rp\n        judge")(t, text))
		}, reviewTurnOffGemini, "spec.review_gate.providers"},
		{"(c) anchored list", func(t *testing.T, text string) string {
			text = reviewReplace("        providers: [claude, codex, gemini]\n        judge", "        providers: &rp [claude, codex, gemini]\n        judge")(t, text)
			return reviewReplace("            providers: [claude, codex, gemini]\n        secure:", "            providers: *rp\n        secure:")(t, text)
		}, reviewTurnOffGemini, "spec.review_gate.providers"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeReviewR3(t)
			require.NoError(t, os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(tc.file(t, readReviewFile(t, dir))), 0o644))
			sum := reviewFileSum(t, dir)
			before, err := config.Load(dir)
			require.NoError(t, err)

			run := runReviewForms(t, dir, nil, reviewLines(reviewOpen, tc.lines)...)

			require.NoError(t, run.err, run.out)
			assert.Contains(t, run.out, "autopus.yaml의 리뷰 설정을 안전하게 고칠 수 없어 아무것도 쓰지 않았습니다: "+tc.key+"\n")
			assert.NotContains(t, run.out, "구성을 적용했습니다.")
			assert.Equal(t, sum, reviewFileSum(t, dir))
			assert.True(t, reflect.DeepEqual(before.Orchestra, run.cfg.Orchestra))
			assert.True(t, reflect.DeepEqual(before.Spec.ReviewGate, run.cfg.Spec.ReviewGate))
		})
	}
}

// S12, RH through the tab: two value spans change and nothing else.
func TestReviewTab_S12HandWrittenFileThroughTheTab(t *testing.T) {
	dir := writeReviewRH(t)
	before := readReviewFile(t, dir)

	run := runReviewForms(t, dir, nil, reviewLines(reviewOpen, []string{"\n", "\n", "2\n", "\n", "3\n", "\n", "\n", "\n", "2\n", "\n"})...)

	require.NoError(t, run.err, run.out)
	assert.Equal(t, [][2]string{
		{"    providers: [claude, codex, gemini] # reviewers", "    providers: [claude, codex] # reviewers"},
		{"      model: openai-codex/gpt-6-astra:max", "      model: openai-codex/gpt-6-luna:high"},
	}, reviewFileDiff(t, before, readReviewFile(t, dir)))
}

// REQ-RVM-12 for `auto quality`: the same writer, from quality's menu.
func TestQualityComposeReview_TabEditWritesTheFile(t *testing.T) {
	newComposeUpdateRun(t, true, "")
	useAccessibleComposeForms(t)
	dir := writeReviewR3(t)
	runner := &ompCLIFakeRunner{catalog: []byte(reviewCatalogCATG)}
	deps := ompPlatformDependencies{newRunner: func() omp.OMPModelCatalogRunner { return runner }}

	out, err := runQualityForms(t, dir, &deps, formInput(reviewLines([]string{"2\n", "5\n"}, reviewTurnOffGemini)...))

	require.NoError(t, err, out)
	assert.Contains(t, out, "구성을 바꿀까요? \n1. 유지\n2. 구성 편집\n3. 기본 프리셋 고르기\n")
	assert.Contains(t, out, reviewTabR3)
	assert.Contains(t, out, "구성을 적용했습니다.\n")
	assert.Equal(t, []string{"claude", "codex"}, reviewLoaderValue(t, dir).Spec.ReviewGate.Providers)
}
