package tui

import "github.com/charmbracelet/huh"

// NavSelect is a select on value whose first frame lists its options from
// the first one with the cursor on *value; in a form too short for them huh
// scrolls to the cursor row. huh's Options scrolls the select to the option
// equal to the field's value at that moment and Value moves only the
// cursor, so options given before the value meet huh's zero value: a select
// listing T's zero value below its first option opened scrolled to it, the
// rows above it hidden (the review tab's bool selects list true first, the
// cursor on it). Options given after the value would open scrolled to the
// cursor instead. The options therefore go in on the first option's value,
// and Value then sets the cursor.
func NavSelect[T comparable](options []huh.Option[T], value *T) *huh.Select[T] {
	top := *value
	if len(options) > 0 {
		top = options[0].Value
	}
	return huh.NewSelect[T]().Value(&top).Options(options...).Value(value)
}
