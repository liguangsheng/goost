package slogctxotel

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/liguangsheng/goost/slogctx"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func newCapturingLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), buf
}

func Test_TraceInjectWithSpan(t *testing.T) {
	tid, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	sid, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	logger, buf := newCapturingLogger()
	ctx := slogctx.ToContext(context.Background(), logger)
	ctx = trace.ContextWithSpanContext(ctx, sc)
	TraceInject(ctx)

	slogctx.Sampled(ctx).Info("yes")
	out := buf.String()
	assert.Contains(t, out, "trace.traceid=0102030405060708090a0b0c0d0e0f10")
	assert.Contains(t, out, "yes")
}

func Test_TraceInjectNoSlogContext(t *testing.T) {
	out := TraceInject(context.Background())
	assert.NotNil(t, out)
}

func Test_OtelTraceInjectAlias(t *testing.T) {
	out := OtelTraceInject(context.Background())
	assert.NotNil(t, out)
}
