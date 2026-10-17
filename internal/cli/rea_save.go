package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/config"
)

// reaKeyPath is the autopus.yaml path of the stored REA decision.
var reaKeyPath = []string{"features", "rea", "enabled"}

// reaKeyRefusal is a features.rea.enabled write the line splice cannot make
// safely. Nothing is written: like the review writer's refusal
// (SPEC-REVMODEL-001), it never falls back to re-encoding the whole file.
type reaKeyRefusal struct{ enabled bool }

func (e reaKeyRefusal) Error() string {
	return fmt.Sprintf("refuse to edit features.rea.enabled in autopus.yaml: the key cannot be changed as plain "+
		"lines alone; nothing was written, set features.rea.enabled: %t by hand", e.enabled)
}

// saveKey stores enabled as features.rea.enabled after the run's platform
// writes. Init has just written autopus.yaml whole from cfg and saves cfg
// again; every other run edits only the key in the existing bytes.
func (r *reaRun) saveKey(dir string, cfg *config.HarnessConfig, enabled bool) error {
	if r.wholeSave {
		cfg.Features.Rea.Enabled = &enabled
		return config.Save(dir, cfg)
	}
	return saveReaKey(dir, cfg, enabled)
}

// saveReaKey writes features.rea.enabled into dir's autopus.yaml by the Line
// Splice Writer Contract: the bytes after pruneRetiredConfig with only the
// key's value replaced, or with the missing features, rea, and enabled lines
// added, checked by verifyReaKeyEdit and written in one atomic replace that
// keeps the file mode. Key order, comments, quoting, and blank lines stay
// byte for byte. A refusal or failure leaves the file and cfg as they were;
// after the write cfg holds the key, so no later save of the run drops it.
func saveReaKey(dir string, cfg *config.HarnessConfig, enabled bool) error {
	path := filepath.Join(dir, "autopus.yaml")
	original, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	// The retired orchestra keys never ride along (SPEC-PANERM-001 REQ-10).
	source, _, err := pruneRetiredConfig(original)
	if err != nil {
		return err
	}
	edited, ok := spliceReaKey(source, enabled)
	if !ok {
		return reaKeyRefusal{enabled: enabled}
	}
	if err := verifyReaKeyEdit(source, edited, enabled); err != nil {
		return err
	}
	if err := atomicWriteQualityConfig(path, edited); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	cfg.Features.Rea.Enabled = &enabled
	return nil
}

// spliceReaKey sets features.rea.enabled in source. Every mapping on the way
// must be a block mapping that no alias names and that holds no merge key, so
// nothing else in the file shares or overrides what the edit changes; the key
// and its value carry no anchor or alias. A present value token is replaced
// on its line, keeping a tag and an inline comment. Missing entries are added
// after the last line of the deepest mapping that exists, one nesting step
// deeper per level. It reports false for anything else.
func spliceReaKey(source []byte, enabled bool) ([]byte, bool) {
	// A line break yaml.v3 sees but a split at \n does not moves node lines
	// away from the lines the splice cuts, as for the review writer.
	if hasYAMLOnlyLineBreak(source) {
		return nil, false
	}
	document, _, err := parseAutopusConfigDocument(source)
	if err != nil {
		return nil, false
	}
	aliased := yamlAliasTargets(document)
	lines := strings.SplitAfter(string(source), "\n")
	newline := detectConfigNewline(source)
	text := strconv.FormatBool(enabled)
	root := document.Content[0]
	node, parent := root, (*yaml.Node)(nil)
	for depth, step := range reaKeyPath {
		if node.Kind != yaml.MappingNode || node.Style&yaml.FlowStyle != 0 || aliased[node] || yamlMappingHasMerge(node) {
			return nil, false
		}
		key, value := yamlMappingPair(node, step)
		var edit reviewLineEdit
		var ok bool
		switch {
		case key == nil:
			texts := reaKeyTexts(reaKeyPath[depth:], text, reaIndentStep(root, parent, node))
			edit, ok = insertReviewKeys(node, lines, newline, texts)
		case depth < len(reaKeyPath)-1:
			node, parent = value, key
			continue
		case key.Anchor != "" || value.Anchor != "" || value.Kind == yaml.AliasNode || aliased[value]:
			return nil, false
		default:
			edit, ok = spliceReviewScalarText(key, value, lines, text)
		}
		if !ok {
			return nil, false
		}
		return applyReviewLineEdits(lines, []reviewLineEdit{edit}, newline), true
	}
	return nil, false
}

// reaKeyTexts returns the lines that add path, relative to the indentation of
// the mapping they go into: one `<key>:` line per level, each nested one step
// deeper, the last one holding value.
func reaKeyTexts(path []string, value, step string) []string {
	texts := make([]string, len(path))
	for i, key := range path {
		texts[i] = strings.Repeat(step, i) + key + ":"
	}
	texts[len(path)-1] += " " + value
	return texts
}

// reaIndentStep is the indentation one nesting level adds in the file: from
// parent's key to node's keys when the lines go below a key, else from the
// first root key whose value is a block mapping to that mapping's keys, else
// two spaces.
func reaIndentStep(root, parent, node *yaml.Node) string {
	if parent != nil && node.Column > parent.Column {
		return strings.Repeat(" ", node.Column-parent.Column)
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		key, value := root.Content[i], root.Content[i+1]
		if value.Kind == yaml.MappingNode && value.Style&yaml.FlowStyle == 0 && value.Column > key.Column {
			return strings.Repeat(" ", value.Column-key.Column)
		}
	}
	return "  "
}

// verifyReaKeyEdit is the whole-document check of the REA key splice, the
// way verifyReviewLineEdit checks the review writer. features.rea.enabled
// must hold enabled as a bool; with the outermost entry the edit added, or
// the key it changed, set aside, edited must parse to the same data as
// source, each alias compared by the data it resolves to; and edited must
// load, strictly and validated, to source's load with only the key changed.
// Any mismatch refuses the write.
func verifyReaKeyEdit(source, edited []byte, enabled bool) error {
	refusal := reaKeyRefusal{enabled: enabled}
	var before, after yaml.Node
	if err := yaml.Unmarshal(source, &before); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	if yaml.Unmarshal(edited, &after) != nil {
		return refusal
	}
	held := yamlPathValue(&after, reaKeyPath)
	if held == nil || held.Kind != yaml.ScalarNode || held.ShortTag() != "!!bool" ||
		held.Value != strconv.FormatBool(enabled) {
		return refusal
	}
	set := reaKeyPath
	for depth := range reaKeyPath {
		if yamlPathValue(&before, reaKeyPath[:depth+1]) == nil {
			set = reaKeyPath[:depth+1]
			break
		}
	}
	dropYAMLPathEntry(&before, set)
	dropYAMLPathEntry(&after, set)
	if !sameYAMLData(&before, &after) {
		return refusal
	}
	want, err := config.ParseConfig(source)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	got, err := config.ParseConfig(edited)
	if err != nil {
		return refusal
	}
	want.Features.Rea.Enabled = &enabled
	if !reflect.DeepEqual(want, got) {
		return refusal
	}
	return nil
}
