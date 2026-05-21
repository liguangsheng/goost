package zapctxgrpc

import (
	"context"
	"fmt"

	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ExampleUnaryServerInterceptor() {
	interceptor := UnaryServerInterceptor(zap.NewNop())
	_, _ = interceptor(
		context.Background(),
		"ping",
		&grpc.UnaryServerInfo{FullMethod: "/example.Service/Ping"},
		func(ctx context.Context, req any) (any, error) {
			fmt.Println(zapctx.Extract(ctx) != nil)
			return "pong", nil
		},
	)

	// Output:
	// true
}
