package reamcp

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// M2 (REA security audit): the three lists partition the 207 audited tools of
// the pin, and no static tool belongs to a family that runs, attaches to,
// instruments, or captures a live target.
func TestToolLists_PartitionThePinsAuditedTools(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "6.5.0", Pin, "a new pin needs a new audit of these lists")
	static, runtime, writes := StaticAnalysisTools(), RuntimeCaptureTools(), CallerPathWriteTools()
	assert.Len(t, static, 135)
	assert.Len(t, runtime, 68)
	assert.Len(t, writes, 4)

	seen := map[string]string{}
	for list, tools := range map[string][]string{"static": static, "runtime": runtime, "writes": writes} {
		assert.True(t, slices.IsSorted(tools), list)
		for _, tool := range tools {
			if other, dup := seen[tool]; dup {
				t.Errorf("%s is in %s and %s", tool, other, list)
			}
			seen[tool] = list
		}
	}
	for _, tool := range static {
		for _, family := range []string{"adb", "frida", "gdb", "rizin_debug", "execute_rizin", "cutter", "capture_", "observe_"} {
			assert.False(t, strings.Contains(tool, family), "%s is a %s tool", tool, family)
		}
	}
	for _, tool := range []string{"capture_process_scenario", "capture_browser_scenario", "instrument_with_frida",
		"start_gdb_session", "observe_native_ui", "analyze_web_bundle"} {
		assert.Contains(t, runtime, tool)
	}
	assert.ElementsMatch(t, []string{"export_evidence_bundle", "export_web_scripts", "extract_firmware",
		"recover_javascript_sources"}, writes)

	static[0] = "mutated"
	assert.NotEqual(t, "mutated", StaticAnalysisTools()[0], "callers get a copy")
}
