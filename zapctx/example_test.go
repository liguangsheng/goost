package zapctx_test

import (
	"context"
	"fmt"

	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func ExampleToContext() {
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	ctx := zapctx.ToContext(context.Background(), logger)
	zapctx.Extract(ctx).AddFields(zap.String("request_id", "abc123"))

	zapctx.L(ctx).Info("handled")

	entry := logs.All()[0]
	fmt.Println(entry.Message)
	fmt.Println(entry.ContextMap()["request_id"])

	// Output:
	// handled
	// abc123
}
