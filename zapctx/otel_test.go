package zapctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func Test_OtelTraceInjectNoSpan(t *testing.T) {
	ctx := ToContext(context.Background(), zap.NewNop())
	out := OtelTraceInject(ctx)
	assert.Equal(t, ctx, out)
}

func Test_OtelTraceInjectNoZapContext(t *testing.T) {
	// Should not panic if ZapContext is missing.
	out := OtelTraceInject(context.Background())
	assert.NotNil(t, out)
}

func Test_OtelTraceInjectWithSpan(t *testing.T) {
	tid, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	assert.NoError(t, err)
	sid, err := trace.SpanIDFromHex("0102030405060708")
	assert.NoError(t, err)
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := ToContext(context.Background(), zap.NewNop())
	ctx = trace.ContextWithSpanContext(ctx, sc)
	OtelTraceInject(ctx)

	z := Extract(ctx)
	assert.True(t, z.Sampled)
	assert.Len(t, z.fields, 3)
}
