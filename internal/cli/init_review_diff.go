package cli

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/config"
)

// initReviewLeaf is one leaf of a marshalled config: its parent path and
// key. A key is never split, so a key holding a dot keeps it.
type initReviewLeaf struct {
	parent, key string
}

// initReviewChanges are the 바뀌는 설정 lines of the init review (Init
// Review Contract): base and current marshalled as config.Save writes them,
// every mapping walked down to its leaves, a list taken whole in flow form.
// An added, changed, or removed leaf is grouped under its parent path, one
// line per parent with its keys in order, and a top-level leaf takes a line
// of its own; a removed leaf reads none. Lines are in path order, and every
// value passes through terminalSafe.
func initReviewChanges(base, current *config.HarnessConfig, none string) ([]string, error) {
	before, err := initReviewLeaves(base)
	if err != nil {
		return nil, err
	}
	after, err := initReviewLeaves(current)
	if err != nil {
		return nil, err
	}
	changed := map[string][][2]string{}
	note := func(leaf initReviewLeaf, value string) {
		changed[leaf.parent] = append(changed[leaf.parent], [2]string{terminalSafe(leaf.key), terminalSafe(value)})
	}
	for leaf, value := range after {
		if old, ok := before[leaf]; !ok || old != value {
			note(leaf, value)
		}
	}
	for leaf := range before {
		if _, ok := after[leaf]; !ok {
			note(leaf, none)
		}
	}
	type line struct{ path, text string }
	var lines []line
	for parent, pairs := range changed {
		slices.SortFunc(pairs, func(a, b [2]string) int { return strings.Compare(a[0], b[0]) })
		if parent == "" {
			for _, pair := range pairs {
				lines = append(lines, line{pair[0], "  " + pair[0] + ": " + pair[1]})
			}
			continue
		}
		joined := make([]string, len(pairs))
		for i, pair := range pairs {
			joined[i] = pair[0] + "=" + pair[1]
		}
		path := terminalSafe(parent)
		lines = append(lines, line{path, "  " + path + ": " + strings.Join(joined, ", ")})
	}
	slices.SortFunc(lines, func(a, b line) int { return strings.Compare(a.path, b.path) })
	texts := make([]string, len(lines))
	for i, l := range lines {
		texts[i] = l.text
	}
	return texts, nil
}

// initReviewLeaves marshals cfg as config.Save does and lists its leaves.
func initReviewLeaves(cfg *config.HarnessConfig) (map[initReviewLeaf]string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("read marshalled config: %w", err)
	}
	leaves := map[initReviewLeaf]string{}
	if len(doc.Content) > 0 {
		initReviewWalk(doc.Content[0], "", leaves)
	}
	return leaves, nil
}

// initReviewWalk adds the leaves under one mapping node at path.
func initReviewWalk(node *yaml.Node, path string, leaves map[initReviewLeaf]string) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i].Value, node.Content[i+1]
		if value.Kind == yaml.MappingNode {
			child := key
			if path != "" {
				child = path + "." + key
			}
			initReviewWalk(value, child, leaves)
			continue
		}
		leaves[initReviewLeaf{parent: path, key: key}] = initReviewFlow(value)
	}
}

// initReviewFlow writes a node on one line: a list as [a, b], a mapping
// inside a list as {k: v}, a scalar as its value.
func initReviewFlow(node *yaml.Node) string {
	switch node.Kind {
	case yaml.SequenceNode:
		items := make([]string, len(node.Content))
		for i, item := range node.Content {
			items[i] = initReviewFlow(item)
		}
		return "[" + strings.Join(items, ", ") + "]"
	case yaml.MappingNode:
		pairs := make([]string, 0, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			pairs = append(pairs, node.Content[i].Value+": "+initReviewFlow(node.Content[i+1]))
		}
		return "{" + strings.Join(pairs, ", ") + "}"
	case yaml.AliasNode:
		return initReviewFlow(node.Alias)
	}
	return node.Value
}
