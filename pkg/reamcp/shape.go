// Package reamcp owns the project-scoped REA MCP entry Autopus writes into the
// Claude Code, Codex, and Antigravity configuration files: its exact shape,
// the ownership class of whatever `rea` entry a file holds, and the add and
// remove transforms (SPEC-REA-001 MCP Entry Shape Contract). Adapters and the
// CLI share these functions so that one rule decides what Autopus may touch.
package reamcp

import (
	"strings"

	"github.com/insajin/autopus-adk/pkg/config"
)

// Pin is the rea-agents version the current Autopus release emits.
const Pin = "6.5.0"

// ManagedPins lists every pin Autopus ever emitted. An entry with the managed
// shape and one of these pins is Autopus-managed; add moves it to Pin.
var ManagedPins = []string{"6.5.0"}

// Platform names a REA target platform by its Autopus platform name.
type Platform string

const (
	Claude      Platform = "claude-code"     // .mcp.json mcpServers.rea
	Antigravity Platform = "antigravity-cli" // .agents/plugins/autopus/mcp_config.json mcpServers.rea
	Codex       Platform = "codex"           // .codex/config.toml [mcp_servers.rea]
)

// TargetRelPath returns the slash-separated project path of p's REA target
// file, or "" for a platform that has none. Antigravity CLI reads MCP servers
// from the user's mcp_config.json and from each plugin's own mcp_config.json,
// never from a project .gemini/settings.json, so the Antigravity entry lives
// in the workspace plugin, where agy runs it as autopus_rea.
func TargetRelPath(p Platform) string {
	switch p {
	case Claude:
		return ".mcp.json"
	case Antigravity:
		return ".agents/plugins/autopus/mcp_config.json"
	case Codex:
		return ".codex/config.toml"
	default:
		return ""
	}
}

// LegacyTargetRelPath returns the slash-separated project path an earlier
// build wrote p's entry to, or "" when there is none. Release candidates of
// v0.50.126 registered Antigravity REA in .gemini/settings.json, which agy
// never reads; a run removes a managed entry left there and never adds one.
func LegacyTargetRelPath(p Platform) string {
	if p == Antigravity {
		return ".gemini/settings.json"
	}
	return ""
}

// Platforms returns every REA target platform in a stable order.
func Platforms() []Platform {
	return []Platform{Claude, Codex, Antigravity}
}

// Class is the ownership class of the `rea` entry of one target file.
type Class string

const (
	ClassAbsent    Class = "absent"
	ClassManaged   Class = "managed"
	ClassOtherPin  Class = "other_pin"
	ClassUserOwned Class = "user_owned"
)

const (
	serverName = "rea"

	jsonCommand       = "/bin/bash"
	jsonScriptPrefix  = "exec npx -y rea-agents@"
	jsonScriptSuffix  = " mcp"
	jsonCommentPrefix = "REA reverse-engineering MCP (rea-agents@"
	jsonCommentSuffix = ")"
)

// ManagedJSONEntry returns the current-pin entry for a JSON platform. Codex
// has no JSON shape and gets nil.
func ManagedJSONEntry(p Platform) map[string]any {
	if p != Claude && p != Antigravity {
		return nil
	}
	return jsonEntry(p, Pin)
}

func jsonEntry(p Platform, pin string) map[string]any {
	entry := map[string]any{
		"command": jsonCommand,
		"args":    []any{"-l", "-c", jsonScriptPrefix + pin + jsonScriptSuffix},
	}
	if p == Claude {
		entry["$comment"] = jsonCommentPrefix + pin + jsonCommentSuffix
	}
	return entry
}

// ClassifyJSON classifies the `rea` entry of a decoded mcpServers object.
// servers may be nil.
func ClassifyJSON(p Platform, servers map[string]any) Class {
	raw, exists := servers[serverName]
	if !exists {
		return ClassAbsent
	}
	pin, shaped := jsonEntryPin(p, raw)
	if !shaped {
		return ClassUserOwned
	}
	return pinClass(pin)
}

// ApplyJSON applies one action to a decoded mcpServers object. It never edits
// servers: a change comes back in a shallow copy, and an unchanged object is
// returned as is. add inserts the entry when the name is absent and moves a
// managed entry to Pin; remove deletes a managed entry; other_pin and
// user_owned entries are never touched.
func ApplyJSON(p Platform, servers map[string]any, a config.ReaMCPAction) (map[string]any, bool) {
	if p != Claude && p != Antigravity {
		return servers, false
	}
	raw, exists := servers[serverName]
	pin, shaped := jsonEntryPin(p, raw)
	managed := shaped && isManagedPin(pin)
	switch {
	case a == config.ReaMCPAdd && (!exists || managed && pin != Pin):
		out := cloneServers(servers)
		out[serverName] = ManagedJSONEntry(p)
		return out, true
	case a == config.ReaMCPRemove && exists && managed:
		out := cloneServers(servers)
		delete(out, serverName)
		return out, true
	default:
		return servers, false
	}
}

// jsonEntryPin returns the pin of a value that has p's managed shape with only
// the pin slots free. Claude carries the pin twice ($comment and the third
// argument) and both must agree; any extra or missing key fails the shape.
func jsonEntryPin(p Platform, raw any) (string, bool) {
	entry, ok := raw.(map[string]any)
	if !ok || (p != Claude && p != Antigravity) {
		return "", false
	}
	keys := 2
	if p == Claude {
		keys = 3
	}
	if len(entry) != keys || entry["command"] != jsonCommand {
		return "", false
	}
	args, ok := entry["args"].([]any)
	if !ok || len(args) != 3 || args[0] != "-l" || args[1] != "-c" {
		return "", false
	}
	script, _ := args[2].(string)
	pin, ok := pinBetween(script, jsonScriptPrefix, jsonScriptSuffix)
	if !ok {
		return "", false
	}
	if p == Claude {
		comment, _ := entry["$comment"].(string)
		commentPin, ok := pinBetween(comment, jsonCommentPrefix, jsonCommentSuffix)
		if !ok || commentPin != pin {
			return "", false
		}
	}
	return pin, true
}

func cloneServers(servers map[string]any) map[string]any {
	out := make(map[string]any, len(servers)+1)
	for name, value := range servers {
		out[name] = value
	}
	return out
}

// pinBetween extracts the pin slot of value framed by prefix and suffix.
func pinBetween(value, prefix, suffix string) (string, bool) {
	if len(value) < len(prefix)+len(suffix) ||
		!strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, suffix) {
		return "", false
	}
	pin := value[len(prefix) : len(value)-len(suffix)]
	return pin, validPin(pin)
}

// validPin admits an npm version or dist-tag: letters, digits, and `.+-_`.
// Whitespace, quotes, and shell syntax are refused, so a value that only looks
// like a pin slot cannot hide an extra argument and is user-owned instead.
func validPin(pin string) bool {
	if pin == "" {
		return false
	}
	for _, r := range pin {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r == '.', r == '+', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

func isManagedPin(pin string) bool {
	for _, managed := range ManagedPins {
		if pin == managed {
			return true
		}
	}
	return false
}

func pinClass(pin string) Class {
	if isManagedPin(pin) {
		return ClassManaged
	}
	return ClassOtherPin
}
