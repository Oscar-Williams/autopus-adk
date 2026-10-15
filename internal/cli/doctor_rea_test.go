package cli

// SPEC-REA-001 T13 (acceptance S8): the REA rows of the real `auto doctor`
// in JSON and text, over fixture projects.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

// reaP5 lists the five platforms of acceptance fixture P5.
var reaP5 = []string{"claude-code", "codex", "antigravity-cli", "omp", "opencode"}

// reaX5 is acceptance S4 row X5: an other-pin Codex rea table.
const reaX5 = reaCodexBase + "\n[mcp_servers.rea]\ncommand = \"npx\"\nargs = [\"-y\", \"rea-agents@6.4.0\", \"mcp\"]\nstartup_timeout_sec = 30\n"

// reaDoctorJSON runs `auto doctor --json` and returns data.rea, the
// doctor.rea.* checks by ID, and the envelope status. The envelope writes
// object keys in sorted order, so callers compare data.rea as JSON values.
func reaDoctorJSON(t *testing.T, dir string) (string, map[string]jsonCheck, string) {
	t.Helper()
	out, _ := runReaCLI(t, "", "doctor", "--dir", dir, "--json")
	var envelope struct {
		Status string `json:"status"`
		Data   struct {
			Rea json.RawMessage `json:"rea"`
		} `json:"data"`
		Checks []jsonCheck `json:"checks"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &envelope), out)
	var compact bytes.Buffer
	require.NoError(t, json.Compact(&compact, envelope.Data.Rea), out)
	checks := map[string]jsonCheck{}
	for _, check := range envelope.Checks {
		if strings.HasPrefix(check.ID, "doctor.rea.") {
			checks[check.ID] = check
		}
	}
	return compact.String(), checks, envelope.Status
}

// S8 (a), (b), (c): data.rea and the text decision line for a managed
// project, the same project turned off, and an undecided one.
func TestDoctorRea_DecisionStates(t *testing.T) {
	recorder := reaCLIEnv(t, "v24.11.0\n")
	managed, _ := reaFixture(t, reaP5...)
	writeManagedRea(t, managed, reamcp.Platforms()...)
	off, offCfg := reaFixture(t, reaP5...)
	offCfg.Features.Rea.Enabled = reaBool(false)
	require.NoError(t, config.Save(off, offCfg))
	undecided, _ := reaFixture(t, reaP5...)
	absent := `"platforms":{"antigravity-cli":"absent","claude-code":"absent","codex":"absent"}`
	cases := []struct {
		name, dir, data, line string
	}{
		{"(a) managed", managed, `{"decision":"on","decision_source":"managed_entry","node":{"status":"pass","version":"v24.11.0"},` +
			`"platforms":{"antigravity-cli":"managed","claude-code":"managed","codex":"managed"},"guidance":["omp","opencode"]}`, "REA: 켜짐 (managed 항목)"},
		{"(b) off", off, `{"decision":"off","decision_source":"features.rea","node":{"status":"pass","version":"v24.11.0"},` + absent + `,"guidance":[]}`, "REA: 꺼짐 (autopus.yaml)"},
		{"(c) undecided", undecided, `{"decision":"undecided","decision_source":"none","node":{"status":"pass","version":"v24.11.0"},` + absent + `,"guidance":["omp","opencode"]}`, "REA: 미결정"},
	}
	for _, tc := range cases {
		recorder.commands = nil
		data, checks, _ := reaDoctorJSON(t, tc.dir)
		assert.JSONEq(t, tc.data, data, tc.name)
		for id, check := range checks {
			assert.Equal(t, "info", check.Severity, tc.name+" "+id)
		}

		text, _ := runReaCLI(t, "", "doctor", "--dir", tc.dir)
		assert.Equal(t, 1, strings.Count(text, tc.line), tc.name+"\n"+text)
		assert.Equal(t, strings.Contains(tc.data, `"omp"`), strings.Contains(text, reaTextKo.Guidance), tc.name)

		require.Len(t, recorder.commands, 2, tc.name+": one probe per doctor run")
		for _, command := range recorder.commands {
			assert.Equal(t, []string{"/bin/bash", "-l", "-c", "node --version"}, command, tc.name)
		}
	}
}

// S8 (d): a stored on decision with a user-owned Claude entry, an other-pin
// Codex table, no Antigravity entry, and node v22.18.0 warns on the Node row
// only, and doctor writes nothing.
func TestDoctorRea_OnWithUnsupportedNode(t *testing.T) {
	reaCLIEnv(t, "v22.18.0\n")
	dir, cfg := reaFixture(t, reaP5...)
	cfg.Features.Rea.Enabled = reaBool(true)
	require.NoError(t, config.Save(dir, cfg))
	writeReaTarget(t, dir, reamcp.Claude, []byte(reaC3))
	writeReaTarget(t, dir, reamcp.Codex, []byte(reaX5))
	before := reaSums(t, dir)

	data, checks, status := reaDoctorJSON(t, dir)

	assert.JSONEq(t, `{"decision":"on","decision_source":"features.rea","node":{"status":"unsupported","version":"v22.18.0"},`+
		`"platforms":{"antigravity-cli":"absent","claude-code":"user_owned","codex":"other_pin"},"guidance":["omp","opencode"]}`, data)
	assert.Equal(t, jsonCheck{ID: "doctor.rea.node", Severity: "warning", Status: "warn", Detail: "unsupported v22.18.0"}, checks["doctor.rea.node"])
	assert.Equal(t, jsonCheck{ID: "doctor.rea.decision", Severity: "info", Status: "pass", Detail: "on (features.rea)"}, checks["doctor.rea.decision"])
	assert.Equal(t, "info", checks["doctor.rea.platform.antigravity-cli"].Severity, "the gate did not pass, so absent is no warning")
	assert.Equal(t, "npx -y rea-agents@6.5.0 setup", checks["doctor.rea.guidance.omp"].Detail)
	assert.Equal(t, "warn", status)

	text, _ := runReaCLI(t, "", "doctor", "--dir", dir)
	assert.Contains(t, text, "REA: 켜짐 (autopus.yaml)")
	assert.Equal(t, before, reaSums(t, dir))
}

// The Doctor Contract rows: decision details, the Node warning rule, the
// platform warning rule, and the guidance rows.
func TestReaDoctorChecks_Rows(t *testing.T) {
	t.Parallel()
	payload := doctorReaPayload{
		Decision: "on", DecisionSource: "managed_entry",
		Node:      doctorReaNode{Status: "pass", Version: "v24.11.0"},
		Platforms: map[string]string{"claude-code": "absent", "codex": "managed"},
		Guidance:  []string{"opencode"},
	}
	assert.Equal(t, []reaDoctorCheck{
		{id: "doctor.rea.decision", detail: "on (managed_entry)"},
		{id: "doctor.rea.node", detail: "pass v24.11.0"},
		{id: "doctor.rea.platform.claude-code", detail: "absent", warning: true},
		{id: "doctor.rea.platform.codex", detail: "managed"},
		{id: "doctor.rea.guidance.opencode", detail: "npx -y rea-agents@6.5.0 setup"},
	}, reaDoctorChecks(payload))

	payload.Decision, payload.Node = "undecided", doctorReaNode{Status: "absent"}
	rows := reaDoctorChecks(payload)
	assert.Equal(t, "undecided", rows[0].detail)
	assert.Equal(t, reaDoctorCheck{id: "doctor.rea.node", detail: "absent"}, rows[1])
	assert.False(t, rows[2].warning, "an undecided project never warns")
	assert.Equal(t, "REA: undecided", reaDoctorDecisionLine(reaTextEn, payload))
}
