package cli

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// The token scanners of the review line splice writer. Each works on one
// line without its line ending and reports false for a token that does not
// end on that line.

// reviewScalarEnd returns the end of the scalar token that starts at start:
// a quoted one at its closing quote, a plain one before an inline comment
// or the end of the line, trailing blanks excluded.
func reviewScalarEnd(body string, start int) (int, bool) {
	if start < len(body) && (body[start] == '\'' || body[start] == '"') {
		return reviewQuotedEnd(body, start)
	}
	end := len(body)
	for i := start + 1; i < len(body); i++ {
		if body[i] == '#' && (body[i-1] == ' ' || body[i-1] == '\t') {
			end = i
			break
		}
	}
	return start + len(strings.TrimRight(body[start:end], " \t")), true
}

// reviewQuotedEnd returns the end of the single- or double-quoted scalar at
// start: a doubled single quote is a quote inside single quotes, and a
// backslash escapes the next character inside double ones.
func reviewQuotedEnd(body string, start int) (int, bool) {
	quote := body[start]
	for i := start + 1; i < len(body); i++ {
		switch {
		case quote == '"' && body[i] == '\\':
			i++
		case body[i] == quote && quote == '\'' && i+1 < len(body) && body[i+1] == '\'':
			i++
		case body[i] == quote:
			return i + 1, true
		}
	}
	return 0, false
}

// reviewFlowEnd returns the end of the flow collection whose `[` is at
// start. A quote opens a quoted scalar only where a token starts; inside a
// plain scalar it is a character.
func reviewFlowEnd(body string, start int) (int, bool) {
	depth, tokenStart := 0, false
	for i := start; i < len(body); i++ {
		switch c := body[i]; {
		case c == '[' || c == '{':
			depth++
			tokenStart = true
		case c == ']' || c == '}':
			if depth--; depth == 0 {
				return i + 1, true
			}
		case c == ',':
			tokenStart = true
		case c == ' ' || c == '\t':
		case c == '#' && (body[i-1] == ' ' || body[i-1] == '\t'):
			return 0, false
		case (c == '\'' || c == '"') && tokenStart:
			end, ok := reviewQuotedEnd(body, i)
			if !ok {
				return 0, false
			}
			i, tokenStart = end-1, false
		default:
			tokenStart = false
		}
	}
	return 0, false
}

// reviewRestIsComment reports a line rest that holds only blanks and an
// optional comment.
func reviewRestIsComment(rest string) bool {
	rest = strings.TrimLeft(rest, " \t")
	return rest == "" || strings.HasPrefix(rest, "#")
}

func splitReviewLineEnding(line string) (body, ending string) {
	body = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	return body, line[len(body):]
}

// reviewColumnOffset turns a yaml.v3 column, which counts runes from 1, into
// a byte offset of body; one past the last rune is the end of the line.
func reviewColumnOffset(body string, column int) int {
	count := 1
	for offset := range body {
		if count == column {
			return offset
		}
		count++
	}
	if count == column {
		return len(body)
	}
	return -1
}

// yamlAliasTargets returns every node an alias of the document names.
func yamlAliasTargets(node *yaml.Node) map[*yaml.Node]bool {
	targets := map[*yaml.Node]bool{}
	var walk func(*yaml.Node)
	walk = func(node *yaml.Node) {
		if node.Kind == yaml.AliasNode {
			targets[node.Alias] = true
			return
		}
		for _, child := range node.Content {
			walk(child)
		}
	}
	walk(node)
	return targets
}

func yamlMappingHasMerge(mapping *yaml.Node) bool {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if key := mapping.Content[i]; key.Kind == yaml.ScalarNode && key.ShortTag() == "!!merge" {
			return true
		}
	}
	return false
}

// yamlMaxLine is the last line any node of the subtree starts on; an alias
// is a leaf.
func yamlMaxLine(node *yaml.Node) int {
	line := node.Line
	for _, child := range node.Content {
		line = max(line, yamlMaxLine(child))
	}
	return line
}
