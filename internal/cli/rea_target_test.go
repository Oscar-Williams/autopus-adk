package cli

// SPEC-REA-001 security audit L5: a REA target file is read without
// following a symlink, only when it is a regular file, and at most 1 MiB, so
// a committed symlink to a FIFO or /dev/zero can neither hang nor exhaust a
// run. The off sweep reports what it could not inspect or change.

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/reamcp"
)

func TestReadReaTarget_ReadsRegularFilesUpToTheCap(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	data, err := readReaTarget(dir, reamcp.Claude)
	require.NoError(t, err)
	assert.Nil(t, data, "a missing file is nil data")

	full := bytes.Repeat([]byte(" "), reaTargetLimit)
	writeReaTarget(t, dir, reamcp.Claude, full)
	data, err = readReaTarget(dir, reamcp.Claude)
	require.NoError(t, err)
	assert.Len(t, data, reaTargetLimit)
}

// A needed removal whose result fails the Codex structure check (rea-mcp
// security L4) is refused: one line names the file, nothing is written, the
// off decision is still stored, and the run fails.
func TestReaRunFinish_SweepReportsARefusedRemoval(t *testing.T) {
	t.Parallel()
	dir, cfg := reaFixture(t, "claude-code")
	codex := "approval_policy\n\n[mcp_servers.rea]\ncommand = \"npx\"\nargs = [\"-y\", \"rea-agents@6.5.0\", \"mcp\"]\nstartup_timeout_sec = 120\n"
	writeReaTarget(t, dir, reamcp.Codex, []byte(codex))
	require.Equal(t, reamcp.ClassManaged, reaFileClass(dir, reamcp.Codex))
	run, _ := testReaRun(t, dir, cfg, "v24.11.0\n")
	var out bytes.Buffer
	run.decide(context.Background(), &out, cfg, reaOff)

	err := run.finish(&out, dir, cfg)

	require.EqualError(t, err, "REA MCP 항목을 바꾸지 못한 대상 파일이 있습니다")
	assert.Equal(t, 1, strings.Count(out.String(),
		"  REA MCP 항목을 바꾸지 않았습니다: .codex/config.toml (바꾼 결과가 설정 구조 검사를 통과하지 못합니다)\n"), out.String())
	assert.Equal(t, codex, string(readReaFile(t, dir, reamcp.Codex)))
	reloaded, loadErr := config.Load(dir)
	require.NoError(t, loadErr)
	assert.Equal(t, reaBool(false), reloaded.Features.Rea.Enabled)
}
