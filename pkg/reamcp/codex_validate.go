package reamcp

import (
	"fmt"
	"strings"
)

// tomlValidateStructure duplicates validateCodexTOMLStructure of the Codex
// adapter (codex_config_structural_validate.go), which this package cannot
// import. ApplyFile runs it on every Codex text a REA action changed, so a
// whole-file write such as the CLI's off sweep refuses exactly the outputs
// the adapter's own write path refuses. pkg/adapter/codex
// TestReaCodexValidatorMatchesMerge keeps the two copies in step.
func tomlValidateStructure(content string) error {
	var state tomlValueState
	continuation := false
	section := ""
	sections := make(map[string]bool)
	assignments := make(map[string]bool)
	for number, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if continuation {
			state.observe(line)
			if state.square < 0 || state.curly < 0 {
				return fmt.Errorf("line %d: unbalanced value", number+1)
			}
			continuation = state.open()
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if parsed, ok := tomlStructuralSection(trimmed); ok {
			if strings.HasPrefix(parsed, "[[") {
				section = fmt.Sprintf("%s#%d", parsed, number)
			} else if sections[parsed] {
				return fmt.Errorf("line %d: duplicate table", number+1)
			} else {
				section, sections[parsed] = parsed, true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			return fmt.Errorf("line %d: invalid table header", number+1)
		}
		equals := strings.IndexByte(trimmed, '=')
		if equals <= 0 || !tomlValidKey(strings.TrimSpace(trimmed[:equals])) {
			return fmt.Errorf("line %d: invalid assignment", number+1)
		}
		value := strings.TrimSpace(trimmed[equals+1:])
		if value == "" {
			return fmt.Errorf("line %d: missing value", number+1)
		}
		assignmentID := section + "\x00" + strings.TrimSpace(trimmed[:equals])
		if assignments[assignmentID] {
			return fmt.Errorf("line %d: duplicate key", number+1)
		}
		assignments[assignmentID] = true
		state = tomlValueState{}
		state.observe(value)
		if state.square < 0 || state.curly < 0 {
			return fmt.Errorf("line %d: unbalanced value", number+1)
		}
		continuation = state.open()
	}
	if continuation {
		return fmt.Errorf("unterminated value")
	}
	return nil
}

func tomlValidKey(key string) bool {
	if key == "" {
		return false
	}
	var quote byte
	for index := range len(key) {
		char := key[index]
		if quote != 0 {
			if char == quote && (quote == '\'' || index == 0 || key[index-1] != '\\') {
				quote = 0
			}
			continue
		}
		switch {
		case char == '"' || char == '\'':
			quote = char
		case char == '.' || char == '-' || char == '_' || char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z':
		default:
			return false
		}
	}
	return quote == 0
}

type tomlValueState struct {
	quote  byte
	triple bool
	escape bool
	square int
	curly  int
}

func (state *tomlValueState) open() bool {
	return state.quote != 0 || state.square > 0 || state.curly > 0
}

func (state *tomlValueState) observe(value string) {
	for index := 0; index < len(value); index++ {
		char := value[index]
		if state.quote != 0 {
			if state.triple && strings.HasPrefix(value[index:], strings.Repeat(string(state.quote), 3)) {
				state.quote, state.triple = 0, false
				index += 2
				continue
			}
			if !state.triple && char == state.quote && (state.quote == '\'' || !state.escape) {
				state.quote = 0
			}
			state.escape = char == '\\' && !state.escape && state.quote == '"'
			if char != '\\' {
				state.escape = false
			}
			continue
		}
		if char == '#' {
			return
		}
		if char == '"' || char == '\'' {
			state.quote = char
			state.triple = strings.HasPrefix(value[index:], strings.Repeat(string(char), 3))
			if state.triple {
				index += 2
			}
			continue
		}
		switch char {
		case '[':
			state.square++
		case ']':
			state.square--
		case '{':
			state.curly++
		case '}':
			state.curly--
		}
	}
}
