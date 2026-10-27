package cli

import (
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/content"
)

// reviewLineEdit replaces lines[first:last+1] with lines; last < first is
// an insertion before first.
type reviewLineEdit struct {
	first, last int
	lines       []string
}

// spliceReviewScalar replaces the value token on its key's line: plain,
// '…', or "…", keeping a leading tag and the inline comment.
func spliceReviewScalar(key, value *yaml.Node, lines []string, replacement string) (reviewLineEdit, bool) {
	return spliceReviewScalarText(key, value, lines, content.OMPYAMLScalar(replacement))
}

// spliceReviewScalarText is spliceReviewScalar for text already encoded,
// such as a flow list for a null args value.
func spliceReviewScalarText(key, value *yaml.Node, lines []string, text string) (reviewLineEdit, bool) {
	if value.Kind != yaml.ScalarNode || value.Line != key.Line || value.Line > len(lines) ||
		value.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 || strings.Contains(value.Value, "${") {
		return reviewLineEdit{}, false
	}
	index := value.Line - 1
	body, ending := splitReviewLineEnding(lines[index])
	start := reviewColumnOffset(body, value.Column)
	if start < 0 {
		return reviewLineEdit{}, false
	}
	properties := qualityValuePropertiesRE.FindString(body[start:])
	end, ok := reviewScalarEnd(body, start+len(properties))
	if !ok || !reviewRestIsComment(body[end:]) {
		return reviewLineEdit{}, false
	}
	head := body[:start] + properties
	if !strings.HasSuffix(head, " ") && !strings.HasSuffix(head, "\t") {
		head += " "
	}
	return reviewLineEdit{first: index, last: index, lines: []string{head + text + body[end:] + ending}}, true
}

// spliceReviewList replaces a list of plain one-line scalars: a flow list
// from its `[` to the matching `]` on the same line, or the item lines of a
// block list with `- <name>` lines at the same indentation.
func spliceReviewList(value *yaml.Node, lines []string, newline string, names []string) (reviewLineEdit, bool) {
	if value.Kind != yaml.SequenceNode || value.Line > len(lines) {
		return reviewLineEdit{}, false
	}
	for _, item := range value.Content {
		if item.Kind != yaml.ScalarNode || item.Anchor != "" ||
			item.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0 || strings.Contains(item.Value, "${") {
			return reviewLineEdit{}, false
		}
	}
	items := make([]string, len(names))
	for i, name := range names {
		items[i] = content.OMPYAMLScalar(name)
	}
	if value.Style&yaml.FlowStyle != 0 {
		index := value.Line - 1
		body, ending := splitReviewLineEnding(lines[index])
		start := reviewColumnOffset(body, value.Column)
		if start < 0 || start >= len(body) || body[start] != '[' {
			return reviewLineEdit{}, false
		}
		end, ok := reviewFlowEnd(body, start)
		if !ok || !reviewRestIsComment(body[end:]) {
			return reviewLineEdit{}, false
		}
		line := body[:start] + "[" + strings.Join(items, ", ") + "]" + body[end:] + ending
		return reviewLineEdit{first: index, last: index, lines: []string{line}}, true
	}
	if len(value.Content) == 0 {
		return reviewLineEdit{}, false
	}
	first, last := value.Content[0].Line-1, value.Content[len(value.Content)-1].Line-1
	if first < 0 || last >= len(lines) || first > last {
		return reviewLineEdit{}, false
	}
	for _, line := range lines[first+1 : last] {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			return reviewLineEdit{}, false
		}
	}
	body, _ := splitReviewLineEnding(lines[first])
	dash := reviewColumnOffset(body, value.Column)
	if dash < 0 || strings.TrimLeft(body[:dash], " ") != "" || !strings.HasPrefix(body[dash:], "-") {
		return reviewLineEdit{}, false
	}
	_, lastEnding := splitReviewLineEnding(lines[last])
	replaced := make([]string, len(items))
	for i, item := range items {
		ending := newline
		if i == len(items)-1 {
			ending = lastEnding
		}
		replaced[i] = body[:dash] + "- " + item + ending
	}
	return reviewLineEdit{first: first, last: last, lines: replaced}, true
}

// insertReviewKeys adds one `<key>: <value>` line per text after the last
// line of mapping, at the indentation of its keys, in the given order.
func insertReviewKeys(mapping *yaml.Node, lines []string, newline string, texts []string) (reviewLineEdit, bool) {
	if len(mapping.Content) == 0 || mapping.Content[0].Line > len(lines) {
		return reviewLineEdit{}, false
	}
	body, _ := splitReviewLineEnding(lines[mapping.Content[0].Line-1])
	indent := reviewColumnOffset(body, mapping.Column)
	if indent < 0 || strings.TrimLeft(body[:indent], " ") != "" {
		return reviewLineEdit{}, false
	}
	last := yamlMaxLine(mapping) - 1
	// Continuation lines of the last value, a block scalar's for one, are
	// indented deeper than the mapping's keys.
	for last+1 < len(lines) {
		next, _ := splitReviewLineEnding(lines[last+1])
		if strings.TrimSpace(next) == "" || len(next)-len(strings.TrimLeft(next, " \t")) <= indent {
			break
		}
		last++
	}
	added := make([]string, len(texts))
	for i, text := range texts {
		added[i] = body[:indent] + text + newline
	}
	return reviewLineEdit{first: last + 1, last: last, lines: added}, true
}

// applyReviewLineEdits applies edits from the bottom up, so every earlier
// line index stays valid; at one index a replacement goes before an
// insertion.
func applyReviewLineEdits(lines []string, edits []reviewLineEdit, newline string) []byte {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].first != edits[j].first {
			return edits[i].first > edits[j].first
		}
		return edits[i].last >= edits[j].last
	})
	for _, edit := range edits {
		if edit.last < edit.first && edit.first > 0 && !strings.HasSuffix(lines[edit.first-1], "\n") {
			lines[edit.first-1] += newline
		}
		lines = slices.Concat(lines[:edit.first], edit.lines, lines[edit.last+1:])
	}
	return []byte(strings.Join(lines, ""))
}
