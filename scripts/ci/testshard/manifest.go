package main

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// manifest is what one shard leaves behind so the aggregating job can prove,
// without rerunning anything, that the shards together cover the package.
type manifest struct {
	Lane      string
	Shard     int
	Total     int
	Tests     int // size of the whole universe the plan was cut from
	Assigned  int // tests this shard owns
	Effective int // owned tests that -skip leaves runnable
	Universe  string
}

func (m manifest) String() string {
	return fmt.Sprintf("lane=%s\nshard=%d\ntotal=%d\ntests=%d\nassigned=%d\neffective=%d\nuniverse=%s\n",
		m.Lane, m.Shard, m.Total, m.Tests, m.Assigned, m.Effective, m.Universe)
}

// parseManifest reads the key=value form String writes and rejects unknown,
// repeated, or missing keys: a truncated manifest must not count as evidence.
func parseManifest(r io.Reader) (manifest, error) {
	var m manifest
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || seen[key] {
			return manifest{}, fmt.Errorf("malformed or repeated manifest line %q", line)
		}
		seen[key] = true
		if err := m.set(key, value); err != nil {
			return manifest{}, err
		}
	}
	if err := scanner.Err(); err != nil {
		return manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	for _, key := range []string{"lane", "shard", "total", "tests", "assigned", "effective", "universe"} {
		if !seen[key] {
			return manifest{}, fmt.Errorf("manifest is missing %q", key)
		}
	}
	return m, nil
}

func (m *manifest) set(key, value string) error {
	numbers := map[string]*int{
		"shard": &m.Shard, "total": &m.Total, "tests": &m.Tests,
		"assigned": &m.Assigned, "effective": &m.Effective,
	}
	if target, ok := numbers[key]; ok {
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("manifest %s=%q is not a non-negative integer", key, value)
		}
		*target = n
		return nil
	}
	switch key {
	case "lane":
		m.Lane = value
	case "universe":
		m.Universe = value
	default:
		return fmt.Errorf("unknown manifest key %q", key)
	}
	return nil
}

// verifyManifests checks, per lane, that the shards form one exact cover: every
// index 0..total-1 appears once, all shards were cut from the same universe,
// the owned counts add up to it, and none was left with nothing to run. Every
// required lane must be present, so a lane that stopped reporting fails loudly.
func verifyManifests(manifests []manifest, requiredLanes []string) error {
	byLane := make(map[string][]manifest)
	for _, m := range manifests {
		byLane[m.Lane] = append(byLane[m.Lane], m)
	}
	for _, lane := range requiredLanes {
		if len(byLane[lane]) == 0 {
			return fmt.Errorf("lane %q has no shard manifests", lane)
		}
	}
	lanes := make([]string, 0, len(byLane))
	for lane := range byLane {
		lanes = append(lanes, lane)
	}
	sort.Strings(lanes)
	for _, lane := range lanes {
		if err := verifyLane(lane, byLane[lane]); err != nil {
			return err
		}
	}
	return nil
}

func verifyLane(lane string, group []manifest) error {
	first := group[0]
	if first.Total < 1 || len(group) != first.Total {
		return fmt.Errorf("lane %q: %d manifests for a plan of %d shards", lane, len(group), first.Total)
	}
	owner := make(map[int]bool, len(group))
	assigned := 0
	for _, m := range group {
		switch {
		case m.Total != first.Total || m.Tests != first.Tests || m.Universe != first.Universe:
			return fmt.Errorf("lane %q: shard %d was cut from a different plan (total %d/%d, tests %d/%d, universe %.12s/%.12s)",
				lane, m.Shard, m.Total, first.Total, m.Tests, first.Tests, m.Universe, first.Universe)
		case m.Shard >= m.Total || owner[m.Shard]:
			return fmt.Errorf("lane %q: shard index %d is out of range or reported twice", lane, m.Shard)
		case m.Effective < 1 || m.Effective > m.Assigned:
			return fmt.Errorf("lane %q: shard %d would run %d of its %d tests", lane, m.Shard, m.Effective, m.Assigned)
		}
		owner[m.Shard] = true
		assigned += m.Assigned
	}
	if assigned != first.Tests {
		return fmt.Errorf("lane %q: shards own %d tests, the package has %d", lane, assigned, first.Tests)
	}
	return nil
}
