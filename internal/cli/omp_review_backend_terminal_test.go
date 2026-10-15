package cli

// SPEC-REVMODEL-001 T6 (REQ-RVM-20; acceptance S16(a)): the two stderr lines
// the OMP review backend writes pass the provider, the selector, the tools,
// and the retry reason through terminalSafe, and stay byte for byte the
// same for text without control characters.

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/config"
	"github.com/insajin/autopus-adk/pkg/orchestra"
)

// executeOMPReviewRetry runs one request whose first attempt fails with a
// transient provider error, so both stderr lines are written, and returns
// what the backend wrote to os.Stderr.
func executeOMPReviewRetry(t *testing.T, provider string) string {
	t.Helper()
	original := ompReviewBackoffUnit
	ompReviewBackoffUnit = 10 * time.Millisecond
	t.Cleanup(func() { ompReviewBackoffUnit = original })
	projectDir, _ := configureOMPReviewRPCFixture(t, "provider-error-once")
	var err error
	stderr := captureStderr(t, func() {
		_, err = newOMPReviewBackend(projectDir).Execute(context.Background(), orchestra.ProviderRequest{
			Provider: provider, Role: "reviewer", Prompt: "review", Timeout: 10 * time.Second,
			Config: orchestra.ProviderConfig{Backend: config.ProviderBackendOMP, Model: "openai/model:high", Tools: []string{"read"}},
		})
	})
	require.NoError(t, err)
	return stderr
}

func TestOMPReviewBackend_StderrLinesAreTerminalSafe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX launcher")
	}
	stderr := executeOMPReviewRetry(t, "gem\x1b[2Jini")

	assert.NotContains(t, stderr, "\x1b")
	assert.Contains(t, stderr, `OMP 리뷰 세션: provider=gem\x1b[2Jini model=openai/model:high tools=read`+"\n")
	assert.Contains(t, stderr, `OMP 리뷰 세션 재시도: provider=gem\x1b[2Jini attempt=2/3 reason=`)
}

func TestOMPReviewBackend_StderrLinesKeepPlainText(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable uses a POSIX launcher")
	}
	stderr := executeOMPReviewRetry(t, "reviewer")

	lines := strings.SplitAfter(stderr, "\n")
	require.GreaterOrEqual(t, len(lines), 2, stderr)
	assert.Equal(t, "OMP 리뷰 세션: provider=reviewer model=openai/model:high tools=read\n", lines[0])
	assert.Equal(t, "OMP 리뷰 세션 재시도: provider=reviewer attempt=2/3 reason=OMP pipeline turn failed with provider error status 529: overloaded_error\n", lines[1])
}
