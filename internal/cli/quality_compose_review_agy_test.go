package cli

// SPEC-REVMODEL-001 T14 (REQ-RVM-24; Agy Catalog Contract): `agy models` is
// read strictly, bounded, and once per tab entry; its ids are filtered by
// the provider's family and grouped by effort suffix.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reviewAgySample is AGY: the observed agy 1.4.0 sample (2026-10-16).
const reviewAgySample = "Fetching available models...\n" +
	"gemini-3.8-flash-high\tGemini 3.8 Flash (High)\n" +
	"gemini-4-argon-medium\tGemini 4 Argon (Medium)\n" +
	"claude-opus-5-5-low\tClaude Opus 5.5 (Low)\n" +
	"gpt-oss-120b-medium\tGPT-OSS 120B (Medium)\n"

// reviewAgyUnreadable are S23(3)'s six variants.
var reviewAgyUnreadable = map[string]struct {
	out string
	err error
}{
	"line without a tab": {out: reviewAgySample + "gemini-x-high Gemini X\n"},
	"upper-case id":      {out: reviewAgySample + "Gemini-x-high\tGemini X\n"},
	"ESC in a name":      {out: reviewAgySample + "gemini-x-high\tGemini \x1b[2JX\n"},
	"header only":        {out: "Fetching available models...\n"},
	"exit 1":             {err: errors.New("exit status 1")},
	"over 64 KiB":        {out: reviewAgySample + strings.Repeat("gemini-x-high\tGemini X\n", agyModelsMaxOutput/20)},
}

// No CLI test runs a real agy: in this test binary `agy models` fails unless
// a test installs a list with useAgyModels, whatever PATH a test keeps.
func init() {
	agyModelsRunner = func(context.Context, string) ([]byte, error) {
		return nil, errors.New("agy models does not run in tests")
	}
}

// useAgyModels answers every `agy models` run with out and err and counts
// the runs.
func useAgyModels(t *testing.T, out string, err error) *int {
	t.Helper()
	original := agyModelsRunner
	t.Cleanup(func() { agyModelsRunner = original })
	runs := new(int)
	agyModelsRunner = func(context.Context, string) ([]byte, error) {
		*runs++
		return []byte(out), err
	}
	return runs
}

func TestParseAgyModels_StrictLines(t *testing.T) {
	models, ok := parseAgyModels([]byte(reviewAgySample))
	require.True(t, ok)
	assert.Equal(t, []agyModel{
		{"gemini-3.8-flash-high", "Gemini 3.8 Flash (High)"}, {"gemini-4-argon-medium", "Gemini 4 Argon (Medium)"},
		{"claude-opus-5-5-low", "Claude Opus 5.5 (Low)"}, {"gpt-oss-120b-medium", "GPT-OSS 120B (Medium)"},
	}, models)
	for name, variant := range reviewAgyUnreadable {
		if variant.err != nil {
			continue
		}
		_, ok := parseAgyModels([]byte(variant.out))
		assert.False(t, ok, name)
	}
	_, ok = parseAgyModels([]byte("Fetching available models... \ngemini-x\tX\n"))
	assert.False(t, ok, "a header that is not exact is a malformed line")
}

func TestAgyCLIModels_FamilyAndGroups(t *testing.T) {
	listed, _ := parseAgyModels([]byte(reviewAgySample + "gemini-3.8-flash-low\tGemini 3.8 Flash (Low)\n"))

	models := agyCLIModels("gemini", listed)

	assert.Equal(t, []reviewCLIModel{
		{ID: "gemini-3.8-flash-high", Base: "gemini-3.8-flash", Suffix: "high", Display: "Gemini 3.8 Flash (High)"},
		{ID: "gemini-3.8-flash-low", Base: "gemini-3.8-flash", Suffix: "low", Display: "Gemini 3.8 Flash (Low)"},
		{ID: "gemini-4-argon-medium", Base: "gemini-4-argon", Suffix: "medium", Display: "Gemini 4 Argon (Medium)"},
	}, models)
	assert.Equal(t, []string{"claude-opus-5-5-low"}, reviewSelectorsOfCLI(agyCLIModels("claude", listed)))
}

func reviewSelectorsOfCLI(models []reviewCLIModel) []string {
	ids := make([]string, len(models))
	for i, model := range models {
		ids[i] = model.ID
	}
	return ids
}

// The production runner runs only a name found on PATH, keeps stdout, fails
// on a non-zero exit, stops an output past 64 KiB, and stops a run past its
// deadline. Its stand-in is a shell script, never a real agy.
func TestRunAgyModels_BoundsTheProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in is a POSIX shell script")
	}
	bin := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	script := func(name, body string) string {
		require.NoError(t, os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755))
		return name
	}
	marker := filepath.Join(t.TempDir(), "ran")
	_, err := runAgyModels(context.Background(), filepath.Join(bin, script("path", ": > "+marker)))
	assert.Error(t, err, "a path is never run")
	assert.NoFileExists(t, marker)
	out, err := runAgyModels(context.Background(), script("ok", `[ "$1" = models ] && printf 'Fetching available models...\ngemini-x-high\tX\n' && echo noise >&2`))
	require.NoError(t, err)
	assert.Equal(t, "Fetching available models...\ngemini-x-high\tX\n", string(out))
	_, err = runAgyModels(context.Background(), script("fails", "exit 1"))
	assert.Error(t, err)
	_, err = runAgyModels(context.Background(), script("floods", "yes gemini-x-high | head -c 70000"))
	assert.Error(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = runAgyModels(ctx, script("hangs", "exec sleep 10"))
	assert.Error(t, err)
	assert.Less(t, time.Since(started), 5*time.Second, "a run past its deadline is stopped")
	assert.Equal(t, 15*time.Second, agyModelsTimeout)
	assert.Equal(t, 64<<10, agyModelsMaxOutput)
}

// One read per tab entry and binary, whatever the passes ask.
func TestReviewAgyCatalog_ReadsOncePerBinary(t *testing.T) {
	runs := useAgyModels(t, reviewAgySample, nil)
	catalog := reviewAgyCatalog{}

	first, ok := catalog.load(context.Background(), "agy")
	again, okAgain := catalog.load(context.Background(), "agy")

	assert.True(t, ok)
	assert.True(t, okAgain)
	assert.Equal(t, first, again)
	assert.Equal(t, 1, *runs)
	catalog.load(context.Background(), "/opt/agy")
	assert.Equal(t, 2, *runs)
}

// L2 on every OS: only a name for PATH runs. Windows reads `C:agy` as a path
// relative to the drive's working directory, so a `:` is a path like either
// separator; an empty or `.`-prefixed name never runs either.
func TestAgyBareName_OnlyANameForPATH(t *testing.T) {
	for _, binary := range []string{"agy", "agy.exe", "agy-1.4"} {
		assert.True(t, agyBareName(binary), binary)
	}
	for _, binary := range []string{"", "C:agy", "c:agy.exe", "./agy", "tools/agy", `tools\agy`, ".agy", ".", ".."} {
		assert.False(t, agyBareName(binary), binary)
	}
}
