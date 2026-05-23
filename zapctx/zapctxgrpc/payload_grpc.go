package zapctxgrpc

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCPayloadOption configures the gRPC payload-logging interceptor.
type GRPCPayloadOption func(*grpcPayloadConfig)

type grpcPayloadConfig struct {
	logBody     bool
	maxBody     int
	sampleEvery int64
	skip        func(method string) bool
}

// GRPCWithBody toggles logging of the request and response messages.
// Bodies are formatted by fmt.Sprintf("%+v", msg).
func GRPCWithBody(on bool) GRPCPayloadOption {
	return func(c *grpcPayloadConfig) { c.logBody = on }
}

// GRPCWithMaxBody caps the number of formatted bytes logged per request and
// response message. 0 means do not log message bodies. Defaults to 4096.
func GRPCWithMaxBody(n int) GRPCPayloadOption {
	return func(c *grpcPayloadConfig) { c.maxBody = n }
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
	cfg := &grpcPayloadConfig{maxBody: 4096, sampleEvery: 1}
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
		if cfg.logBody && cfg.maxBody > 0 {
			fields = append(fields,
				zap.Stringer("request", stringerFunc{v: req, max: cfg.maxBody}),
				zap.Stringer("response", stringerFunc{v: resp, max: cfg.maxBody}),
			)
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
		}
		log := zapctx.L(ctx)
		if zapctx.Extract(ctx) == nil && logger != nil {
			log = logger
		}
		log.With(fields...).Info("grpc")
		return resp, err
	}
}

// stringerFunc is a lazy Stringer that fmt-formats the wrapped value.
// Avoids allocating "%+v" output for skipped log levels.
type stringerFunc struct {
	v   any
	max int
}

func (s stringerFunc) String() string {
	if s.v == nil {
		return "<nil>"
	}
	out := fmt.Sprintf("%+v", s.v)
	if s.max > 0 && len(out) > s.max {
		return out[:s.max]
	}
	return out
}
