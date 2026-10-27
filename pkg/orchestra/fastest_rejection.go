package orchestra

import (
	"fmt"
	"strings"
	"sync"
)

// fastestRejections remembers why runFastest rejected each provider so the
// all-failed error names the causes instead of only the summary line. Before
// this, a fork/exec failure of the only provider was reported as nothing more
// than "모든 프로바이더가 응답하지 않았습니다".
type fastestRejections struct {
	mu      sync.Mutex
	reasons []string
}

// record stores one rejection; provider goroutines call it concurrently.
func (r *fastestRejections) record(provider ProviderConfig, resp *ProviderResponse, err error) {
	reason := fastestRejectionReason(resp, err)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reasons = append(r.reasons, provider.Name+": "+reason)
}

// err builds the all-failed error once every provider has been rejected.
func (r *fastestRejections) err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.reasons) == 0 {
		return fmt.Errorf("모든 프로바이더가 응답하지 않았습니다")
	}
	return fmt.Errorf("모든 프로바이더가 응답하지 않았습니다: %s", strings.Join(r.reasons, "; "))
}

// fastestRejectionReason mirrors the rejection order of runFastest: a runner
// error wins, then a missing response, then the timeout and empty-output
// rejections that runParallel also applies.
func fastestRejectionReason(resp *ProviderResponse, err error) string {
	switch {
	case err != nil:
		return redactFailureText(err.Error())
	case resp == nil:
		return "no response"
	case resp.TimedOut:
		return "timeout"
	case resp.EmptyOutput:
		return "empty output"
	default:
		return "rejected"
	}
}
