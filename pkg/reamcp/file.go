package reamcp

import (
	"encoding/json"
	"fmt"

	"github.com/insajin/autopus-adk/pkg/config"
)

// ClassifyFile classifies the rea entry of one whole target file. data may be
// nil or empty for a missing file, which is absent. A JSON file that does not
// decode, or whose mcpServers is not an object, is an error.
func ClassifyFile(p Platform, data []byte) (Class, error) {
	switch p {
	case Codex:
		return ClassifyCodexTOML(string(data)), nil
	case Claude, Antigravity:
		_, servers, err := decodeJSONTarget(p, data)
		if err != nil {
			return "", err
		}
		return ClassifyJSON(p, servers), nil
	default:
		return "", fmt.Errorf("unknown REA MCP platform %q", p)
	}
}

// ApplyFile applies one action to a whole target file; data may be nil or
// empty for a missing file. An unchanged file comes back as data itself. A
// changed JSON file is re-encoded the way the Claude and Antigravity adapter
// merges encode it (json.MarshalIndent with two spaces and a trailing
// newline), so the result equals what that merge would write. A changed Codex
// text must pass the structural check the Codex adapter applies to its own
// REA writes; otherwise the change is refused with an error naming the file.
func ApplyFile(p Platform, data []byte, a config.ReaMCPAction) ([]byte, bool, error) {
	switch p {
	case Codex:
		out, changed := ApplyCodexTOML(string(data), a)
		if !changed {
			return data, false, nil
		}
		if err := tomlValidateStructure(out); err != nil {
			return nil, false, fmt.Errorf("%s: REA MCP change fails the TOML structure check: %w", TargetRelPath(p), err)
		}
		return []byte(out), true, nil
	case Claude, Antigravity:
		if a != config.ReaMCPAdd && a != config.ReaMCPRemove {
			return data, false, nil
		}
		doc, servers, err := decodeJSONTarget(p, data)
		if err != nil {
			return nil, false, err
		}
		out, changed := ApplyJSON(p, servers, a)
		if !changed {
			return data, false, nil
		}
		doc["mcpServers"] = out
		encoded, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", TargetRelPath(p), err)
		}
		return append(encoded, '\n'), true, nil
	default:
		return nil, false, fmt.Errorf("unknown REA MCP platform %q", p)
	}
}

// decodeJSONTarget decodes a JSON target file and returns the document and its
// mcpServers object (nil when the key is absent).
func decodeJSONTarget(p Platform, data []byte) (map[string]any, map[string]any, error) {
	doc := map[string]any{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", TargetRelPath(p), err)
		}
		if doc == nil {
			doc = map[string]any{}
		}
	}
	raw, exists := doc["mcpServers"]
	if !exists {
		return doc, nil, nil
	}
	servers, ok := raw.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("%s: mcpServers is not an object", TargetRelPath(p))
	}
	return doc, servers, nil
}
