package taskroute

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
)

// Bounds of the Topology Facts Contract.
const (
	maxTeamUnits      = 32
	maxTeamOwnedPaths = 256
	maxTeamUnitID     = 128
)

// teamFactsKeyPattern is the only key spelling the facts accept. encoding/json
// matches field names case-insensitively, with Unicode folding, and lets a
// later key win, so "OWNED_PATHS" or "owned_pathſ" after "owned_paths" would
// silently replace a declared value that uniqueValue, comparing bytes, saw as
// a distinct key. Every schema key is lowercase ASCII with underscores.
var teamFactsKeyPattern = regexp.MustCompile(`^[a-z_]+$`)

var errTeamFactsKey = errors.New("facts key outside [a-z_]")

// teamShellMetacharacters are refused in owned paths although cleanPath allows
// most of them. The session writes facts through a shell, and one of these in
// a path would run or split a command if the facts were ever written with echo
// or an unquoted heredoc.
const teamShellMetacharacters = "$`" + `\"';&|<>` + "\n"

// TeamFacts is a caller-declared topology facts document: the session copies
// the Phase 1 assignment table into it. Nothing checks the filesystem.
type TeamFacts struct {
	Version int           `json:"version"`
	Flags   TopologyFlags `json:"flags"`
	Units   []TeamUnit    `json:"units"`
}

// TeamFactsError names the Topology Facts Contract code a document violates.
// The message never echoes document content.
type TeamFactsError struct{ Code string }

func (e *TeamFactsError) Error() string { return "invalid topology facts: " + e.Code }

// DecodeTeamFacts decodes and validates a topology facts document with the
// hardening Decode applies: duplicate keys, unknown fields, excessive nesting,
// and trailing values are rejected as code "json".
func DecodeTeamFacts(data []byte) (TeamFacts, error) {
	var facts TeamFacts
	if len(data) > MaxFactsBytes {
		return TeamFacts{}, &TeamFactsError{Code: "json"}
	}
	tokens := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueValue(tokens, 0); err != nil {
		return TeamFacts{}, &TeamFactsError{Code: "json"}
	}
	if _, err := tokens.Token(); err != io.EOF {
		return TeamFacts{}, &TeamFactsError{Code: "json"}
	}
	if err := lowercaseTeamKeys(json.NewDecoder(bytes.NewReader(data))); err != nil {
		return TeamFacts{}, &TeamFactsError{Code: "json"}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&facts); err != nil {
		return TeamFacts{}, &TeamFactsError{Code: "json"}
	}
	if code := teamFactsViolation(facts); code != "" {
		return TeamFacts{}, &TeamFactsError{Code: code}
	}
	return facts, nil
}

// teamFactsViolation returns the first violated code in contract table order,
// except that the unit and path counts are bounded before any per-path work.
func teamFactsViolation(facts TeamFacts) string {
	if facts.Version != 1 {
		return "version"
	}
	flags := 0
	for _, set := range []bool{facts.Flags.Team, facts.Flags.Workflow, facts.Flags.Solo} {
		if set {
			flags++
		}
	}
	if flags > 1 {
		return "topology_conflict"
	}
	if len(facts.Units) > maxTeamUnits {
		return "unit_count"
	}
	ids := map[string]bool{}
	for _, unit := range facts.Units {
		if !validTeamUnitID(unit.ID) || ids[unit.ID] {
			return "unit_id"
		}
		ids[unit.ID] = true
	}
	total := 0
	for _, unit := range facts.Units {
		if unit.Execution != "worker" && unit.Execution != "inline" {
			return "execution"
		}
		total += len(unit.OwnedPaths)
	}
	if total > maxTeamOwnedPaths {
		return "owned_paths"
	}
	for _, unit := range facts.Units {
		if unit.Execution == "worker" && len(unit.OwnedPaths) == 0 {
			return "owned_paths"
		}
		// Only one trailing "/**" is a scope; any glob left after it fails cleanPath.
		for _, root := range ownedRoots(unit.OwnedPaths) {
			if _, err := cleanPath(root); err != nil || strings.ContainsAny(root, teamShellMetacharacters) {
				return "owned_paths"
			}
		}
	}
	for _, unit := range facts.Units {
		for _, id := range unit.DependsOn {
			if id == unit.ID || !ids[id] {
				return "depends_on"
			}
		}
	}
	return ""
}

// lowercaseTeamKeys rejects any object key outside teamFactsKeyPattern. It
// runs after uniqueValue, which already bounds the nesting depth.
func lowercaseTeamKeys(d *json.Decoder) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	for d.More() {
		if delimiter == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			if name, _ := key.(string); !teamFactsKeyPattern.MatchString(name) {
				return errTeamFactsKey
			}
		}
		if err := lowercaseTeamKeys(d); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

// validTeamUnitID accepts the worker ID alphabet validate uses, so an ID is
// safe to print in a reason line.
func validTeamUnitID(id string) bool {
	if id == "" || len(id) > maxTeamUnitID {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}
