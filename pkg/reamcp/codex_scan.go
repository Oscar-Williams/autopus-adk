package reamcp

import (
	"strconv"
	"strings"
)

// The scanner below decides headers with the rule of the Codex adapter's
// structural config merge (parseStructuralCodexSection in pkg/adapter/codex),
// which this package cannot import, but reads strings by TOML's own rules:
// one-line basic and literal strings, escapes included, are skipped before a
// multiline delimiter counts; every line outside a multiline string is read,
// whatever its key looks like; and a line that closes a multiline string is
// read on past the delimiter. A `[` inside any string is therefore never a
// header, including the strings the merge's scanner misses (a `'''` inside a
// one-line string, a dotted key, an array element). Where the two differ this
// one follows TOML, and every Codex text a REA action changes is re-checked
// by the structural validator before it is written.
// pkg/adapter/codex TestReaCodexHeaderDetectionMatchesMerge keeps both on the
// same headers for ordinary TOML.

type tomlScanState struct {
	multiline string // delimiter of the multiline string being read, or ""
}

// skipSyntaxLine reports whether line starts inside a multiline string. When
// the string closes on line, the rest of the line is read for a new one.
func (s *tomlScanState) skipSyntaxLine(line string) bool {
	if s.multiline == "" {
		return false
	}
	end := tomlMultilineEnd(line, 0, s.multiline)
	if end < 0 {
		return true
	}
	s.multiline = tomlOpenMultiline(line[end:])
	return true
}

// observeLine reads a line that starts outside any string and records the
// multiline string it leaves open.
func (s *tomlScanState) observeLine(line string) {
	s.multiline = tomlOpenMultiline(line)
}

// tomlOpenMultiline walks text that starts outside any string and returns the
// delimiter of a multiline string still open at its end, or "".
func tomlOpenMultiline(text string) string {
	for i := 0; i < len(text); {
		switch {
		case text[i] == '#':
			return ""
		case strings.HasPrefix(text[i:], `"""`), strings.HasPrefix(text[i:], `'''`):
			delimiter := text[i : i+3]
			end := tomlMultilineEnd(text, i+3, delimiter)
			if end < 0 {
				return delimiter
			}
			i = end
		case text[i] == '"':
			i = tomlBasicStringEnd(text, i+1)
		case text[i] == '\'':
			end := strings.IndexByte(text[i+1:], '\'')
			if end < 0 {
				return ""
			}
			i += end + 2
		default:
			i++
		}
	}
	return ""
}

// tomlMultilineEnd returns the index just past the delimiter that closes a
// multiline string whose content resumes at from, or -1 when text does not
// close it. A basic string's delimiter may not be escaped, and up to two
// quotes right before the delimiter are content, so a run of four or five
// quotes closes at its end.
func tomlMultilineEnd(text string, from int, delimiter string) int {
	for i := from; i+len(delimiter) <= len(text); i++ {
		if !strings.HasPrefix(text[i:], delimiter) {
			continue
		}
		if delimiter == `"""` && tomlIsEscapedAt(text, i) {
			continue
		}
		end := i + len(delimiter)
		for extra := 0; extra < 2 && end < len(text) && text[end] == delimiter[0]; extra++ {
			end++
		}
		return end
	}
	return -1
}

// tomlBasicStringEnd returns the index just past the quote that closes a
// one-line basic string whose content starts at from, or len(text) when the
// line ends first.
func tomlBasicStringEnd(text string, from int) int {
	for i := from; i < len(text); i++ {
		switch text[i] {
		case '\\':
			i++
		case '"':
			return i + 1
		}
	}
	return len(text)
}

func tomlValueWithoutComment(value string) string {
	const (
		outside = iota
		basic
		literal
		multilineBasic
		multilineLiteral
	)
	state := outside
	escaped := false
	for i := 0; i < len(value); i++ {
		switch state {
		case outside:
			switch {
			case value[i] == '#':
				return value[:i]
			case strings.HasPrefix(value[i:], `"""`):
				state = multilineBasic
				i += 2
			case strings.HasPrefix(value[i:], `'''`):
				state = multilineLiteral
				i += 2
			case value[i] == '"':
				state = basic
			case value[i] == '\'':
				state = literal
			}
		case basic:
			switch {
			case escaped:
				escaped = false
			case value[i] == '\\':
				escaped = true
			case value[i] == '"':
				state = outside
			}
		case literal:
			if value[i] == '\'' {
				state = outside
			}
		case multilineBasic:
			if strings.HasPrefix(value[i:], `"""`) && !tomlIsEscapedAt(value, i) {
				state = outside
				i += 2
			}
		case multilineLiteral:
			if strings.HasPrefix(value[i:], `'''`) {
				state = outside
				i += 2
			}
		}
	}
	return value
}

func tomlIsEscapedAt(value string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && value[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

// tomlStructuralSection mirrors parseStructuralCodexSection: an array table
// keeps its `[[...]]` brackets in the returned name.
func tomlStructuralSection(line string) (string, bool) {
	header := strings.TrimSpace(tomlValueWithoutComment(line))
	if strings.HasPrefix(header, "[[") && strings.HasSuffix(header, "]]") {
		name := strings.TrimSpace(header[2 : len(header)-2])
		return "[[" + name + "]]", name != ""
	}
	if !strings.HasPrefix(header, "[") || !strings.HasSuffix(header, "]") {
		return "", false
	}
	section := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(header, "["), "]"))
	if section == "" || strings.Contains(section, "[") || strings.Contains(section, "]") {
		return "", false
	}
	return section, true
}

// tomlKeyPath splits a TOML dotted key into its unquoted segments, so that
// `mcp_servers."rea"` and `mcp_servers . rea` name the same table. It is used
// only to recognise the rea server and never changes the scan state.
func tomlKeyPath(raw string) ([]string, bool) {
	var path []string
	rest := strings.TrimSpace(raw)
	for {
		if rest == "" {
			return nil, false
		}
		var segment string
		switch rest[0] {
		case '"':
			end := tomlClosingQuote(rest)
			if end < 0 {
				return nil, false
			}
			value, err := strconv.Unquote(rest[:end+1])
			if err != nil {
				return nil, false
			}
			segment, rest = value, rest[end+1:]
		case '\'':
			end := strings.IndexByte(rest[1:], '\'')
			if end < 0 {
				return nil, false
			}
			segment, rest = rest[1:end+1], rest[end+2:]
		default:
			end := strings.IndexByte(rest, '.')
			if end < 0 {
				end = len(rest)
			}
			segment = strings.TrimSpace(rest[:end])
			if segment == "" || strings.ContainsAny(segment, " \t\"'") {
				return nil, false
			}
			rest = rest[end:]
		}
		path = append(path, segment)
		rest = strings.TrimSpace(rest)
		if rest == "" {
			return path, true
		}
		if rest[0] != '.' {
			return nil, false
		}
		rest = strings.TrimSpace(rest[1:])
	}
}

func tomlClosingQuote(value string) int {
	escaped := false
	for i := 1; i < len(value); i++ {
		switch {
		case escaped:
			escaped = false
		case value[i] == '\\':
			escaped = true
		case value[i] == '"':
			return i
		}
	}
	return -1
}
