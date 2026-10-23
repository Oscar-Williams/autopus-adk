package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func listOutput(names []string) string {
	return strings.Join(names, "\n") + "\nok  \tgithub.com/insajin/autopus-adk/internal/cli\t0.785s\n"
}

func runTool(t *testing.T, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), code
}

func TestRunPlan_PrintsTheShardExpressionAndLeavesEvidence(t *testing.T) {
	names := syntheticNames(500)
	names = append(names, "TestDetectHeavy", "TestValidateReadinessHeavy")
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.txt")
	namesPath := filepath.Join(dir, "names.txt")

	stdout, stderr, code := runTool(t, listOutput(names), "plan",
		"-total", "4", "-index", "2", "-lane", "unit", "-skip", "^(TestDetect|TestValidateReadiness)",
		"-manifest", manifestPath, "-names", namesPath)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	expr := strings.TrimSpace(stdout)
	re := regexp.MustCompile(expr)

	written, err := os.ReadFile(namesPath)
	if err != nil {
		t.Fatal(err)
	}
	owned := strings.Fields(string(written))
	if len(owned) == 0 {
		t.Fatal("names file is empty")
	}
	for _, name := range owned {
		if !re.MatchString(name) {
			t.Errorf("the printed expression does not select %s from the names file", name)
		}
	}
	selected := 0
	for _, name := range names {
		if re.MatchString(name) {
			selected++
		}
	}
	if selected != len(owned) {
		t.Errorf("expression selects %d names, names file lists %d", selected, len(owned))
	}

	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	got, err := parseManifest(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}
	if got.Lane != "unit" || got.Shard != 2 || got.Total != 4 || got.Tests != len(names) || got.Assigned != len(owned) {
		t.Errorf("manifest = %+v, want lane unit shard 2/4 over %d tests owning %d", got, len(names), len(owned))
	}
	if !strings.Contains(stderr, "shard 3/4") {
		t.Errorf("summary should name the human shard number: %q", stderr)
	}
}

func TestRunPlan_EveryShardTogetherSelectsEachListedTestOnce(t *testing.T) {
	names := syntheticNames(700)
	counts := make(map[string]int)
	var manifests []manifest
	for index := 0; index < 6; index++ {
		path := filepath.Join(t.TempDir(), "manifest.txt")
		stdout, stderr, code := runTool(t, listOutput(names), "plan",
			"-total", "6", "-index", fmt.Sprint(index), "-manifest", path)
		if code != 0 {
			t.Fatalf("shard %d: exit %d: %s", index, code, stderr)
		}
		re := regexp.MustCompile(strings.TrimSpace(stdout))
		for _, name := range names {
			if re.MatchString(name) {
				counts[name]++
			}
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		m, err := parseManifest(file)
		_ = file.Close()
		if err != nil {
			t.Fatal(err)
		}
		manifests = append(manifests, m)
	}
	var wrong []string
	for _, name := range names {
		if counts[name] != 1 {
			wrong = append(wrong, fmt.Sprintf("%s x%d", name, counts[name]))
		}
	}
	sort.Strings(wrong)
	if len(wrong) > 0 {
		t.Fatalf("tests not selected exactly once: %v", wrong)
	}
	if err := verifyManifests(manifests, []string{"default"}); err != nil {
		t.Fatalf("manifests of a real plan failed verification: %v", err)
	}
}

func TestRunPlan_RefusesBadInvocations(t *testing.T) {
	list := listOutput(syntheticNames(60))
	for name, args := range map[string][]string{
		"index out of range": {"plan", "-total", "3", "-index", "3"},
		"missing index":      {"plan", "-total", "3"},
		"zero total":         {"plan", "-total", "0", "-index", "0"},
		"bad skip":           {"plan", "-total", "3", "-index", "0", "-skip", "(["},
		"skip empties shard": {"plan", "-total", "3", "-index", "0", "-skip", "^Test"},
		"too long":           {"plan", "-total", "3", "-index", "0", "-max-expr-bytes", "10"},
		"unwritable output":  {"plan", "-total", "3", "-index", "0", "-manifest", filepath.Join(t.TempDir(), "no", "such", "m.txt")},
		"unknown flag":       {"plan", "-nope"},
	} {
		if stdout, _, code := runTool(t, list, args...); code == 0 {
			t.Errorf("%s: exit 0 with stdout %q", name, stdout)
		}
	}
	if _, _, code := runTool(t, "FAIL\tgithub.com/x/y [build failed]\n", "plan", "-total", "2", "-index", "0"); code == 0 {
		t.Error("a build failure in the list must stop the plan")
	}
	if _, _, code := runTool(t, ""); code != 2 {
		t.Errorf("no command exit = %d, want 2", code)
	}
	if _, _, code := runTool(t, "", "shuffle"); code != 2 {
		t.Errorf("unknown command exit = %d, want 2", code)
	}
}

func TestRunVerify_ReadsManifestFiles(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for _, m := range laneManifests("unit", 3, 30) {
		path := filepath.Join(dir, fmt.Sprintf("m%d.txt", m.Shard))
		if err := os.WriteFile(path, []byte(m.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	stdout, stderr, code := runTool(t, "", append([]string{"verify", "-lanes", "unit"}, paths...)...)
	if code != 0 || !strings.Contains(stdout, "3 shard manifests") {
		t.Fatalf("verify exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	if _, _, code := runTool(t, "", append([]string{"verify", "-lanes", "unit,integration"}, paths...)...); code == 0 {
		t.Error("a required lane with no manifests must fail verification")
	}
	if _, _, code := runTool(t, "", "verify", "-lanes", "unit", paths[0], paths[1]); code == 0 {
		t.Error("a missing shard must fail verification")
	}
	if _, _, code := runTool(t, "", "verify"); code == 0 {
		t.Error("verify without manifests must fail")
	}
	if _, _, code := runTool(t, "", "verify", filepath.Join(dir, "absent.txt")); code == 0 {
		t.Error("an unreadable manifest must fail verification")
	}
	junk := filepath.Join(dir, "junk.txt")
	if err := os.WriteFile(junk, []byte("not a manifest\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, code := runTool(t, "", "verify", junk); code == 0 {
		t.Error("a malformed manifest must fail verification")
	}
}
