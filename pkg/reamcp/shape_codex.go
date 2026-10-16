package reamcp

import (
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

const (
	codexHeader      = "[mcp_servers.rea]"
	codexCommandLine = `command = "npx"`
	codexArgsPrefix  = `args = ["-y", "rea-agents@`
	codexArgsSuffix  = `", "mcp"]`
	codexTimeoutLine = "startup_timeout_sec = 30"
)

func codexTableLines(pin string) []string {
	return []string{codexHeader, codexCommandLine, codexArgsPrefix + pin + codexArgsSuffix, codexTimeoutLine}
}

// codexTable is one `[mcp_servers.rea]` or `[mcp_servers.rea.*]` header and
// the lines it owns. A trailing run of blank and comment lines is left to
// whatever follows, so a comment written above the next table never makes
// the rea table user-owned.
type codexTable struct {
	header int  // index of the header line
	end    int  // index one past the table's last own line
	nested bool // a [mcp_servers.rea.*] subtable or an array-of-tables header
}

type codexScan struct {
	tables []codexTable
	inline bool // rea named by a dotted or inline key outside its own table
}

// ClassifyCodexTOML classifies the rea server of a Codex config text. The
// rea server is every `[mcp_servers.rea]` and `[mcp_servers.rea.*]` table plus
// any dotted or inline key naming it; only one exact four-line table is
// managed.
func ClassifyCodexTOML(text string) Class {
	lines := strings.Split(text, "\n")
	class, _ := classifyCodex(lines, scanCodex(lines))
	return class
}

// ApplyCodexTOML applies one action to a Codex config text. add appends one
// blank line and the managed table at the end when no rea server exists and
// moves a managed table to Pin in place; remove deletes a managed table and
// the blank line before it, so add then remove restores the input bytes.
// Every line outside a managed table keeps its bytes.
func ApplyCodexTOML(text string, a config.ReaMCPAction) (string, bool) {
	if a != config.ReaMCPAdd && a != config.ReaMCPRemove {
		return text, false
	}
	lines := strings.Split(text, "\n")
	scan := scanCodex(lines)
	class, pin := classifyCodex(lines, scan)
	switch {
	case a == config.ReaMCPAdd && class == ClassAbsent:
		return appendCodexTable(text), true
	case a == config.ReaMCPAdd && class == ClassManaged && pin != Pin:
		return repinCodexTable(lines, scan.tables[0]), true
	case a == config.ReaMCPRemove && class == ClassManaged:
		return removeCodexTable(lines, scan.tables[0]), true
	default:
		return text, false
	}
}

func scanCodex(lines []string) codexScan {
	var (
		scan    tomlScanState
		result  codexScan
		section []string
		open    = -1 // index in result.tables of the rea table being read
	)
	for i, line := range lines {
		if scan.skipSyntaxLine(line) {
			if open >= 0 {
				result.tables[open].end = i + 1
			}
			continue
		}
		trimmed := strings.TrimSpace(line)
		if name, ok := tomlStructuralSection(trimmed); ok {
			path, array := codexSectionPath(name)
			section, open = path, -1
			if isReaPath(path) {
				result.tables = append(result.tables, codexTable{header: i, end: i + 1, nested: array || len(path) > 2})
				open = len(result.tables) - 1
			}
			continue
		}
		scan.observeLine(trimmed)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if open >= 0 {
			result.tables[open].end = i + 1
		} else if namesReaOutsideTable(section, trimmed) {
			result.inline = true
		}
	}
	return result
}

func classifyCodex(lines []string, scan codexScan) (Class, string) {
	switch {
	case len(scan.tables) == 0 && !scan.inline:
		return ClassAbsent, ""
	case scan.inline || len(scan.tables) != 1 || scan.tables[0].nested:
		return ClassUserOwned, ""
	}
	table := scan.tables[0]
	if lines[table.header] != codexHeader {
		return ClassUserOwned, ""
	}
	var body []string
	for _, line := range lines[table.header+1 : table.end] {
		if strings.TrimSpace(line) != "" {
			body = append(body, line)
		}
	}
	if len(body) != 3 || body[0] != codexCommandLine || body[2] != codexTimeoutLine {
		return ClassUserOwned, ""
	}
	pin, ok := pinBetween(body[1], codexArgsPrefix, codexArgsSuffix)
	if !ok {
		return ClassUserOwned, ""
	}
	return pinClass(pin), pin
}

func appendCodexTable(text string) string {
	block := strings.Join(codexTableLines(Pin), "\n") + "\n"
	if text == "" {
		return block
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return text + "\n" + block
}

func repinCodexTable(lines []string, table codexTable) string {
	out := append([]string(nil), lines...)
	for i := table.header + 1; i < table.end; i++ {
		if strings.HasPrefix(out[i], codexArgsPrefix) {
			out[i] = codexTableLines(Pin)[2]
		}
	}
	return strings.Join(out, "\n")
}

func removeCodexTable(lines []string, table codexTable) string {
	start, end := table.header, table.end
	switch {
	case start > 0 && lines[start-1] == "":
		start--
	case start == 0 && end < len(lines)-1 && lines[end] == "":
		end++
	}
	out := append(append([]string(nil), lines[:start]...), lines[end:]...)
	return strings.Join(out, "\n")
}

// codexSectionPath turns a header name from tomlStructuralSection into key
// segments and reports whether it was an array-of-tables header.
func codexSectionPath(name string) ([]string, bool) {
	array := strings.HasPrefix(name, "[[") && strings.HasSuffix(name, "]]")
	if array {
		name = name[2 : len(name)-2]
	}
	path, ok := tomlKeyPath(name)
	if !ok {
		return nil, array
	}
	return path, array
}

// namesReaOutsideTable reports whether an assignment line defines the rea
// server through a dotted or inline key, such as `rea = { ... }` under
// [mcp_servers] or a root `mcp_servers.rea.command = ...`.
func namesReaOutsideTable(section []string, trimmed string) bool {
	equals := strings.IndexByte(trimmed, '=')
	if equals <= 0 {
		return false
	}
	key, ok := tomlKeyPath(trimmed[:equals])
	if !ok {
		return false
	}
	full := append(append([]string(nil), section...), key...)
	return isReaPath(full)
}

func isReaPath(path []string) bool {
	return len(path) >= 2 && path[0] == "mcp_servers" && path[1] == serverName
}
