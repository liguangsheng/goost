package zapctx

import (
	"context"

	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

// OpenTraceInject adds OpenCensus trace IDs to the context-bound logger.
// Safe to use as a HookFunc; a no-op when ctx has no ZapContext or no span.
//
// Deprecated: OpenCensus is no longer maintained. New code should use the
// OpenTelemetry equivalent provided by go.opentelemetry.io packages.
func OpenTraceInject(ctx context.Context) context.Context {
	zapCtx := Extract(ctx)
	if zapCtx == nil {
		return ctx
	}
	span := trace.FromContext(ctx)
	if span == nil {
		return ctx
	}
	spanCtx := span.SpanContext()
	zapCtx.Sampled = spanCtx.IsSampled()
	zapCtx.AddFields(
		zap.String("trace.traceid", spanCtx.TraceID.String()),
		zap.String("trace.spanid", spanCtx.SpanID.String()),
		zap.Bool("trace.sampled", spanCtx.IsSampled()),
	)
	return ctx
}
