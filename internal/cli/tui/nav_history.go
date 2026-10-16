package tui

// NavHistory is the back stack of one navigation flow (REQ-TUX-02,
// REQ-TUX-10). Every form the flow shows on its way forward gets one entry:
// the state saved immediately before the form was first shown and, once the
// form is answered, its summary line. Going back drops the current form's
// entry and returns its saved state, so the previous form, shown again from
// that state, finds its own earlier answer and is the current form again.
type NavHistory[S any] struct {
	entries []navEntry[S]
}

type navEntry[S any] struct {
	state S
	line  string
}

// Show records a form shown for the first time on the way forward, with the
// state saved before it.
func (h *NavHistory[S]) Show(state S) {
	h.entries = append(h.entries, navEntry[S]{state: state})
}

// Answer records the summary line of the current form; an empty line adds
// none to the summary.
func (h *NavHistory[S]) Answer(line string) {
	if len(h.entries) > 0 {
		h.entries[len(h.entries)-1].line = line
	}
}

// Back leaves the current form. It returns the state saved before that form
// and makes the previous form current again, its line out of the summary
// until it is answered again. On the first form it returns false and changes
// nothing.
func (h *NavHistory[S]) Back() (S, bool) {
	var zero S
	if len(h.entries) < 2 {
		return zero, false
	}
	top := h.entries[len(h.entries)-1]
	h.entries = h.entries[:len(h.entries)-1]
	h.entries[len(h.entries)-1].line = ""
	return top.state, true
}

// Current returns the state saved before the current form.
func (h *NavHistory[S]) Current() (S, bool) {
	var zero S
	if len(h.entries) == 0 {
		return zero, false
	}
	return h.entries[len(h.entries)-1].state, true
}

// Len is the number of forms on the stack, the current one included.
func (h *NavHistory[S]) Len() int {
	return len(h.entries)
}

// First reports whether the current form is the flow's first.
func (h *NavHistory[S]) First() bool {
	return len(h.entries) <= 1
}

// Summary lists the lines of the forms before the current one, in the
// order they were answered.
func (h *NavHistory[S]) Summary() []string {
	var lines []string
	for _, entry := range h.entries[:max(len(h.entries)-1, 0)] {
		if entry.line != "" {
			lines = append(lines, entry.line)
		}
	}
	return lines
}
