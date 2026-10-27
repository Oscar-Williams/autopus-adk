package omp

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

const (
	ompNativeSmokeToken    = "AUTOPUS_OMP_NATIVE_TASK_HUB_SMOKE"
	ompNativeCredential    = "AUTOPUS_NATIVE_CREDENTIAL_MUST_NOT_LEAK"
	ompNativeAlphaID       = "NativeAlpha"
	ompNativeBetaID        = "NativeBeta"
	ompNativeSharedContext = "Read-only native lifecycle smoke. Do not modify files or use process or network tools. Your final message is the required five-field JSON receipt and nothing else."
	ompNativeAlphaTask     = "NATIVE_CHILD_ALPHA: inspect `.omp/skills/agent-pipeline/SKILL.md` read-only and return its strict receipt naming exactly that path without modifying anything."
	ompNativeBetaTask      = "NATIVE_CHILD_BETA: inspect `.omp/rules/autopus-branding.md` read-only and return its strict receipt naming exactly that path without modifying anything."
	// ompNativeSolutionSpace fills the item field omp 18.3+ requires on every
	// task; nothing about the smoke's children is open to decide.
	ompNativeSolutionSpace = "one fix: path given, receipt fields given"
)

type ompNativeChildReceipt struct {
	OwnedPaths       []string `json:"owned_paths"`
	ChangedFiles     []string `json:"changed_files"`
	Verification     []string `json:"verification"`
	Blockers         []string `json:"blockers"`
	NextRequiredStep string   `json:"next_required_step"`
}

// ompNativeTaskArguments is the one batch the smoke issues. omp 18.3+ has no
// per-task outputSchema or schemaMode: a child's final message is its result,
// so the receipt shape is asked for in the shared context and verified where
// the parent receives it (ompNativeRequireOutputReceipt).
func ompNativeTaskArguments() map[string]any {
	return map[string]any{
		"i":       "Spawning read-only lifecycle probes",
		"context": ompNativeSharedContext,
		"tasks": []any{
			map[string]any{
				"name": ompNativeAlphaID, "agent": "scout", "task": ompNativeAlphaTask,
				"solutionSpace": ompNativeSolutionSpace,
			},
			map[string]any{
				"name": ompNativeBetaID, "agent": "reviewer", "task": ompNativeBetaTask,
				"solutionSpace": ompNativeSolutionSpace,
			},
		},
	}
}

func ompNativeExpectedReceipt(id string) ompNativeChildReceipt {
	switch id {
	case ompNativeAlphaID:
		return ompNativeChildReceipt{
			OwnedPaths: []string{".omp/skills/agent-pipeline/SKILL.md"}, ChangedFiles: []string{},
			Verification: []string{"native-alpha-read-only"}, Blockers: []string{}, NextRequiredStep: "none",
		}
	case ompNativeBetaID:
		return ompNativeChildReceipt{
			OwnedPaths: []string{".omp/rules/autopus-branding.md"}, ChangedFiles: []string{},
			Verification: []string{"native-beta-read-only"}, Blockers: []string{}, NextRequiredStep: "none",
		}
	default:
		panic("unknown native child receipt " + id)
	}
}

func validateOMPNativeTaskArguments(args any) error {
	actual, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("encode task arguments: %w", err)
	}
	expectedArgs := ompNativeTaskArguments()
	delete(expectedArgs, "i")
	expected, err := json.Marshal(expectedArgs)
	if err != nil {
		return fmt.Errorf("encode expected task arguments: %w", err)
	}
	if string(actual) != string(expected) {
		return fmt.Errorf("task arguments drifted: %s", actual)
	}
	return nil
}

// validateOMPNativeParentTools pins the parent surface omp/18.12.3 offers for
// `--tools task,wait`. The task item carries `model` and `tools` since omp
// 18.3 let a parent pick a child's model or tool list; the smoke's batch sets
// neither, and the task tool's progress details prove both children ran as
// bundled agents without an override (ompNativeRPCTrace.childSources).
func validateOMPNativeParentTools(tools []json.RawMessage) error {
	if names := ompNativeToolNames(tools); len(names) != 2 || !names["task"] || !names["wait"] {
		return fmt.Errorf("parent tools drifted: %v", names)
	}
	parameters, ok := ompNativeToolParameters(tools, "task")
	if !ok {
		return fmt.Errorf("task tool missing")
	}
	properties := ompNativeMap(parameters["properties"])
	if !ompNativeExactKeys(properties, "context", "i", "tasks") ||
		!ompNativeStringSet(parameters["required"], "context", "i", "tasks") {
		encoded, _ := json.Marshal(properties)
		return fmt.Errorf("task batch shape drifted: properties=%s", encoded)
	}
	tasks := ompNativeMap(properties["tasks"])
	items := ompNativeMap(tasks["items"])
	itemProperties := ompNativeMap(items["properties"])
	if !ompNativeExactKeys(itemProperties, "agent", "model", "name", "solutionSpace", "sourceSession", "task", "tools") ||
		!ompNativeStringSet(items["required"], "solutionSpace", "task") {
		encoded, _ := json.Marshal(items)
		return fmt.Errorf("task item shape drifted: items=%s", encoded)
	}
	listed, err := ompTaskDefinitionAgentNames(ompNativeToolDefinition(tools, "task"))
	if err != nil {
		return fmt.Errorf("task definition: %w", err)
	}
	for _, agent := range []string{"scout", "reviewer"} {
		if !slices.Contains(listed, agent) {
			return fmt.Errorf("bundled agent %s missing from task definition: %v", agent, listed)
		}
	}
	wait, ok := ompNativeToolParameters(tools, "wait")
	if !ok {
		return fmt.Errorf("wait tool missing")
	}
	additional, isBool := wait["additionalProperties"].(bool)
	if !ompNativeExactKeys(ompNativeMap(wait["properties"]), "i") || !isBool || additional {
		encoded, _ := json.Marshal(wait)
		return fmt.Errorf("wait shape drifted: %s", encoded)
	}
	return nil
}

// ompNativeChildToolContract is the toolset a spawned child must present, per
// bundled agent, as OMP declares it: `scout` investigates without a shell,
// `reviewer` gets bash and, since it declares `spawns: [scout]`, task and
// wait. `find`, `lsp`, and `ast_grep` sit in the bundled definitions but are
// runtime gated (an index, a language server, an installed ast-grep), so
// they are not required; a superset still passes, the write set never may.
// omp 18.3+ gives a child neither `hub` nor `yield`: its final message is its
// result, and `wait` replaced `hub` for the one child that may spawn.
//
// Measured against omp/18.12.3 (`omp agents unpack` and the child requests
// the fake provider receives).
var ompNativeChildToolContract = map[string]struct {
	required  []string
	forbidden []string
}{
	ompNativeAlphaID: {
		required:  []string{"glob", "grep", "read", "web_search"},
		forbidden: []string{"bash", "edit", "hub", "task", "wait", "write", "yield"},
	},
	ompNativeBetaID: {
		required:  []string{"bash", "glob", "grep", "read", "task", "wait", "web_search"},
		forbidden: []string{"edit", "hub", "write", "yield"},
	},
}

func validateOMPNativeChildTools(child string, tools []json.RawMessage) error {
	contract, known := ompNativeChildToolContract[child]
	if !known {
		return fmt.Errorf("no tool contract for child %s", child)
	}
	names := ompNativeToolNames(tools)
	for _, required := range contract.required {
		if !names[required] {
			return fmt.Errorf("%s child tool %s missing", child, required)
		}
	}
	for _, forbidden := range contract.forbidden {
		if names[forbidden] {
			return fmt.Errorf("%s child exposed %s", child, forbidden)
		}
	}
	return nil
}

// ompNativeRequireOutputReceipt checks that the parent received the child's
// final message verbatim: omp renders a settled job as a `<task-result>`
// whose `<output>` block is the text the child ended on, and that text must
// decode to exactly the expected five-field receipt.
func ompNativeRequireOutputReceipt(text, id, agent string) error {
	header := fmt.Sprintf(`<task-result id=%q agent=%q status="completed"`, id, agent)
	_, after, found := strings.Cut(text, header)
	if !found {
		return fmt.Errorf("%s completed task-result missing", id)
	}
	_, after, found = strings.Cut(after, "<output>")
	if !found {
		return fmt.Errorf("%s output missing", id)
	}
	raw, _, found := strings.Cut(after, "</output>")
	if !found {
		return fmt.Errorf("%s output unterminated", id)
	}
	var got ompNativeChildReceipt
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &got); err != nil {
		return fmt.Errorf("%s output is not the JSON receipt: %w", id, err)
	}
	actual, _ := json.Marshal(got)
	expected, _ := json.Marshal(ompNativeExpectedReceipt(id))
	if string(actual) != string(expected) {
		return fmt.Errorf("%s receipt drifted: %s", id, actual)
	}
	return nil
}

func ompNativeToolNames(tools []json.RawMessage) map[string]bool {
	result := make(map[string]bool, len(tools))
	for _, raw := range tools {
		var tool struct {
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
		}
		if json.Unmarshal(raw, &tool) == nil && tool.Function.Name != "" {
			result[tool.Function.Name] = true
		}
	}
	return result
}
func ompNativeToolDefinition(tools []json.RawMessage, name string) string {
	for _, raw := range tools {
		var tool struct {
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
		}
		if json.Unmarshal(raw, &tool) == nil && tool.Function.Name == name {
			return string(raw)
		}
	}
	return ""
}

func ompNativeToolParameters(tools []json.RawMessage, name string) (map[string]any, bool) {
	for _, raw := range tools {
		var tool struct {
			Function struct {
				Name       string         `json:"name"`
				Parameters map[string]any `json:"parameters"`
			} `json:"function"`
		}
		if json.Unmarshal(raw, &tool) == nil && tool.Function.Name == name {
			return tool.Function.Parameters, true
		}
	}
	return nil, false
}

func ompNativeMap(value any) map[string]any {
	result, _ := value.(map[string]any)
	return result
}

func ompNativeExactKeys(values map[string]any, expected ...string) bool {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sort.Strings(expected)
	return strings.Join(keys, "\x00") == strings.Join(expected, "\x00")
}

func ompNativeStringSet(value any, expected ...string) bool {
	items, ok := value.([]any)
	if !ok || len(items) != len(expected) {
		return false
	}
	actual := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return false
		}
		actual = append(actual, text)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	return strings.Join(actual, "\x00") == strings.Join(expected, "\x00")
}
