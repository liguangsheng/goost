package zapctx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// OtelTraceInject adds OpenTelemetry trace IDs to the context-bound logger
// and forwards the sample flag to ZapContext.Sampled. Safe to use as a
// HookFunc; a no-op when ctx has no ZapContext or no recording span.
func OtelTraceInject(ctx context.Context) context.Context {
	z := Extract(ctx)
	if z == nil {
		return ctx
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ctx
	}
	z.Sampled = sc.IsSampled()
	z.AddFields(
		zap.String("trace.traceid", sc.TraceID().String()),
		zap.String("trace.spanid", sc.SpanID().String()),
		zap.Bool("trace.sampled", sc.IsSampled()),
	)
	return ctx
}
