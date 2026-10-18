package taskroute

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// teamFactsJSON wraps unit objects in a version 1 document with no flags.
func teamFactsJSON(units ...string) string {
	return `{"version": 1, "flags": {"team": false, "workflow": false, "solo": false}, "units": [` + strings.Join(units, ",") + `]}`
}

func workerJSON(id string, paths ...string) string {
	return fmt.Sprintf(`{"id": %q, "execution": "worker", "owned_paths": ["%s"], "depends_on": [], "done": false}`, id, strings.Join(paths, `", "`))
}

func TestDecodeTeamFactsAcceptsContractDocument(t *testing.T) {
	raw := `{"version": 1,
 "flags": {"team": false, "workflow": false, "solo": false},
 "units": [{"id": "T1", "execution": "worker", "owned_paths": ["pkg/a/**"], "depends_on": [], "done": false},
           {"id": "T2", "execution": "inline", "owned_paths": [], "depends_on": ["T1"], "done": true}]}`
	facts, err := DecodeTeamFacts([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	want := TeamFacts{Version: 1, Units: []TeamUnit{
		{ID: "T1", Execution: "worker", OwnedPaths: []string{"pkg/a/**"}, DependsOn: []string{}},
		{ID: "T2", Execution: "inline", OwnedPaths: []string{}, DependsOn: []string{"T1"}, Done: true},
	}}
	if !reflect.DeepEqual(facts, want) {
		t.Fatalf("%+v", facts)
	}
	// Optional members default: no flags, no units, a missing dependency list.
	for _, raw := range []string{`{"version": 1}`, `{"version": 1, "units": []}`, `{"version": 1, "units": [{"id": "A", "execution": "worker", "owned_paths": ["a"]}]}`} {
		if _, err := DecodeTeamFacts([]byte(raw)); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
	}
}

func TestDecodeTeamFactsErrorCodes(t *testing.T) {
	units := func(n int) string {
		items := make([]string, n)
		for i := range items {
			items[i] = workerJSON(fmt.Sprintf("U%d", i), fmt.Sprintf("pkg/u%d", i))
		}
		return teamFactsJSON(items...)
	}
	manyPaths := func(n int) string {
		paths := make([]string, n)
		for i := range paths {
			paths[i] = fmt.Sprintf("pkg/p%d", i)
		}
		return teamFactsJSON(workerJSON("T1", paths...))
	}
	cases := []struct {
		raw  string
		code string
	}{
		{`{"version": 2, "units": []}`, "version"},
		{`{"units": []}`, "version"},
		{`null`, "version"},
		{teamFactsJSON(`{"id": "T1", "execution": "worker", "owned_paths": ["a"], "depends_on": ["T9"]}`), "depends_on"},
		{teamFactsJSON(`{"id": "T1", "execution": "worker", "owned_paths": ["a"], "depends_on": ["T1"]}`), "depends_on"},
		{teamFactsJSON(`{"id": "T1", "id": "T2", "execution": "worker", "owned_paths": ["a"]}`), "json"},
		{`{"version": 1, "units": [], "extra": 1}`, "json"},
		{teamFactsJSON(`{"id": "T1", "execution": "worker", "owned_paths": ["a"], "private": "secret"}`), "json"},
		{`{`, "json"},
		{`{"version": 1} {}`, "json"},
		{`[]`, "json"},
		{`{"version": 1, "units": [{"id": "T1", "done": "yes"}]}`, "json"},
		{strings.Repeat("[", 20) + strings.Repeat("]", 20), "json"},
		{teamFactsJSON(`{"id": "T1", "execution": "worker", "owned_paths": []}`), "owned_paths"},
		{teamFactsJSON(`{"id": "T1", "execution": "worker"}`), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "../x")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "/abs")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "pkg/*/x")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "pkg/a/**/**")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "./**")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "**")), "owned_paths"},
		{teamFactsJSON(workerJSON("T1", "/**")), "owned_paths"},
		{teamFactsJSON(`{"id": "T1", "execution": "inline", "owned_paths": ["C:/x"]}`), "owned_paths"},
		{manyPaths(257), "owned_paths"},
		{teamFactsJSON(`{"id": "T1", "execution": "parallel", "owned_paths": ["a"]}`), "execution"},
		{teamFactsJSON(`{"id": "T1", "owned_paths": ["a"]}`), "execution"},
		{units(33), "unit_count"},
		{teamFactsJSON(workerJSON("T 1", "a")), "unit_id"},
		{teamFactsJSON(workerJSON("", "a")), "unit_id"},
		{teamFactsJSON(workerJSON(strings.Repeat("x", 129), "a")), "unit_id"},
		{teamFactsJSON(workerJSON("T1", "a"), workerJSON("T1", "b")), "unit_id"},
		{teamFactsJSON(workerJSON("T/1", "a")), "unit_id"},
		{`{"version": 1, "flags": {"solo": true, "team": true}}`, "topology_conflict"},
		{`{"version": 1, "flags": {"solo": true, "team": true, "workflow": true}}`, "topology_conflict"},
	}
	for _, tc := range cases {
		_, err := DecodeTeamFacts([]byte(tc.raw))
		var factsErr *TeamFactsError
		if !errors.As(err, &factsErr) || factsErr.Code != tc.code || err.Error() != "invalid topology facts: "+tc.code {
			t.Fatalf("%.80s: want %s, got %v", tc.raw, tc.code, err)
		}
		if strings.Contains(err.Error(), "secret") {
			t.Fatal("error echoes input")
		}
	}
	if _, err := DecodeTeamFacts(make([]byte, MaxFactsBytes+1)); err == nil || err.Error() != "invalid topology facts: json" {
		t.Fatalf("oversize document: %v", err)
	}
}

func TestDecodeTeamFactsBoundsAreInclusive(t *testing.T) {
	items := make([]string, 32)
	for i := range items {
		items[i] = workerJSON(fmt.Sprintf("U%d", i), fmt.Sprintf("pkg/u%d", i))
	}
	paths := make([]string, 256)
	for i := range paths {
		paths[i] = fmt.Sprintf("pkg/p%d/**", i)
	}
	for _, raw := range []string{
		teamFactsJSON(items...),
		teamFactsJSON(workerJSON("T1", paths...)),
		teamFactsJSON(workerJSON(strings.Repeat("A", 128), "a"), workerJSON("a-b_c.9", "b")),
		`{"version": 1, "flags": {"solo": true}}`,
	} {
		if _, err := DecodeTeamFacts([]byte(raw)); err != nil {
			t.Fatalf("%.80s: %v", raw, err)
		}
	}
}

// encoding/json matches field names case-insensitively, with Unicode folding,
// and lets a later key win, so these keys would silently replace or set a
// declared value that uniqueValue, which compares bytes, sees as distinct.
func TestDecodeTeamFactsRejectsKeysTheDecoderWouldFold(t *testing.T) {
	for name, raw := range map[string]string{
		"later upper-case owned_paths": `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/a"],"OWNED_PATHS":["pkg/z"]}]}`,
		"long s":                       `{"version":1,"units":[{"id":"T1","execution":"worker","owned_path` + "ſ" + `":["pkg/a"]}]}`,
		"Done":                         `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["a"],"Done":true}]}`,
		"TEAM":                         `{"version":1,"flags":{"TEAM":true},"units":[]}`,
		"Kelvin sign":                  `{"version":1,"flags":{"wor` + "K" + `flow":true},"units":[]}`,
		"Version":                      `{"Version":1}`,
		"Units":                        `{"version":1,"Units":[]}`,
		"ID":                           `{"version":1,"units":[{"ID":"T1","execution":"worker","owned_paths":["a"]}]}`,
		"empty key":                    `{"version":1,"":1}`,
	} {
		if _, err := DecodeTeamFacts([]byte(raw)); err == nil || err.Error() != "invalid topology facts: json" {
			t.Fatalf("%s: %v", name, err)
		}
	}
	// The overlap a folded key would hide: T1 reads as pkg/a to a reviewer.
	hidden := `{"version":1,"units":[{"id":"T1","execution":"worker","owned_paths":["pkg/a"],"OWNED_PATHS":["pkg/z"]},{"id":"T2","execution":"worker","owned_paths":["pkg/a"]}]}`
	if _, err := DecodeTeamFacts([]byte(hidden)); err == nil {
		t.Fatal("a folded key replaced owned_paths")
	}
	// An escaped lowercase key is still exactly the field name.
	if _, err := DecodeTeamFacts([]byte(`{"version":1}`)); err != nil {
		t.Fatal(err)
	}
}

// The session writes facts through a shell, so an owned path that would run
// or split a command if written with echo or an unquoted heredoc is refused.
func TestDecodeTeamFactsRejectsShellMetacharactersInOwnedPaths(t *testing.T) {
	facts := func(execution, owned string) []byte {
		raw, err := json.Marshal(map[string]any{"version": 1, "units": []any{
			map[string]any{"id": "T1", "execution": execution, "owned_paths": []string{owned}},
		}})
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	quote, backslash := string(rune(34)), string(rune(92))
	for _, owned := range []string{"pkg/$(id)/x", "pkg/`id`/x", "pkg/$HOME", "pkg/a" + backslash + "b", "pkg/a" + quote + "b", "pkg/a'b", "pkg/a;b", "pkg/a&b", "pkg/a|b", "pkg/a<b", "pkg/a>b", "pkg/a\nb", "pkg/$x/**"} {
		for _, execution := range []string{"worker", "inline"} {
			if _, err := DecodeTeamFacts(facts(execution, owned)); err == nil || err.Error() != "invalid topology facts: owned_paths" {
				t.Fatalf("%s %q: %v", execution, owned, err)
			}
		}
	}
	for _, owned := range []string{"pkg/a b", "pkg/a-b_c.d", "pkg/(x)", "pkg/a+b", "pkg/a#b", "pkg/@x", "pkg/a,b", "pkg/~x", "pkg/a=b/**"} {
		if _, err := DecodeTeamFacts(facts("worker", owned)); err != nil {
			t.Fatalf("%q: %v", owned, err)
		}
	}
}
