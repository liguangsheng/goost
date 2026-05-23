package zapctxgrpc

import (
	"context"
	"errors"
	"testing"

	"github.com/liguangsheng/goost/zapctx"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_PayloadUnaryServerInterceptor_Success(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	chain := func(ctx context.Context, req any) (any, error) {
		ctx = zapctx.ToContext(ctx, logger)
		return PayloadUnaryServerInterceptor(logger, GRPCWithBody(true))(
			ctx, req,
			&grpc.UnaryServerInfo{FullMethod: "/svc/Ping"},
			func(ctx context.Context, req any) (any, error) { return "pong", nil },
		)
	}
	resp, err := chain(context.Background(), "ping")
	assert.NoError(t, err)
	assert.Equal(t, "pong", resp)

	assert.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.Equal(t, "/svc/Ping", fields["method"])
	assert.Equal(t, "OK", fields["code"])
	assert.Equal(t, "ping", fields["request"])
	assert.Equal(t, "pong", fields["response"])
}

func Test_PayloadUnaryServerInterceptor_MaxBodyTruncatesMessages(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	resp, err := PayloadUnaryServerInterceptor(logger, GRPCWithBody(true), GRPCWithMaxBody(4))(
		zapctx.ToContext(context.Background(), logger), "hello world",
		&grpc.UnaryServerInfo{FullMethod: "/svc/Ping"},
		func(ctx context.Context, req any) (any, error) { return "pong pong", nil },
	)
	assert.NoError(t, err)
	assert.Equal(t, "pong pong", resp)

	if assert.Equal(t, 1, logs.Len()) {
		fields := logs.All()[0].ContextMap()
		assert.Equal(t, "hell", fields["request"])
		assert.Equal(t, "pong", fields["response"])
	}
}

func Test_PayloadUnaryServerInterceptor_MaxBodyZeroDisablesMessages(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	_, err := PayloadUnaryServerInterceptor(logger, GRPCWithBody(true), GRPCWithMaxBody(0))(
		zapctx.ToContext(context.Background(), logger), "secret",
		&grpc.UnaryServerInfo{FullMethod: "/svc/Ping"},
		func(ctx context.Context, req any) (any, error) { return "response", nil },
	)
	assert.NoError(t, err)

	if assert.Equal(t, 1, logs.Len()) {
		fields := logs.All()[0].ContextMap()
		assert.NotContains(t, fields, "request")
		assert.NotContains(t, fields, "response")
	}
}

func Test_PayloadUnaryServerInterceptor_Error(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	want := status.Error(codes.PermissionDenied, "no")
	_, err := PayloadUnaryServerInterceptor(logger)(
		zapctx.ToContext(context.Background(), logger), nil,
		&grpc.UnaryServerInfo{FullMethod: "/svc/Op"},
		func(ctx context.Context, req any) (any, error) { return nil, want },
	)
	assert.ErrorIs(t, err, want)

	fields := logs.All()[0].ContextMap()
	assert.Equal(t, "PermissionDenied", fields["code"])
	assert.Contains(t, fields["error"], "no")
}

func Test_PayloadUnaryServerInterceptor_Sampling(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	icp := PayloadUnaryServerInterceptor(logger, GRPCWithSampling(3))

	for range 9 {
		_, _ = icp(zapctx.ToContext(context.Background(), logger), nil,
			&grpc.UnaryServerInfo{FullMethod: "/svc/M"},
			func(ctx context.Context, req any) (any, error) { return nil, nil },
		)
	}
	assert.Equal(t, 3, logs.Len())
}

func Test_PayloadUnaryServerInterceptor_Skipper(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	icp := PayloadUnaryServerInterceptor(logger, GRPCWithSkipper(func(m string) bool {
		return m == "/svc/Health"
	}))

	for _, m := range []string{"/svc/Health", "/svc/Do"} {
		_, _ = icp(zapctx.ToContext(context.Background(), logger), nil,
			&grpc.UnaryServerInfo{FullMethod: m},
			func(ctx context.Context, req any) (any, error) { return nil, nil },
		)
	}
	assert.Equal(t, 1, logs.Len())
}

// ensure errors.Is path works through gRPC status wrapper
var _ = errors.Is
