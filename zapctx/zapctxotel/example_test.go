package zapctxotel

import (
	"context"
	"fmt"

	"github.com/liguangsheng/goost/zapctx"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func ExampleTraceInject() {
	tid, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	sid, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	core, logs := observer.New(zapcore.InfoLevel)
	ctx := zapctx.ToContext(context.Background(), zap.New(core))
	ctx = trace.ContextWithSpanContext(ctx, sc)

	TraceInject(ctx)
	zapctx.Sampled(ctx).Info("sampled")

	entry := logs.All()[0]
	fmt.Println(entry.ContextMap()["trace.traceid"])
	fmt.Println(entry.ContextMap()["trace.sampled"])

	// Output:
	// 0102030405060708090a0b0c0d0e0f10
	// true
}
