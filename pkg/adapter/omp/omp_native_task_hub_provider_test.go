package omp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

type ompNativeChatRequest struct {
	Model    string            `json:"model"`
	Stream   bool              `json:"stream"`
	Tools    []json.RawMessage `json:"tools"`
	Messages []json.RawMessage `json:"messages"`
}

type ompNativeProviderReceipt struct {
	Requests            int
	ParentStages        int
	ChildRequests       map[string]int
	AuthHeaders         int
	CredentialLeaks     int
	UnexpectedEndpoints int
	Failure             string
}

type ompNativeProvider struct {
	server *httptest.Server

	mu                                sync.Mutex
	parentStage                       int
	childRequests                     map[string]int
	requestCount                      int
	authHeaders                       int
	credentialLeaks                   int
	unexpectedEndpoints               int
	failure                           string
	childrenReady                     chan struct{}
	childrenReadyOnce                 sync.Once
	alphaRelease, betaRelease         chan struct{}
	alphaReleaseOnce, betaReleaseOnce sync.Once
}

func newOMPNativeProvider(t *testing.T) *ompNativeProvider {
	t.Helper()
	provider := &ompNativeProvider{
		childRequests: map[string]int{ompNativeAlphaID: 0, ompNativeBetaID: 0},
		childrenReady: make(chan struct{}),
		alphaRelease:  make(chan struct{}),
		betaRelease:   make(chan struct{}),
	}
	provider.server = httptest.NewServer(http.HandlerFunc(provider.handle))
	t.Cleanup(provider.server.Close)
	return provider
}

func (p *ompNativeProvider) handle(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.requestCount++
	p.mu.Unlock()
	if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
		p.mu.Lock()
		p.unexpectedEndpoints++
		p.mu.Unlock()
		p.reject(w, "unexpected_endpoint")
		return
	}
	if ompNativeRequestHasAuth(r.Header) {
		p.mu.Lock()
		p.authHeaders++
		p.mu.Unlock()
		p.reject(w, "unexpected_auth_header")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
	if err != nil {
		p.reject(w, "request_oversized")
		return
	}
	if strings.Contains(string(body), ompNativeCredential) {
		p.mu.Lock()
		p.credentialLeaks++
		p.mu.Unlock()
		p.reject(w, "credential_leak")
		return
	}
	var request ompNativeChatRequest
	if json.Unmarshal(body, &request) != nil || request.Model != ompLiveModel || !request.Stream {
		p.reject(w, "invalid_request_shape")
		return
	}
	// The parent conversation is the one carrying the smoke prompt; a child
	// only ever receives the shared context plus its own assignment. Tool
	// shape is no longer a discriminator: the bundled `reviewer` declares
	// `spawns: [scout]`, so its request carries a `task` tool too, and
	// "has task tool ⇒ parent" misrouted that child into the parent staging
	// sequence. That rule held only while the project shipped its own
	// non-spawning agent definitions.
	if strings.Contains(string(body), ompNativeSmokeToken) {
		p.handleParent(w, r, request, body)
		return
	}
	p.handleChild(w, r, request, body)
}

// handleParent scripts the omp 18.3+ lifecycle: one task batch, then one
// `wait` per child. Each wait returns the first settled job as a
// `## Completed` block carrying the child's final message, with the other
// child under `## Still Running` while it lasts; the matching child is
// released when the wait's tool execution starts, so the order is fixed.
func (p *ompNativeProvider) handleParent(
	w http.ResponseWriter,
	r *http.Request,
	request ompNativeChatRequest,
	body []byte,
) {
	p.mu.Lock()
	stage := p.parentStage
	p.mu.Unlock()
	lastTool := ompNativeLastToolContent(request.Messages)
	var validationErr error
	switch stage {
	case 0:
		if !strings.Contains(string(body), ompNativeSmokeToken) {
			validationErr = fmt.Errorf("smoke prompt missing")
		} else {
			validationErr = validateOMPNativeParentTools(request.Tools)
		}
	case 1:
		validationErr = ompNativeRequireTokens(lastTool,
			"Spawned 2 background agents using scout, reviewer",
			"`"+ompNativeAlphaID+"` (job `"+ompNativeAlphaID+"`)",
			"`"+ompNativeBetaID+"` (job `"+ompNativeBetaID+"`)")
		if validationErr == nil {
			select {
			case <-p.childrenReady:
			case <-r.Context().Done():
				validationErr = fmt.Errorf("children did not reach provider")
			case <-time.After(8 * time.Second):
				validationErr = fmt.Errorf("children did not start")
			}
		}
	case 2:
		validationErr = ompNativeRequireTokens(lastTool,
			"## Completed (1)", "## Still Running (1)", "`"+ompNativeBetaID+"`")
		if validationErr == nil {
			validationErr = ompNativeRequireOutputReceipt(lastTool, ompNativeAlphaID, "scout")
		}
	case 3:
		validationErr = ompNativeRequireTokens(lastTool, "## Completed (1)")
		if validationErr == nil && strings.Contains(lastTool, "Still Running") {
			validationErr = fmt.Errorf("a child is still running after the last wait")
		}
		if validationErr == nil {
			validationErr = ompNativeRequireOutputReceipt(lastTool, ompNativeBetaID, "reviewer")
		}
	default:
		validationErr = fmt.Errorf("parent request budget exceeded")
	}
	if validationErr != nil {
		p.reject(w, fmt.Sprintf("parent_stage_%d:%v", stage, validationErr))
		return
	}
	p.mu.Lock()
	if p.parentStage != stage {
		p.mu.Unlock()
		p.reject(w, "parent_stage_race")
		return
	}
	p.parentStage++
	p.mu.Unlock()

	switch stage {
	case 0:
		writeOMPNativeToolSSE(w, 1, "task", ompNativeTaskArguments())
	case 1:
		writeOMPNativeToolSSE(w, 2, "wait", map[string]any{"i": "Awaiting alpha receipt"})
	case 2:
		writeOMPNativeToolSSE(w, 3, "wait", map[string]any{"i": "Awaiting beta receipt"})
	case 3:
		writeOMPNativeFinalSSE(w, 4, "native task lifecycle complete")
	}
}

// handleChild holds each child until the parent waits for it, then ends the
// child's turn on the receipt as plain text: omp 18.3+ children have no
// `yield`, their final message is the result the parent receives.
func (p *ompNativeProvider) handleChild(
	w http.ResponseWriter,
	r *http.Request,
	request ompNativeChatRequest,
	body []byte,
) {
	bodyText := string(body)
	id := ""
	if strings.Contains(bodyText, "NATIVE_CHILD_ALPHA") {
		id = ompNativeAlphaID
	}
	if strings.Contains(bodyText, "NATIVE_CHILD_BETA") {
		if id != "" {
			p.reject(w, "child_assignment_ambiguous")
			return
		}
		id = ompNativeBetaID
	}
	if id == "" {
		p.reject(w, "unknown_child_assignment")
		return
	}
	if err := validateOMPNativeChildTools(id, request.Tools); err != nil {
		p.reject(w, "child_schema:"+err.Error())
		return
	}
	p.mu.Lock()
	p.childRequests[id]++
	count := p.childRequests[id]
	ready := p.childRequests[ompNativeAlphaID] == 1 && p.childRequests[ompNativeBetaID] == 1
	p.mu.Unlock()
	if count != 1 {
		p.reject(w, "child_request_budget_exceeded:"+id)
		return
	}
	if ready {
		p.childrenReadyOnce.Do(func() { close(p.childrenReady) })
	}
	release := p.alphaRelease
	if id == ompNativeBetaID {
		release = p.betaRelease
	}
	select {
	case <-release:
	case <-r.Context().Done():
		p.reject(w, "child_request_cancelled:"+id)
		return
	case <-time.After(30 * time.Second):
		p.reject(w, "child_release_timeout:"+id)
		return
	}
	receipt, err := json.Marshal(ompNativeExpectedReceipt(id))
	if err != nil {
		p.reject(w, "child_receipt_encode:"+id)
		return
	}
	writeOMPNativeFinalSSE(w, 100+count, string(receipt))
}

func (p *ompNativeProvider) release(id string) {
	if id == ompNativeAlphaID {
		p.alphaReleaseOnce.Do(func() { close(p.alphaRelease) })
		return
	}
	p.betaReleaseOnce.Do(func() { close(p.betaRelease) })
}

func (p *ompNativeProvider) releaseAll() {
	p.release(ompNativeAlphaID)
	p.release(ompNativeBetaID)
}

func (p *ompNativeProvider) receipt() ompNativeProviderReceipt {
	p.mu.Lock()
	defer p.mu.Unlock()
	children := make(map[string]int, len(p.childRequests))
	for id, count := range p.childRequests {
		children[id] = count
	}
	return ompNativeProviderReceipt{
		Requests: p.requestCount, ParentStages: p.parentStage, ChildRequests: children,
		AuthHeaders: p.authHeaders, CredentialLeaks: p.credentialLeaks,
		UnexpectedEndpoints: p.unexpectedEndpoints, Failure: p.failure,
	}
}

func (p *ompNativeProvider) reject(w http.ResponseWriter, reason string) {
	p.mu.Lock()
	if p.failure == "" {
		p.failure = reason
	}
	p.mu.Unlock()
	p.releaseAll()
	http.Error(w, "native fixture rejected request", http.StatusBadRequest)
}
