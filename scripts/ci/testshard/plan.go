package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"regexp"
	"sort"
	"strings"
)

// runnableName matches the names `go test -list` prints for targets that
// `go test -run` selects. Benchmarks are listed too, but they only run under
// -bench, so they are not part of a -run partition.
var (
	runnableName  = regexp.MustCompile(`^(Test|Example|Fuzz)[\p{L}\p{N}_]*$`)
	benchmarkName = regexp.MustCompile(`^Benchmark[\p{L}\p{N}_]*$`)
	summaryLine   = regexp.MustCompile(`^ok\s`)
)

// parseList reads `go test -list .` output and returns the sorted runnable
// names. Anything it does not recognise (a build failure, a FAIL line, output
// printed by an init function) is an error: a partition computed from a
// polluted list could silently drop tests.
func parseList(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	seen := make(map[string]struct{})
	var names []string
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimRight(scanner.Text(), "\r")
		switch {
		case text == "", benchmarkName.MatchString(text), summaryLine.MatchString(text):
			continue
		case runnableName.MatchString(text):
			if _, dup := seen[text]; dup {
				return nil, fmt.Errorf("line %d: duplicate test name %q", number, text)
			}
			seen[text] = struct{}{}
			names = append(names, text)
		default:
			return nil, fmt.Errorf("line %d: unexpected go test -list output %q", number, text)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read test list: %w", err)
	}
	if len(names) == 0 {
		return nil, errors.New("go test -list printed no runnable tests")
	}
	sort.Strings(names)
	return names, nil
}

// shardOf maps a test name to its shard. It is a pure function of the name and
// the shard count, so a test keeps its shard when unrelated tests are added or
// removed, and no committed assignment list can go stale.
func shardOf(name string, total int) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(name))
	return int(hash.Sum32() % uint32(total))
}

// plan is a complete partition of one package's runnable tests.
type plan struct {
	names  []string   // every runnable test, sorted
	shards [][]string // shards[i] holds the sorted names shard i owns
	exprs  []string   // exprs[i] is the -run expression for shard i
}

// newPlan partitions names into total shards and rejects any partition that is
// empty somewhere or whose -run expression would be too long for one argument.
func newPlan(names []string, total, maxExprBytes int) (*plan, error) {
	if total < 1 {
		return nil, fmt.Errorf("shard total must be at least 1, got %d", total)
	}
	if len(names) < total {
		return nil, fmt.Errorf("%d tests cannot fill %d shards", len(names), total)
	}
	p := &plan{names: names, shards: make([][]string, total), exprs: make([]string, total)}
	for _, name := range names {
		index := shardOf(name, total)
		p.shards[index] = append(p.shards[index], name)
	}
	for index, owned := range p.shards {
		if len(owned) == 0 {
			return nil, fmt.Errorf("shard %d of %d owns no tests", index, total)
		}
		p.exprs[index] = runExpression(owned)
		if len(p.exprs[index]) > maxExprBytes {
			return nil, fmt.Errorf(
				"shard %d of %d needs a %d-byte -run expression, above the %d-byte limit; add shards",
				index, total, len(p.exprs[index]), maxExprBytes)
		}
	}
	return p, nil
}

// runExpression builds an anchored alternation of exact test names.
func runExpression(names []string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = regexp.QuoteMeta(name)
	}
	return "^(" + strings.Join(quoted, "|") + ")$"
}

// verify re-derives ownership from the expressions themselves, with the same
// regexp engine `go test -run` uses, and requires every test to be selected by
// exactly one shard, the one shardOf names. An unassigned or double-assigned
// test fails here rather than going missing from a green run.
func (p *plan) verify() error {
	compiled := make([]*regexp.Regexp, len(p.exprs))
	for i, expr := range p.exprs {
		re, err := regexp.Compile(expr)
		if err != nil {
			return fmt.Errorf("shard %d -run expression does not compile: %w", i, err)
		}
		compiled[i] = re
	}
	total := len(p.exprs)
	var unassigned, doubled, misplaced []string
	for _, name := range p.names {
		var owners []int
		for i, re := range compiled {
			if re.MatchString(name) {
				owners = append(owners, i)
			}
		}
		switch {
		case len(owners) == 0:
			unassigned = append(unassigned, name)
		case len(owners) > 1:
			doubled = append(doubled, name)
		case owners[0] != shardOf(name, total):
			misplaced = append(misplaced, name)
		}
	}
	assigned := 0
	for _, owned := range p.shards {
		assigned += len(owned)
	}
	if len(unassigned)+len(doubled)+len(misplaced) > 0 || assigned != len(p.names) {
		return fmt.Errorf("shard partition is not exact: unassigned=%v double-assigned=%v misplaced=%v owned=%d of %d",
			head(unassigned), head(doubled), head(misplaced), assigned, len(p.names))
	}
	return nil
}

// effective counts the names in shard index that a -skip expression leaves
// runnable. A shard whose every test is skipped would pass having run nothing.
func (p *plan) effective(index int, skip *regexp.Regexp) int {
	if skip == nil {
		return len(p.shards[index])
	}
	count := 0
	for _, name := range p.shards[index] {
		if !skip.MatchString(name) {
			count++
		}
	}
	return count
}

// universeDigest identifies the exact set of tests a plan was cut from, so a
// later step can prove every shard saw the same universe.
func (p *plan) universeDigest() string {
	sum := sha256.Sum256([]byte(strings.Join(p.names, "\n")))
	return hex.EncodeToString(sum[:])
}

// head trims a diagnostic list so a broken partition does not print thousands
// of names.
func head(names []string) []string {
	if len(names) > 8 {
		return append(append([]string{}, names[:8]...), fmt.Sprintf("... %d more", len(names)-8))
	}
	return names
}
