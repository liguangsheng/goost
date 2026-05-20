package slogctx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func newCapturingLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), buf
}

func Test_ToAndExtract(t *testing.T) {
	logger, _ := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	assert.NotNil(t, Extract(ctx))
}

func Test_ExtractMissing(t *testing.T) {
	var nilCtx context.Context
	assert.Nil(t, Extract(nilCtx))
	assert.Nil(t, Extract(context.Background()))
}

func Test_L_FallbackToDefault(t *testing.T) {
	assert.Equal(t, slog.Default(), L(context.Background()))
}

func Test_LWithAddedAttrs(t *testing.T) {
	logger, buf := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	Extract(ctx).AddAttrs(slog.String("k", "v"))
	L(ctx).Info("hello")
	assert.True(t, strings.Contains(buf.String(), "k=v"))
}

func Test_SampledNoOpWhenNotSampled(t *testing.T) {
	logger, buf := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	Sampled(ctx).Info("nope")
	assert.Empty(t, buf.String())
}

func Test_OtelTraceInjectWithSpan(t *testing.T) {
	tid, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	sid, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	logger, buf := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	ctx = trace.ContextWithSpanContext(ctx, sc)
	OtelTraceInject(ctx)

	Sampled(ctx).Info("yes") // sampled now, must log
	out := buf.String()
	assert.Contains(t, out, "trace.traceid=0102030405060708090a0b0c0d0e0f10")
	assert.Contains(t, out, "yes")
}

func Test_OtelTraceInjectNoZapContext(t *testing.T) {
	out := OtelTraceInject(context.Background())
	assert.NotNil(t, out)
}
