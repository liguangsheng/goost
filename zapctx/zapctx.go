package zapctx

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapContext struct {
	logger  *zap.Logger
	fields  []zapcore.Field
	Sampled bool
}

func (c *zapContext) Logger() *zap.Logger {
	return c.logger.With(c.fields...)
}

func (c *zapContext) AddFields(fields ...zapcore.Field) {
	c.fields = append(c.fields, fields...)
}

type ctxMarker struct{}

var (
	ctxMarkerKey = ctxMarker{}
	nopLogger    = zap.NewNop()
)

func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, &zapContext{logger: logger})
}

func Extract(ctx context.Context) *zapContext {
	if ctx == nil {
		return nil
	}
	if _ctx, ok := ctx.Value(ctxMarkerKey).(*zapContext); ok {
		return _ctx
	}
	return nil
}

func L(ctx context.Context) *zap.Logger {
	_ctx := Extract(ctx)
	if _ctx != nil {
		return _ctx.Logger()
	}
	return zap.L()
}

func S(ctx context.Context) *zap.SugaredLogger {
	return L(ctx).Sugar()
}

func Sampled(ctx context.Context) *zap.Logger {
	_ctx := Extract(ctx)
	if _ctx != nil && _ctx.Sampled {
		return _ctx.Logger()
	}
	return nopLogger
}
