package cli

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/spf13/cobra"

	"github.com/insajin/autopus-adk/pkg/adapter"
)

// reaFallbackLine is the Form Text line of one REA analysis-rung fallback
// (SPEC-REA-001): Korean for "ko", English otherwise. The OMP wording is the
// SPEC's; the Codex one names its catalog the same way.
func reaFallbackLine(fact adapter.ReaFallback, lang string) string {
	agent, requested := terminalSafe(fact.Agent), terminalSafe(fact.Requested)
	switch {
	case fact.Catalog == "OMP" && lang == "ko":
		return fmt.Sprintf("%s: %s이 OMP catalog에 없어 세션 모델을 씁니다", agent, requested)
	case fact.Catalog == "OMP":
		return fmt.Sprintf("%s: %s not in catalog → session model", agent, requested)
	case lang == "ko":
		return fmt.Sprintf("%s: %s이 %s catalog에 없어 세션 모델을 씁니다", agent, requested, terminalSafe(fact.Catalog))
	default:
		return fmt.Sprintf("%s: %s not in %s catalog → session model", agent, requested, terminalSafe(fact.Catalog))
	}
}

// reaFallbackNotice prints each distinct fallback a command's generations
// report once, however many generations (preview and apply) take it. It goes
// to stderr so JSON on stdout stays parseable.
type reaFallbackNotice struct {
	mu   sync.Mutex
	out  io.Writer
	lang string
	seen map[adapter.ReaFallback]bool
}

func (n *reaFallbackNotice) report(fact adapter.ReaFallback) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.seen[fact] {
		return
	}
	n.seen[fact] = true
	_, _ = fmt.Fprintln(n.out, "  "+reaFallbackLine(fact, n.lang))
}

// withReaFallbackNotice returns ctx carrying the fallback notice of cmd.
func withReaFallbackNotice(ctx context.Context, cmd *cobra.Command) context.Context {
	notice := &reaFallbackNotice{out: cmd.ErrOrStderr(), lang: composeLanguage(), seen: map[adapter.ReaFallback]bool{}}
	return adapter.WithReaFallbackSink(ctx, notice.report)
}
