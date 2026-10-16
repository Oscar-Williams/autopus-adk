package cli

// SPEC-TUIUX-001 test helpers (REQ-TUX-16): the composition step's form path
// with every form run through the accessible navigation helper, so a test
// reads the frames as plain lines and goes back with a `b` line.

import (
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

// useNavComposeForms makes the composition step take its form path and run
// each form with tui.RunNavAccessible on the command's streams. It returns
// the frames of the forms run so far.
func useNavComposeForms(t *testing.T) *[]tui.NavFrame {
	t.Helper()
	return useNavComposeFormsWith(t, nil)
}

// useNavComposeFormsWith is useNavComposeForms with a hook that sees each
// form, counted from 1, before it runs; an error it returns is the form's.
func useNavComposeFormsWith(t *testing.T, before func(run int, frame tui.NavFrame) error) *[]tui.NavFrame {
	t.Helper()
	// The forms stand for an interactive terminal: a CI value of the test
	// run must not hide the review tab (SPEC-REVMODEL-001 REQ-RVM-01).
	t.Setenv("CI", "")
	originalProbe, originalRun := composeFormsProbe, composeRunForm
	t.Cleanup(func() { composeFormsProbe, composeRunForm = originalProbe, originalRun })
	frames := &[]tui.NavFrame{}
	composeFormsProbe = func(*cobra.Command) bool { return true }
	composeRunForm = func(cmd *cobra.Command, form *huh.Form, frame tui.NavFrame) error {
		*frames = append(*frames, frame)
		if before != nil {
			if err := before(len(*frames), frame); err != nil {
				return err
			}
		}
		return tui.RunNavAccessible(form, frame, cmd.InOrStdin(), cmd.OutOrStdout())
	}
	return frames
}
