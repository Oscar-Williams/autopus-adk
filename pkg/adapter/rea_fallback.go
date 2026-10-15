package adapter

import "context"

// ReaFallback is one REA analysis-rung route a platform generation left on
// the session model because the platform's model catalog cannot serve the
// rung (SPEC-REA-001 Gate 1 and its Codex scope expansion). Adapters report
// the fact; the CLI prints it in the run's language, so no adapter writes an
// English-only line for it.
type ReaFallback struct {
	// Catalog names the catalog that lacked the candidate: "Codex" or "OMP".
	Catalog string
	Agent   string
	// Requested is the candidate as that catalog names it: model/effort for
	// Codex, the provider/model selector for OMP.
	Requested string
}

type reaFallbackSinkKey struct{}

// WithReaFallbackSink returns a context whose generations hand every REA
// fallback they take to sink.
func WithReaFallbackSink(ctx context.Context, sink func(ReaFallback)) context.Context {
	return context.WithValue(ctx, reaFallbackSinkKey{}, sink)
}

// ReportReaFallbacks hands facts to the sink ctx carries. A context without a
// sink, such as a generation outside a CLI command, drops them.
func ReportReaFallbacks(ctx context.Context, facts ...ReaFallback) {
	if ctx == nil || len(facts) == 0 {
		return
	}
	sink, _ := ctx.Value(reaFallbackSinkKey{}).(func(ReaFallback))
	if sink == nil {
		return
	}
	for _, fact := range facts {
		sink(fact)
	}
}
