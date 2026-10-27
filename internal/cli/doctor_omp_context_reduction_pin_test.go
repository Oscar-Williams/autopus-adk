package cli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two hand-edited copies hold one verdict: ompContextReductionVerified here and
// the case arms of scripts/release-tools/advance-omp-pin.sh. A cohort result
// recorded in one and not the other reaches an operator as a contradiction:
// doctor calls the installed OMP measured good while the pin script refuses to
// move to it without --measure.
//
// The script is run, not parsed. Its verdict table runs before the asset
// download, so with gh stubbed to fail, an admitted version stops at
// "cannot download" and a refused one never reaches it.
func TestOMPContextReductionVerified_AdmittedByThePinScriptWithoutMeasure(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, ompContextReductionVerified)
	for _, version := range ompContextReductionVerified {
		target := strings.TrimPrefix(version, "omp/")
		output := runOMPPinScriptVerdict(t, target)
		assert.Contains(t, output, "cannot download omp-darwin-arm64 for v"+target,
			"doctor reports %s as measured good, so advance-omp-pin.sh must admit it without --measure", version)
		assert.NotContains(t, output, "has not been measured", version)
	}
}

// The control that makes the test above load-bearing: a measured row admits
// that exact version and nothing near it, so "cannot download" is reachable
// only through the verdict table.
func TestOMPPinScript_RefusesAnUnmeasuredNeighbourOfTheMeasuredPin(t *testing.T) {
	t.Parallel()

	output := runOMPPinScriptVerdict(t, "18.12.4")
	assert.Contains(t, output, "omp/18.12.4 has not been measured against the reduction floor")
	assert.NotContains(t, output, "cannot download")
	assert.Equal(t, "reduction_unmeasured", ompContextReductionReason("omp/18.12.4"))
}

// runOMPPinScriptVerdict runs advance-omp-pin.sh without a mode flag from a
// sandbox whose declared pin is a version no verdict names, with gh replaced by
// a stub that always fails, and returns the combined output of the refusal.
func runOMPPinScriptVerdict(t *testing.T, target string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("advance-omp-pin.sh is a bash release tool")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "release-tools", "advance-omp-pin.sh"))
	require.NoError(t, err)
	sandbox := t.TempDir()
	declaration := filepath.Join(sandbox, "scripts", "companion-release", "prepare-release.sh")
	require.NoError(t, os.MkdirAll(filepath.Dir(declaration), 0o700))
	require.NoError(t, os.WriteFile(declaration,
		[]byte(`[[ "$("$staged_omp" --version)" == 'omp/0.0.1' ]] || fail 'version'`+"\n"), 0o600))
	bin := filepath.Join(sandbox, "bin")
	require.NoError(t, os.MkdirAll(bin, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nexit 1\n"), 0o700))

	command := exec.CommandContext(t.Context(), "/bin/bash", script, target)
	command.Dir = sandbox
	command.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "HOME=" + sandbox, "TMPDIR=" + sandbox}
	output, err := command.CombinedOutput()
	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr),
		"advance-omp-pin.sh %s must stop before moving a pin: err=%v\n%s", target, err, output)
	require.Contains(t, string(output), "advance omp pin: ", "the script never ran its own checks")
	return string(output)
}
