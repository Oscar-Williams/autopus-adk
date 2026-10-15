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

// spliceReviewTargets replaces the value span of every target in source by
// the Line Splice Writer Contract, locating each through the raw document
// (no environment expansion) and encoding new values with OMPYAMLScalar. A
// target it cannot edit safely refuses the whole edit with its path.
func spliceReviewTargets(source []byte, targets []reviewTarget) ([]byte, error) {
	// A line break yaml.v3 sees but a split at \n does not moves node lines
	// away from the lines the splice cuts, as for the quality writers.
	if hasYAMLOnlyLineBreak(source) {
		return nil, reviewWriteRefusal{key: targets[0].key()}
	}
	document, _, err := parseAutopusConfigDocument(source)
	if err != nil {
		return nil, reviewWriteRefusal{key: targets[0].key()}
	}
	aliased := yamlAliasTargets(document)
	lines := strings.SplitAfter(string(source), "\n")
	newline := detectConfigNewline(source)
	edits := make([]reviewLineEdit, 0, len(targets))
	for _, target := range targets {
		edit, ok := planReviewSpan(document.Content[0], aliased, lines, newline, target)
		if !ok {
			return nil, reviewWriteRefusal{key: target.key()}
		}
		edits = append(edits, edit)
	}
	return applyReviewLineEdits(lines, edits, newline), nil
}

// planReviewSpan finds one target's span. Every mapping on the way must be a
// block mapping no alias names, so nothing else in the file shares what the
// edit changes; the parent may hold no merge key; the key and value carry
// no anchor or alias. Only a missing judge is added.
func planReviewSpan(root *yaml.Node, aliased map[*yaml.Node]bool, lines []string, newline string, target reviewTarget) (reviewLineEdit, bool) {
	node := root
	for depth, step := range target.path {
		if node.Kind != yaml.MappingNode || node.Style&yaml.FlowStyle != 0 || aliased[node] {
			return reviewLineEdit{}, false
		}
		key, value := yamlMappingPair(node, step)
		if depth < len(target.path)-1 {
			if key == nil || value == nil {
				return reviewLineEdit{}, false
			}
			node = value
			continue
		}
		switch {
		case yamlMappingHasMerge(node):
			return reviewLineEdit{}, false
		case key == nil && step == "judge":
			return insertReviewJudge(node, lines, newline, target.value)
		case key == nil || key.Anchor != "" || value.Anchor != "" || value.Kind == yaml.AliasNode || aliased[value]:
			return reviewLineEdit{}, false
		case target.isList:
			return spliceReviewList(value, lines, newline, target.list)
		}
		return spliceReviewScalar(key, value, lines, target.value)
	}
	return reviewLineEdit{}, false
}

// spliceReviewScalar replaces the value token on its key's line: plain,
// '…', or "…", keeping a leading tag and the inline comment.
func spliceReviewScalar(key, value *yaml.Node, lines []string, replacement string) (reviewLineEdit, bool) {
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
	line := head + content.OMPYAMLScalar(replacement) + body[end:] + ending
	return reviewLineEdit{first: index, last: index, lines: []string{line}}, true
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

// insertReviewJudge adds `judge: <j>` after the last line of the review_gate
// mapping, at the indentation of its keys.
func insertReviewJudge(mapping *yaml.Node, lines []string, newline, judge string) (reviewLineEdit, bool) {
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
	line := body[:indent] + "judge: " + content.OMPYAMLScalar(judge) + newline
	return reviewLineEdit{first: last + 1, last: last, lines: []string{line}}, true
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
