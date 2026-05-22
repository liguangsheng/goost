package slogctx_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/liguangsheng/goost/slogctx"
)

func Example() {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	ctx := slogctx.ToContext(context.Background(), logger)
	slogctx.Extract(ctx).AddAttrs(slog.String("request_id", "req-123"))
	slogctx.L(ctx).Info("handled")

	out := buf.String()
	fmt.Println(strings.Contains(out, "msg=handled"))
	fmt.Println(strings.Contains(out, "request_id=req-123"))

	// Output:
	// true
	// true
}
