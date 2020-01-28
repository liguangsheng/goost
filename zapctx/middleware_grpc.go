package zapctx

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type HookFunc func(ctx context.Context) context.Context

func UnaryServerInterceptor(logger *zap.Logger, hooks ...HookFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp interface{}, err error) {
		newCtx := ToContext(ctx, logger)

		for _, hook := range hooks {
			newCtx = hook(newCtx)
		}

		resp, err = handler(newCtx, req)
		return resp, err
	}
}
