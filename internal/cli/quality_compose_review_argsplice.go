package cli

import (
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/content"
)

// reviewArgsValues is the raw value of every args item, unexpanded. An item
// that is not a scalar gets a value no flag spelling equals, so it is never
// read as a flag and stays as it is.
func reviewArgsValues(value *yaml.Node) []string {
	values := make([]string, len(value.Content))
	for i, item := range value.Content {
		values[i] = "\x00"
		if item.Kind == yaml.ScalarNode {
			values[i] = item.Value
		}
	}
	return values
}

// reviewArgsTargetsSafe reports whether every item edit touches is a plain
// one-line scalar without an anchor or a placeholder (REQ-RVM-25, 26).
func reviewArgsTargetsSafe(value *yaml.Node, edit reviewArgsEdit) bool {
	touched := slices.Clone(edit.remove)
	for index := range edit.set {
		touched = append(touched, index)
	}
	for _, index := range touched {
		item := value.Content[index]
		if item.Kind != yaml.ScalarNode || item.Anchor != "" ||
			item.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 || strings.Contains(item.Value, "${") {
			return false
		}
	}
	return true
}

// reviewFlowListText encodes values as a flow list.
func reviewFlowListText(values []string) string {
	encoded := make([]string, len(values))
	for i, value := range values {
		encoded[i] = content.OMPYAMLScalar(value)
	}
	return "[" + strings.Join(encoded, ", ") + "]"
}

// spliceReviewArgs turns edit into raw text edits of a non-empty args list. A
// flow list on one line gets in-place token replacements, removals with
// their separators, and elements appended before its `]`; a block list gets
// item lines replaced, removed, or appended at the items' indentation. Every
// other byte stays.
func spliceReviewArgs(value *yaml.Node, lines []string, newline string, edit reviewArgsEdit) ([]reviewLineEdit, bool) {
	if value.Line > len(lines) || !reviewArgsTargetsSafe(value, edit) {
		return nil, false
	}
	if value.Style&yaml.FlowStyle != 0 {
		edited, ok := spliceReviewFlowArgs(value, lines[value.Line-1], edit)
		if !ok {
			return nil, false
		}
		return []reviewLineEdit{{first: value.Line - 1, last: value.Line - 1, lines: []string{edited}}}, true
	}
	return spliceReviewBlockArgs(value, lines, newline, edit)
}

// reviewTextEdit replaces text[start:end].
type reviewTextEdit struct {
	start, end int
	text       string
}

func spliceReviewFlowArgs(value *yaml.Node, line string, edit reviewArgsEdit) (string, bool) {
	body, ending := splitReviewLineEnding(line)
	open := reviewColumnOffset(body, value.Column)
	if open < 0 || open >= len(body) || body[open] != '[' {
		return "", false
	}
	end, ok := reviewFlowEnd(body, open)
	if !ok || !reviewRestIsComment(body[end:]) {
		return "", false
	}
	spans := make([][2]int, len(value.Content))
	for i, item := range value.Content {
		start := reviewColumnOffset(body, item.Column)
		if item.Line != value.Line || start < 0 {
			return "", false
		}
		stop, ok := reviewFlowScalarEnd(body, start)
		if !ok {
			return "", false
		}
		spans[i] = [2]int{start, stop}
	}
	var edits []reviewTextEdit
	for index, text := range edit.set {
		edits = append(edits, reviewTextEdit{spans[index][0], spans[index][1], content.OMPYAMLScalar(text)})
	}
	removed := slices.Sorted(slices.Values(edit.remove))
	for i := 0; i < len(removed); {
		first, last := removed[i], removed[i]
		for i++; i < len(removed) && removed[i] == last+1; i++ {
			last = removed[i]
		}
		switch {
		case first > 0:
			edits = append(edits, reviewTextEdit{spans[first-1][1], spans[last][1], ""})
		case last+1 < len(spans):
			edits = append(edits, reviewTextEdit{spans[first][0], spans[last+1][0], ""})
		default:
			edits = append(edits, reviewTextEdit{spans[first][0], spans[last][1], ""})
		}
	}
	if len(edit.add) > 0 {
		added := strings.TrimSuffix(strings.TrimPrefix(reviewFlowListText(edit.add), "["), "]")
		if len(spans) > len(removed) {
			at := spans[len(spans)-1][1]
			edits = append(edits, reviewTextEdit{at, at, ", " + added})
		} else {
			edits = append(edits, reviewTextEdit{open + 1, open + 1, added})
		}
	}
	// Right to left; at one position a removal goes before an insertion.
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start > edits[j].start
		}
		return edits[i].end > edits[j].end
	})
	for _, e := range edits {
		body = body[:e.start] + e.text + body[e.end:]
	}
	return body + ending, true
}

func spliceReviewBlockArgs(value *yaml.Node, lines []string, newline string, edit reviewArgsEdit) ([]reviewLineEdit, bool) {
	items := value.Content
	if len(items) == 0 || items[len(items)-1].Line > len(lines) {
		return nil, false
	}
	first, _ := splitReviewLineEnding(lines[items[0].Line-1])
	dash := reviewColumnOffset(first, value.Column)
	if dash < 0 || strings.TrimLeft(first[:dash], " ") != "" || !strings.HasPrefix(first[dash:], "-") {
		return nil, false
	}
	var edits []reviewLineEdit
	for index, text := range edit.set {
		item := items[index]
		body, ending := splitReviewLineEnding(lines[item.Line-1])
		start := reviewColumnOffset(body, item.Column)
		if start < 0 {
			return nil, false
		}
		stop, ok := reviewScalarEnd(body, start)
		if !ok || !reviewRestIsComment(body[stop:]) {
			return nil, false
		}
		line := body[:start] + content.OMPYAMLScalar(text) + body[stop:] + ending
		edits = append(edits, reviewLineEdit{first: item.Line - 1, last: item.Line - 1, lines: []string{line}})
	}
	for _, index := range edit.remove {
		at := items[index].Line - 1
		edits = append(edits, reviewLineEdit{first: at, last: at})
	}
	if len(edit.add) > 0 {
		added := make([]string, len(edit.add))
		for i, text := range edit.add {
			added[i] = first[:dash] + "- " + content.OMPYAMLScalar(text) + newline
		}
		last := items[len(items)-1].Line - 1
		edits = append(edits, reviewLineEdit{first: last + 1, last: last, lines: added})
	}
	return edits, true
}

// reviewFlowScalarEnd returns the end of the flow scalar at start: a quoted
// one at its closing quote, a plain one before `,`, `]`, `}`, or an inline
// comment, trailing blanks excluded.
func reviewFlowScalarEnd(body string, start int) (int, bool) {
	if start < len(body) && (body[start] == '\'' || body[start] == '"') {
		return reviewQuotedEnd(body, start)
	}
	end := len(body)
	for i := start; i < len(body); i++ {
		if c := body[i]; c == ',' || c == ']' || c == '}' || c == '#' && i > start && (body[i-1] == ' ' || body[i-1] == '\t') {
			end = i
			break
		}
	}
	return start + len(strings.TrimRight(body[start:end], " \t")), true
}
