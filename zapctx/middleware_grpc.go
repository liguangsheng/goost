package zapctx

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor(logger *zap.Logger, hooks ...HookFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		newCtx := ToContext(ctx, logger)
		for _, hook := range hooks {
			newCtx = hook(newCtx)
		}
		return handler(newCtx, req)
	}
}
