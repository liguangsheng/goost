package zapctx

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCPayloadOption configures the gRPC payload-logging interceptor.
type GRPCPayloadOption func(*grpcPayloadConfig)

type grpcPayloadConfig struct {
	logBody     bool
	sampleEvery int64
	skip        func(method string) bool
}

// GRPCWithBody toggles logging of the request and response messages.
// Bodies are formatted by fmt.Sprintf("%+v", msg).
func GRPCWithBody(on bool) GRPCPayloadOption {
	return func(c *grpcPayloadConfig) { c.logBody = on }
}

// GRPCWithSampling logs every n-th RPC. Defaults to 1.
func GRPCWithSampling(n int) GRPCPayloadOption {
	return func(c *grpcPayloadConfig) {
		if n < 1 {
			n = 1
		}
		c.sampleEvery = int64(n)
	}
}

// GRPCWithSkipper skips RPCs whose full method (e.g. "/foo.Svc/Ping")
// matches the predicate.
func GRPCWithSkipper(fn func(method string) bool) GRPCPayloadOption {
	return func(c *grpcPayloadConfig) { c.skip = fn }
}

// PayloadUnaryServerInterceptor logs RPC method, gRPC status code, latency
// and (optionally) the request/response messages, using the context-bound
// logger if one was attached upstream by UnaryServerInterceptor.
func PayloadUnaryServerInterceptor(logger *zap.Logger, opts ...GRPCPayloadOption) grpc.UnaryServerInterceptor {
	cfg := &grpcPayloadConfig{sampleEvery: 1}
	for _, o := range opts {
		o(cfg)
	}
	var counter atomic.Int64

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if cfg.skip != nil && cfg.skip(info.FullMethod) {
			return handler(ctx, req)
		}
		if cfg.sampleEvery > 1 && counter.Add(1)%cfg.sampleEvery != 0 {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err).String()

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("code", code),
			zap.Duration("latency", time.Since(start)),
		}
		if cfg.logBody {
			fields = append(fields,
				zap.Stringer("request", stringerFunc{v: req}),
				zap.Stringer("response", stringerFunc{v: resp}),
			)
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
		}
		L(ctx).With(fields...).Info("grpc")
		_ = logger
		return resp, err
	}
}

// stringerFunc is a lazy Stringer that fmt-formats the wrapped value.
// Avoids allocating "%+v" output for skipped log levels.
type stringerFunc struct{ v any }

func (s stringerFunc) String() string {
	if s.v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", s.v)
}
