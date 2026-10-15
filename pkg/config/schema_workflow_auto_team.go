package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// checkWorkflowAutoTeamNode rejects a workflow.auto_team whose YAML node is not
// tagged !!bool. The typed decode alone cannot: yaml.v3 turns YAML 1.1 words
// such as yes or "off" into a bool and leaves null, ~, and an empty value as a
// nil *bool without an error, so a typo would silently read as a setting. The
// check walks the parsed document, whose nodes keep the user's line numbers,
// and follows aliases and merge keys the way the decoder does.
func checkWorkflowAutoTeamNode(doc *yaml.Node) error {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 {
		return nil
	}
	for _, workflow := range yamlMappingEntries(doc.Content[0], "workflow") {
		for _, entry := range yamlMappingEntries(workflow[1], "auto_team") {
			if entry[1].ShortTag() != "!!bool" {
				return fmt.Errorf("workflow.auto_team: must be true or false (line %d)", entry[0].Line)
			}
		}
	}
	return nil
}

// yamlMappingEntries returns the key and value nodes of every name entry of a
// mapping node, including the entries its merge keys ("<<") bring in. Each
// mapping is walked once, so shared aliases cannot multiply the work.
func yamlMappingEntries(node *yaml.Node, name string) [][2]*yaml.Node {
	var entries [][2]*yaml.Node
	visited := map[*yaml.Node]bool{}
	var walk func(*yaml.Node)
	walk = func(node *yaml.Node) {
		for node != nil && node.Kind == yaml.AliasNode {
			node = node.Alias
		}
		if node == nil || node.Kind != yaml.MappingNode || visited[node] {
			return
		}
		visited[node] = true
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			if isYAMLMergeKey(key) {
				sources := []*yaml.Node{value}
				if value.Kind == yaml.SequenceNode {
					sources = value.Content
				}
				for _, source := range sources {
					walk(source)
				}
				continue
			}
			// The decoder names a field by decoding its key, which resolves an
			// alias key and a !!binary key alike, so the key is read that way.
			var decoded string
			if key.Decode(&decoded) == nil && decoded == name {
				entries = append(entries, [2]*yaml.Node{key, value})
			}
		}
	}
	walk(node)
	return entries
}

// isYAMLMergeKey mirrors yaml.v3's isMerge: only a "<<" scalar node merges,
// so an alias that resolves to "<<" stays an ordinary key.
func isYAMLMergeKey(key *yaml.Node) bool {
	return key.Kind == yaml.ScalarNode && key.Value == "<<" && (key.Tag == "" || key.Tag == "!" || key.ShortTag() == "!!merge")
}
