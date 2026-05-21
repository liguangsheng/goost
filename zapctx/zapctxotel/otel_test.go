package zapctxotel

import (
	"context"
	"testing"

	"github.com/liguangsheng/goost/zapctx"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func Test_TraceInjectNoSpan(t *testing.T) {
	ctx := zapctx.ToContext(context.Background(), zap.NewNop())
	out := TraceInject(ctx)
	assert.Equal(t, ctx, out)
}

func Test_TraceInjectNoZapContext(t *testing.T) {
	// Should not panic if ZapContext is missing.
	out := TraceInject(context.Background())
	assert.NotNil(t, out)
}

func Test_TraceInjectWithSpan(t *testing.T) {
	tid, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	assert.NoError(t, err)
	sid, err := trace.SpanIDFromHex("0102030405060708")
	assert.NoError(t, err)
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

	z := zapctx.Extract(ctx)
	assert.True(t, z.Sampled)
	entry := logs.FilterMessage("sampled").All()[0]
	assert.Equal(t, "0102030405060708090a0b0c0d0e0f10", entry.ContextMap()["trace.traceid"])
	assert.Equal(t, "0102030405060708", entry.ContextMap()["trace.spanid"])
	assert.Equal(t, true, entry.ContextMap()["trace.sampled"])
}
