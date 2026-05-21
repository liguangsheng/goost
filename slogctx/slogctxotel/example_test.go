package slogctxotel

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/liguangsheng/goost/slogctx"
	"go.opentelemetry.io/otel/trace"
)

func ExampleTraceInject() {
	tid, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	sid, _ := trace.SpanIDFromHex("0102030405060708")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
	})

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(buf, nil))
	ctx := slogctx.ToContext(context.Background(), logger)
	ctx = trace.ContextWithSpanContext(ctx, sc)

	TraceInject(ctx)
	slogctx.Sampled(ctx).Info("sampled")

	out := buf.String()
	fmt.Println(strings.Contains(out, "trace.traceid=0102030405060708090a0b0c0d0e0f10"))
	fmt.Println(strings.Contains(out, "trace.sampled=true"))

	// Output:
	// true
	// true
}
