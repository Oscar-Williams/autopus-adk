package tui_test

// SPEC-TUIUX-001 T4 (REQ-TUX-02, REQ-TUX-10): the back stack keeps the state
// saved before each form was first shown and the summary of the forms
// before the current one.

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/insajin/autopus-adk/internal/cli/tui"
)

func TestNavHistory_BackRestoresTheStateBeforeTheCurrentForm(t *testing.T) {
	var h tui.NavHistory[int]
	h.Show(0)
	assert.True(t, h.First())
	_, ok := h.Back()
	assert.False(t, ok, "a first form has nothing to go back to")
	assert.Equal(t, 1, h.Len())

	h.Answer("  one: 1")
	h.Show(1)
	h.Answer("  two: 2")
	h.Show(2)
	assert.False(t, h.First())
	assert.Equal(t, []string{"  one: 1", "  two: 2"}, h.Summary())

	state, ok := h.Back()
	assert.True(t, ok)
	assert.Equal(t, 2, state, "the state saved before the form left")
	current, _ := h.Current()
	assert.Equal(t, 1, current, "the previous form keeps the state saved before its first showing")
	assert.Equal(t, []string{"  one: 1"}, h.Summary())

	// Going back again returns the state before that form's first showing,
	// so an answer it took on the way forward is gone.
	state, _ = h.Back()
	assert.Equal(t, 1, state)
	assert.Nil(t, h.Summary())
	assert.True(t, h.First())

	// Answering again puts the new line in the summary of the next form.
	h.Answer("  one: 3")
	h.Show(3)
	assert.Equal(t, []string{"  one: 3"}, h.Summary())
}

func TestNavHistory_EmptyLinesAreNoSummaryLines(t *testing.T) {
	var h tui.NavHistory[string]
	h.Show("list")
	h.Answer("")
	h.Show("value")
	h.Answer("  planner: opus")
	h.Show("effort")
	assert.Equal(t, []string{"  planner: opus"}, h.Summary())
	_, ok := (&tui.NavHistory[string]{}).Current()
	assert.False(t, ok)
	assert.Nil(t, (&tui.NavHistory[string]{}).Summary())
}
