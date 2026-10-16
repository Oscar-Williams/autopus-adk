package reamcp

import (
	"strconv"
	"strings"
)

// The scanner below duplicates the line rules of the Codex adapter's
// structural config merge (codexTOMLScanState, parseStructuralCodexSection,
// and parseCodexConfigAssignment in pkg/adapter/codex), which this package
// cannot import. Sharing the rules means the rea table edited here is the
// table that merge sees, and a `[` inside a multiline string is never a
// header. pkg/adapter/codex TestReaCodexHeaderDetectionMatchesMerge holds the
// two copies to the same header decisions.

type tomlScanState struct {
	multiline string
}

func (s *tomlScanState) skipSyntaxLine(line string) bool {
	if s.multiline == "" {
		return false
	}
	if tomlHasUnescapedToken(line, s.multiline) {
		s.multiline = ""
	}
	return true
}

func (s *tomlScanState) observeValue(value string) {
	value = tomlValueWithoutComment(value)
	for _, delimiter := range []string{`"""`, `'''`} {
		if tomlHasOddUnescapedToken(value, delimiter) {
			s.multiline = delimiter
			return
		}
	}
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

func tomlHasUnescapedToken(value, token string) bool {
	for offset := 0; offset < len(value); {
		index := strings.Index(value[offset:], token)
		if index < 0 {
			return false
		}
		index += offset
		if !tomlIsEscapedAt(value, index) {
			return true
		}
		offset = index + len(token)
	}
	return false
}

func tomlIsEscapedAt(value string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && value[i] == '\\'; i-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func tomlHasOddUnescapedToken(value, token string) bool {
	count := 0
	for offset := 0; offset < len(value); {
		index := strings.Index(value[offset:], token)
		if index < 0 {
			break
		}
		index += offset
		if !tomlIsEscapedAt(value, index) {
			count++
		}
		offset = index + len(token)
	}
	return count%2 == 1
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

// tomlAssignmentValue mirrors parseCodexConfigAssignment. It admits only a
// single-segment key, and only such a line moves the multiline state, exactly
// as the merge does.
func tomlAssignmentValue(trimmed string) (string, bool) {
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	equals := strings.IndexByte(trimmed, '=')
	if equals < 0 || !tomlSingleKey(strings.TrimSpace(trimmed[:equals])) {
		return "", false
	}
	value := strings.TrimSpace(trimmed[equals+1:])
	return value, value != ""
}

func tomlSingleKey(raw string) bool {
	if raw == "" {
		return false
	}
	if raw[0] == '"' {
		value, err := strconv.Unquote(raw)
		return err == nil && value != "" && !strings.Contains(value, ".")
	}
	if raw[0] == '\'' {
		if len(raw) < 2 || raw[len(raw)-1] != '\'' {
			return false
		}
		value := raw[1 : len(raw)-1]
		return value != "" && !strings.Contains(value, ".")
	}
	return !strings.ContainsAny(raw, " \t.\"'")
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
