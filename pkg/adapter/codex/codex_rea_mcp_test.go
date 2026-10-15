package codex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// codexReaX1 is the literal managed table of the MCP Entry Shape Contract
// (acceptance S4 row X1).
const codexReaX1 = "[mcp_servers.rea]\n" +
	"command = \"npx\"\n" +
	"args = [\"-y\", \"rea-agents@6.5.0\", \"mcp\"]\n" +
	"startup_timeout_sec = 30\n"

func reaCodexConfig(action config.ReaMCPAction) *config.HarnessConfig {
	cfg := config.DefaultFullConfig("rea")
	cfg.Runtime.ReaMCP = action
	return cfg
}

// reaCodexAdapter skips the codex CLI probes so the rendered config does not
// depend on the machine running the test.
func reaCodexAdapter(root string) *Adapter {
	a := NewWithRoot(root)
	a.codexCatalogProbed = true
	a.codexVersionProbed = true
	return a
}

func readReaCodexConfig(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, codexConfigRelPath))
	require.NoError(t, err)
	return string(data)
}

func writeReaCodexConfig(t *testing.T, root, text string) {
	t.Helper()
	path := filepath.Join(root, codexConfigRelPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(text), 0o644))
}

// mergeHeaderSections replays the line loop of mergeCodexConfig and returns
// the section name of every line that loop treats as a table header.
func mergeHeaderSections(text string) []string {
	var sections []string
	var scan codexTOMLScanState
	for _, line := range strings.Split(text, "\n") {
		if scan.skipSyntaxLine(line) {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if section, ok := parseStructuralCodexSection(trimmed); ok {
			sections = append(sections, section)
			continue
		}
		if _, value, ok := parseCodexConfigAssignment(trimmed); ok {
			scan.observeValue(value)
		}
	}
	return sections
}

func TestReaCodexHeaderDetectionMatchesMerge(t *testing.T) {
	t.Parallel()
	base := "approval_policy = \"on-request\"\n\n[mcp_servers.context7]\ncommand = \"npx\"\n"
	corpus := map[string]string{
		"plain":                 base + "\n" + codexReaX1,
		"basic multiline":       base + "note = \"\"\"\n" + codexReaX1 + "\"\"\"\n",
		"literal multiline":     base + "note = '''\n" + codexReaX1 + "'''\n",
		"escaped delimiter":     base + "note = \"\"\"\nstill open \\\"\"\"\n" + codexReaX1 + "\"\"\"\n",
		"one-line multiline":    base + "note = \"\"\"one line\"\"\"\n\n" + codexReaX1,
		"commented header":      base + "# " + strings.ReplaceAll(strings.TrimSuffix(codexReaX1, "\n"), "\n", "\n# ") + "\n",
		"hash inside string":    base + "note = \"a # [mcp_servers.rea]\"\n\n" + codexReaX1,
		"after literal closes":  base + "note = '''\nx\n'''\n\n" + codexReaX1,
		"header with comment":   base + "\n" + strings.Replace(codexReaX1, "[mcp_servers.rea]", "[mcp_servers.rea] # mine", 1),
		"multiline in rea body": base + "\n[mcp_servers.rea]\nnote = \"\"\"\n[mcp_servers.rea.env]\n\"\"\"\n",
	}
	for name, doc := range corpus {
		mergeSees := false
		for _, section := range mergeHeaderSections(doc) {
			mergeSees = mergeSees || section == "mcp_servers.rea" || strings.HasPrefix(section, "mcp_servers.rea.")
		}
		class := reamcp.ClassifyCodexTOML(doc)
		assert.Equal(t, mergeSees, class != reamcp.ClassAbsent, name)

		validInput := validateCodexTOMLStructure(doc) == nil
		for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
			out, changed := reamcp.ApplyCodexTOML(doc, action)
			if validInput {
				require.NoError(t, validateCodexTOMLStructure(out), "%s %s", name, action)
			}
			if !changed {
				continue
			}
			reaHeaders := 0
			for _, section := range mergeHeaderSections(out) {
				if section == "mcp_servers.rea" {
					reaHeaders++
				}
			}
			want := 1
			if action == config.ReaMCPRemove {
				want = 0
			}
			assert.Equal(t, want, reaHeaders, "%s %s", name, action)
		}
	}
}

func TestReaMCP_CodexGenerateAddAppendsOneManagedTable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := reaCodexAdapter(root)
	_, err := a.Generate(context.Background(), reaCodexConfig(config.ReaMCPAdd))
	require.NoError(t, err)

	text := readReaCodexConfig(t, root)
	assert.True(t, strings.HasSuffix(text, "\n\n"+codexReaX1), text)
	assert.Equal(t, 1, strings.Count(text, "[mcp_servers.rea]"))
	assert.Equal(t, reamcp.ClassManaged, reamcp.ClassifyCodexTOML(text))

	// The base merge keeps the table: later add and keep runs write the same bytes.
	for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPKeep} {
		_, err = a.Generate(context.Background(), reaCodexConfig(action))
		require.NoError(t, err)
		assert.Equal(t, text, readReaCodexConfig(t, root), action)
	}
}

func TestReaMCP_CodexAddThenRemoveRestoresBaseMergeBytes(t *testing.T) {
	t.Parallel()
	userTable := "[mcp_servers.mine]\ncommand = \"mine-mcp\"  # keep\n"
	keepRoot, reaRoot := t.TempDir(), t.TempDir()
	for _, root := range []string{keepRoot, reaRoot} {
		_, err := reaCodexAdapter(root).Generate(context.Background(), reaCodexConfig(config.ReaMCPKeep))
		require.NoError(t, err)
		writeReaCodexConfig(t, root, readReaCodexConfig(t, root)+userTable)
	}
	_, err := reaCodexAdapter(keepRoot).Generate(context.Background(), reaCodexConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	base := readReaCodexConfig(t, keepRoot)
	assert.NotContains(t, base, "rea-agents")

	a := reaCodexAdapter(reaRoot)
	_, err = a.Generate(context.Background(), reaCodexConfig(config.ReaMCPAdd))
	require.NoError(t, err)
	added := readReaCodexConfig(t, reaRoot)
	assert.Equal(t, base+"\n"+codexReaX1, added)
	assert.Contains(t, added, userTable)

	viaFile, changed, err := reamcp.ApplyFile(reamcp.Codex, []byte(base), config.ReaMCPAdd)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, added, string(viaFile))

	_, err = a.Generate(context.Background(), reaCodexConfig(config.ReaMCPRemove))
	require.NoError(t, err)
	assert.Equal(t, base, readReaCodexConfig(t, reaRoot))
}

func TestReaMCP_CodexMergeLeavesOtherPinAndUserOwnedTablesByteIdentical(t *testing.T) {
	t.Parallel()
	rendered, err := NewWithRoot(t.TempDir()).prepareConfigFile(reaCodexConfig(config.ReaMCPKeep))
	require.NoError(t, err)
	base := string(rendered[0].Content)
	rows := map[string]string{
		"X2": "[mcp_servers.rea]\ncommand = \"npx\"\nargs = [\"-y\", \"rea-agents@6.5.0\", \"mcp\"]\n",
		"X3": codexReaX1 + "tool_timeout_sec = 120\n",
		"X4": "[mcp_servers.rea]\n# mine\ncommand = \"npx\"\nargs = [\"-y\", \"rea-agents@6.5.0\", \"mcp\"]\nstartup_timeout_sec = 30\n",
		"X5": strings.Replace(codexReaX1, "6.5.0", "6.4.0", 1),
		"X7": codexReaX1 + "\n[mcp_servers.rea.env]\nA = \"1\"\n",
	}
	prepare := func(root string, action config.ReaMCPAction) string {
		t.Helper()
		files, err := NewWithRoot(root).prepareConfigFile(reaCodexConfig(action))
		require.NoError(t, err)
		return string(files[0].Content)
	}
	for id, table := range rows {
		root := t.TempDir()
		writeReaCodexConfig(t, root, base+"\n"+table)
		require.NotEqual(t, reamcp.ClassManaged, reamcp.ClassifyCodexTOML(base+"\n"+table), id)
		// The oracle is the base merge of the same file: REA changes nothing.
		keep := prepare(root, config.ReaMCPKeep)
		require.True(t, strings.HasSuffix(keep, "\n\n"+table), id)
		for _, action := range []config.ReaMCPAction{config.ReaMCPAdd, config.ReaMCPRemove} {
			assert.Equal(t, keep, prepare(root, action), "%s %s", id, action)
		}
	}

	// X6: rea-agents is another server; add appends rea and keeps it as is.
	x6 := strings.Replace(codexReaX1, "[mcp_servers.rea]", "[mcp_servers.rea-agents]", 1)
	root := t.TempDir()
	writeReaCodexConfig(t, root, base+"\n"+x6)
	keep := prepare(root, config.ReaMCPKeep)
	require.True(t, strings.HasSuffix(keep, "\n\n"+x6))
	assert.Equal(t, keep+"\n"+codexReaX1, prepare(root, config.ReaMCPAdd))
}

func TestReaMCP_CodexCleanRemovesManagedTableBeforeGenericClean(t *testing.T) {
	t.Parallel()
	keepRoot, reaRoot, userRoot := t.TempDir(), t.TempDir(), t.TempDir()
	for root, action := range map[string]config.ReaMCPAction{
		keepRoot: config.ReaMCPKeep, reaRoot: config.ReaMCPAdd, userRoot: config.ReaMCPKeep,
	} {
		_, err := reaCodexAdapter(root).Generate(context.Background(), reaCodexConfig(action))
		require.NoError(t, err)
	}
	userTable := codexReaX1 + "tool_timeout_sec = 120\n"
	writeReaCodexConfig(t, userRoot, readReaCodexConfig(t, userRoot)+"\n"+userTable)

	for _, root := range []string{keepRoot, reaRoot, userRoot} {
		require.NoError(t, reaCodexAdapter(root).Clean(context.Background()))
	}
	cleaned := readReaCodexConfig(t, keepRoot)
	assert.Equal(t, cleaned, readReaCodexConfig(t, reaRoot))
	assert.NotContains(t, readReaCodexConfig(t, reaRoot), "mcp_servers.rea")
	assert.Contains(t, readReaCodexConfig(t, userRoot), userTable)
}

func TestReaMCP_CodexRefusesSymlinkedTargetWhenReaChangesContent(t *testing.T) {
	t.Parallel()
	root, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "config.toml")
	original := []byte("approval_policy = \"never\"\n")
	require.NoError(t, os.WriteFile(target, original, 0o644))
	if err := os.Symlink(outside, filepath.Join(root, ".codex")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	_, err := reaCodexAdapter(root).Generate(context.Background(), reaCodexConfig(config.ReaMCPAdd))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "REA MCP")
	assert.Contains(t, err.Error(), ".codex/config.toml")
	data, readErr := os.ReadFile(target)
	require.NoError(t, readErr)
	assert.Equal(t, original, data)
}
