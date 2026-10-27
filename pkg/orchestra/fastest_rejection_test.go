package orchestra

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runFastest used to drop every per-provider error and report only its
// summary line, so a fork/exec failure in CI surfaced as nothing more than
// "모든 프로바이더가 응답하지 않았습니다". The aggregated error must name each
// rejected provider together with the cause.
func TestRunFastest_AllRejected_ErrorNamesEachProviderCause(t *testing.T) {
	tests := []struct {
		name   string
		build  func() command
		expect string
	}{
		{
			name: "start failure keeps the exec error",
			build: func() command {
				return &fakeCommand{waitCh: make(chan error, 1), startFn: func(*fakeCommand) error {
					return errors.New("fork/exec /tmp/recorder: text file busy")
				}}
			},
			expect: "recorder: recorder 시작 실패: fork/exec /tmp/recorder: text file busy",
		},
		{
			name: "exit-0 empty stdout is named as empty output",
			build: func() command {
				fake := &fakeCommand{waitCh: make(chan error, 1)}
				fake.waitCh <- nil
				return fake
			},
			expect: "recorder: empty output",
		},
	}
	binary, err := os.Executable()
	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubSingleProviderCommands(t, func(string, []string) command { return tt.build() })

			_, err := RunOrchestra(context.Background(), OrchestraConfig{
				Providers:      []ProviderConfig{{Name: "recorder", Binary: binary}},
				Strategy:       StrategyFastest,
				Prompt:         "where are you",
				TimeoutSeconds: 5,
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "모든 프로바이더가 응답하지 않았습니다")
			assert.Contains(t, err.Error(), tt.expect)
		})
	}
}
