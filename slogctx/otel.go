package slogctx

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// OtelTraceInject adds OpenTelemetry trace IDs to the context-bound logger
// and mirrors the span's sampled flag onto SlogContext.Sampled.
// Safe to use as a HookFunc; a no-op when ctx has no SlogContext or no
// valid span context.
func OtelTraceInject(ctx context.Context) context.Context {
	s := Extract(ctx)
	if s == nil {
		return ctx
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ctx
	}
	s.Sampled = sc.IsSampled()
	s.AddAttrs(
		slog.String("trace.traceid", sc.TraceID().String()),
		slog.String("trace.spanid", sc.SpanID().String()),
		slog.Bool("trace.sampled", sc.IsSampled()),
	)
	return ctx
}
