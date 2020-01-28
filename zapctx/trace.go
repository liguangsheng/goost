package zapctx

import (
	"context"

	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

func OpenTraceInject(ctx context.Context) context.Context {
	zapCtx := Extract(ctx)
	if span := trace.FromContext(ctx); span != nil {
		spanCtx := span.SpanContext()
		zapCtx.Sampled = spanCtx.IsSampled()
		zapCtx.AddFields(
			zap.String("trace.traceid", spanCtx.TraceID.String()),
			zap.String("trace.spanid", spanCtx.SpanID.String()),
			zap.Bool("trace.sampled", spanCtx.IsSampled()),
		)
	}
	return ctx
}
