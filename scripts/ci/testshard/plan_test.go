package main

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func syntheticNames(count int) []string {
	names := make([]string, count)
	for i := range names {
		names[i] = fmt.Sprintf("TestFamily%03d_Case%04d", i%37, i)
	}
	sort.Strings(names)
	return names
}

func TestParseList_ReturnsSortedRunnableNames(t *testing.T) {
	input := "TestB\nTestA\r\nBenchmarkSpeed\nExampleThing\nFuzzParse\n\nok  \tgithub.com/x/y/internal/cli\t0.785s\n"
	got, err := parseList(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseList: %v", err)
	}
	want := []string{"ExampleThing", "FuzzParse", "TestA", "TestB"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %v, want %v (benchmarks and the summary line are not runnable)", got, want)
	}
}

func TestParseList_RejectsOutputThatCouldHideTests(t *testing.T) {
	for name, input := range map[string]string{
		"build failure":   "# github.com/x/y\n./a.go:1:1: syntax error\nFAIL\tgithub.com/x/y [build failed]\n",
		"fail line":       "TestA\nFAIL\tgithub.com/x/y\t0.1s\n",
		"init noise":      "starting up...\nTestA\n",
		"duplicate":       "TestA\nTestB\nTestA\n",
		"no tests":        "?   \tgithub.com/x/y\t[no test files]\n",
		"only benchmarks": "BenchmarkA\nok  \tgithub.com/x/y\t0.1s\n",
		"empty":           "",
	} {
		if names, err := parseList(strings.NewReader(input)); err == nil {
			t.Errorf("%s: parseList accepted %q and returned %v", name, input, names)
		}
	}
}

func TestNewPlan_CoversEveryTestExactlyOnce(t *testing.T) {
	names := syntheticNames(1500)
	for total := 1; total <= 16; total++ {
		p, err := newPlan(names, total, defaultMaxExprBytes)
		if err != nil {
			t.Fatalf("total %d: %v", total, err)
		}
		if err := p.verify(); err != nil {
			t.Fatalf("total %d: %v", total, err)
		}
		var union []string
		for _, owned := range p.shards {
			union = append(union, owned...)
		}
		sort.Strings(union)
		if !reflect.DeepEqual(union, names) {
			t.Fatalf("total %d: shard union has %d names, want the %d listed", total, len(union), len(names))
		}
	}
}

func TestNewPlan_IsDeterministicAndStableAcrossUniverseChanges(t *testing.T) {
	names := syntheticNames(800)
	first, err := newPlan(names, 5, defaultMaxExprBytes)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := newPlan(append([]string{}, names...), 5, defaultMaxExprBytes)
	if !reflect.DeepEqual(first.exprs, second.exprs) {
		t.Fatal("the same list produced different -run expressions")
	}
	// Adding a test must not move any other test: the only churn is the new one.
	grown, err := newPlan(append(append([]string{}, names...), "TestZzzBrandNew"), 5, defaultMaxExprBytes)
	if err != nil {
		t.Fatal(err)
	}
	for index := range first.shards {
		var kept []string
		for _, name := range grown.shards[index] {
			if name != "TestZzzBrandNew" {
				kept = append(kept, name)
			}
		}
		if !reflect.DeepEqual(kept, first.shards[index]) {
			t.Fatalf("shard %d changed membership when an unrelated test was added", index)
		}
	}
	// Pin a few assignments: the hash is part of the contract that a failing
	// shard names the same tests on a rerun.
	for name, want := range map[string]int{"TestA": 4, "TestWorkflowGate_EmitsExitCodeVerdictJSON": 0} {
		if got := shardOf(name, 5); got != want {
			t.Errorf("shardOf(%q, 5) = %d, want %d; changing the hash reshuffles every shard", name, got, want)
		}
	}
}

func TestNewPlan_RefusesPartitionsThatCannotRunEverything(t *testing.T) {
	names := syntheticNames(40)
	if _, err := newPlan(names, 0, defaultMaxExprBytes); err == nil {
		t.Error("total 0 was accepted")
	}
	if _, err := newPlan(names[:3], 8, defaultMaxExprBytes); err == nil {
		t.Error("more shards than tests was accepted")
	}
	if _, err := newPlan(names, 4, 100); err == nil || !strings.Contains(err.Error(), "add shards") {
		t.Errorf("oversized -run expression error = %v", err)
	}
	// With few names and many shards some shard ends up empty: refuse rather
	// than schedule a job that selects nothing.
	if _, err := newPlan(syntheticNames(12), 12, defaultMaxExprBytes); err == nil ||
		!strings.Contains(err.Error(), "owns no tests") {
		t.Errorf("empty shard error = %v", err)
	}
}

func TestPlanVerify_FailsOnUnassignedDoubleAssignedAndMisplacedTests(t *testing.T) {
	names := syntheticNames(300)
	build := func() *plan {
		p, err := newPlan(names, 4, defaultMaxExprBytes)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	victim := build().shards[0][0]

	dropped := build()
	dropped.shards[0] = dropped.shards[0][1:]
	dropped.exprs[0] = runExpression(dropped.shards[0])
	if err := dropped.verify(); err == nil || !strings.Contains(err.Error(), "unassigned=["+victim) {
		t.Errorf("dropped test not reported as unassigned: %v", err)
	}

	doubled := build()
	doubled.shards[1] = append([]string{victim}, doubled.shards[1]...)
	doubled.exprs[1] = runExpression(doubled.shards[1])
	if err := doubled.verify(); err == nil || !strings.Contains(err.Error(), "double-assigned=["+victim) {
		t.Errorf("duplicated test not reported as double-assigned: %v", err)
	}

	misplaced := build()
	misplaced.shards[0], misplaced.shards[1] = misplaced.shards[1], misplaced.shards[0]
	misplaced.exprs[0], misplaced.exprs[1] = misplaced.exprs[1], misplaced.exprs[0]
	if err := misplaced.verify(); err == nil || !strings.Contains(err.Error(), "misplaced=") {
		t.Errorf("swapped shards not reported as misplaced: %v", err)
	}

	broken := build()
	broken.exprs[2] = "^(unterminated"
	if err := broken.verify(); err == nil || !strings.Contains(err.Error(), "does not compile") {
		t.Errorf("uncompilable expression not reported: %v", err)
	}
}

func TestRunExpression_SelectsExactNamesOnly(t *testing.T) {
	expr := runExpression([]string{"TestA", "TestA_Long", "Test测试"})
	re := regexp.MustCompile(expr)
	for _, name := range []string{"TestA", "TestA_Long", "Test测试"} {
		if !re.MatchString(name) {
			t.Errorf("%q should be selected by %s", name, expr)
		}
	}
	for _, name := range []string{"TestAB", "XTestA", "TestA_", "Test"} {
		if re.MatchString(name) {
			t.Errorf("%q must not be selected by %s", name, expr)
		}
	}
}

func TestPlanEffective_CountsOnlyTestsTheSkipLeavesRunnable(t *testing.T) {
	p, err := newPlan(syntheticNames(200), 3, defaultMaxExprBytes)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.effective(0, nil); got != len(p.shards[0]) {
		t.Fatalf("effective without -skip = %d, want %d", got, len(p.shards[0]))
	}
	everything := regexp.MustCompile(`^Test`)
	if got := p.effective(0, everything); got != 0 {
		t.Fatalf("effective when -skip matches everything = %d", got)
	}
}
