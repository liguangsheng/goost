// Package slogctxotel provides OpenTelemetry hooks for slogctx.
package slogctxotel

import (
	"context"
	"log/slog"

	"github.com/liguangsheng/goost/slogctx"
	"go.opentelemetry.io/otel/trace"
)

// TraceInject adds OpenTelemetry trace IDs to the context-bound logger
// and mirrors the span's sampled flag onto SlogContext.Sampled.
// Safe to use as a HookFunc; a no-op when ctx has no SlogContext or no
// valid span context.
func TraceInject(ctx context.Context) context.Context {
	s := slogctx.Extract(ctx)
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

// OtelTraceInject is kept as a migration alias for TraceInject.
//
// Deprecated: use TraceInject.
func OtelTraceInject(ctx context.Context) context.Context {
	return TraceInject(ctx)
}
