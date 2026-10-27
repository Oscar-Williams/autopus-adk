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

// reaFallbackNotice prints each distinct fallback of one command once,
// however many generations (preview and apply) report it, and not at all when
// a composition table of the run already showed it (claim). It goes to stderr
// so JSON on stdout stays parseable.
type reaFallbackNotice struct {
	mu   sync.Mutex
	out  io.Writer
	lang string
	seen map[adapter.ReaFallback]bool
}

func (n *reaFallbackNotice) report(fact adapter.ReaFallback) {
	if n.claim(fact) {
		_, _ = fmt.Fprintln(n.out, "  "+reaFallbackLine(fact, n.lang))
	}
}

// claim marks fact as printed for the run, the same platform, agent, and
// candidate key report dedupes on, and reports whether it was new.
func (n *reaFallbackNotice) claim(fact adapter.ReaFallback) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.seen[fact] {
		return false
	}
	n.seen[fact] = true
	return true
}

// fresh keeps the facts the run has not printed yet and claims them; a table
// that prints their lines calls it so no later generation repeats them.
func (n *reaFallbackNotice) fresh(facts []adapter.ReaFallback) []adapter.ReaFallback {
	var kept []adapter.ReaFallback
	for _, fact := range facts {
		if n.claim(fact) {
			kept = append(kept, fact)
		}
	}
	return kept
}

type reaFallbackNoticeKey struct{}

// withReaFallbackNotice returns ctx carrying the fallback notice of cmd, both
// as the adapters' sink and for the run's composition tables.
func withReaFallbackNotice(ctx context.Context, cmd *cobra.Command) context.Context {
	notice := &reaFallbackNotice{out: cmd.ErrOrStderr(), lang: composeLanguage(), seen: map[adapter.ReaFallback]bool{}}
	ctx = context.WithValue(ctx, reaFallbackNoticeKey{}, notice)
	return adapter.WithReaFallbackSink(ctx, notice.report)
}

// reaFallbackNoticeFrom returns the notice ctx carries, or nil outside a
// command run.
func reaFallbackNoticeFrom(ctx context.Context) *reaFallbackNotice {
	if ctx == nil {
		return nil
	}
	notice, _ := ctx.Value(reaFallbackNoticeKey{}).(*reaFallbackNotice)
	return notice
}
