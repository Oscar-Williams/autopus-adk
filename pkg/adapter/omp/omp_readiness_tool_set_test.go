package omp

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ompReleaseBuiltinTools are the names each release accepts in --tools, as
// its "Unknown tool in --tools" error lists them. Any other name exits the
// release before the RPC server starts, so every rpc.* capability reads
// exit_nonzero. omp/18.3.0 replaced hub with wait.
var ompReleaseBuiltinTools = map[string]string{
	"omp/17.2.7": "read, bash, edit, ast_grep, ast_edit, ask, debug, eval, github, glob, grep, lsp, " +
		"inspect_image, browser, computer, checkpoint, rewind, security_scan, task, hub, todo, " +
		"web_search, write, memory_edit, retain, recall, reflect, learn, manage_skill, yield, goal",
	"omp/18.6.1": "read, bash, edit, ast_grep, ast_edit, ask, debug, ida, eval, github, glob, grep, " +
		"find, lsp, checkpoint, rewind, context_notes, new_context, security_scan, task, wait, todo, " +
		"web_search, write, memory_edit, retain, recall, reflect, learn, manage_skill, " +
		"goal, init_experiment, run_experiment, log_experiment, update_notes",
	// The release canary's pin; the CI omp-native-smoke job runs it.
	"omp/18.12.3": "read, bash, edit, ast_grep, ast_edit, ask, debug, ida, eval, github, glob, grep, " +
		"find, lsp, checkpoint, rewind, context_notes, new_context, security_scan, task, wait, todo, " +
		"web_search, write, memory_edit, retain, recall, reflect, learn, manage_skill, " +
		"goal, init_experiment, run_experiment, log_experiment, update_notes",
}

// Guards against enabling a tool one supported release lacks: hub exists only
// before omp/18.3.0 and wait only from it, so neither may reach --tools.
func TestOMPProviderFreeRPCArgs_EnableOnlyToolsEveryReleaseAccepts(t *testing.T) {
	args := ompProviderFreeRPCArgs(".")
	index := slices.Index(args, "--tools")
	require.True(t, index >= 0 && index+1 < len(args), "the probe must restrict its tools")
	requested := strings.Split(args[index+1], ",")
	assert.Equal(t, []string{"task", "todo"}, requested)
	for release, builtins := range ompReleaseBuiltinTools {
		accepted := strings.Split(builtins, ", ")
		for _, tool := range requested {
			assert.Contains(t, accepted, tool, "%s rejects --tools %s before RPC starts", release, tool)
		}
	}
}

// Guards against a pre-intent verdict that hinges on a release-specific tool:
// the get_state dump each release returns for the probe's tools must pass,
// while a dump without parameter schemas cannot prove the pre-intent shape.
func TestParseOMPPreIntentTools_AcceptsEveryReleaseDumpOfTheProbeTools(t *testing.T) {
	batchTask := `{"name":"task","parameters":{"type":"object","properties":{"context":{"type":"string"},"tasks":{"type":"array"}}}}`
	modelTask := `{"name":"task","parameters":{"type":"object","properties":{"context":{"type":"string"},"tasks":{"type":"array"},"model":{"not":true}}}}`
	todo := `{"name":"todo","parameters":{"type":"object","properties":{"op":{"type":"string"},"list":{"type":"array"}}}}`
	wait := `{"name":"wait","parameters":{"type":"object","properties":{}}}`
	hub := `{"name":"hub","parameters":{"type":"object","properties":{"op":{"type":"string"},"ids":{"type":"array"}}}}`

	for _, test := range []struct {
		name  string
		tools []string
		want  bool
	}{
		{name: "omp/18.6.1", tools: []string{modelTask, todo}, want: true},
		{name: "omp/18.1.13", tools: []string{batchTask, todo}, want: true},
		{name: "a dump that adds wait", tools: []string{modelTask, wait, todo}, want: true},
		{name: "a dump that adds hub", tools: []string{batchTask, hub, todo}, want: true},
		{name: "omp/17.2.7 without schemas", tools: []string{
			`{"name":"task","description":"spawn"}`, `{"name":"todo","description":"progress"}`,
		}},
		{name: "task missing", tools: []string{todo, wait}},
		{name: "todo missing", tools: []string{modelTask, wait}},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := make([]json.RawMessage, 0, len(test.tools))
			for _, tool := range test.tools {
				raw = append(raw, json.RawMessage(tool))
			}
			assert.Equal(t, test.want, parseOMPPreIntentTools(raw))
		})
	}
}
