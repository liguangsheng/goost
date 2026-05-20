// Package zapctx carries a *zap.Logger and structured fields through a
// context.Context so handlers can log with request-scoped fields without
// threading the logger through every function signature.
package zapctx

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// HookFunc enriches a context with additional fields before it is handed
// to a handler. Middleware in this package calls hooks in order.
type HookFunc func(ctx context.Context) context.Context

// ZapContext holds the logger and accumulated fields associated with a
// context.Context.
type ZapContext struct {
	logger  *zap.Logger
	fields  []zapcore.Field
	Sampled bool
}

// Logger returns a logger with all accumulated fields applied.
func (c *ZapContext) Logger() *zap.Logger {
	return c.logger.With(c.fields...)
}

// AddFields appends fields to be included with every future log call.
func (c *ZapContext) AddFields(fields ...zapcore.Field) {
	c.fields = append(c.fields, fields...)
}

type ctxMarker struct{}

var (
	ctxMarkerKey = ctxMarker{}
	nopLogger    = zap.NewNop()
)

// ToContext attaches logger to ctx so it can be retrieved with L/S/Extract.
func ToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, &ZapContext{logger: logger})
}

// Extract returns the ZapContext stored on ctx, or nil if absent.
func Extract(ctx context.Context) *ZapContext {
	if ctx == nil {
		return nil
	}
	if z, ok := ctx.Value(ctxMarkerKey).(*ZapContext); ok {
		return z
	}
	return nil
}

// L returns the context-bound logger or the zap global if none is attached.
func L(ctx context.Context) *zap.Logger {
	if z := Extract(ctx); z != nil {
		return z.Logger()
	}
	return zap.L()
}

// S is the sugared variant of L.
func S(ctx context.Context) *zap.SugaredLogger {
	return L(ctx).Sugar()
}

// Sampled returns the context-bound logger only when the request is sampled
// for tracing; otherwise a no-op logger is returned. Useful for verbose
// per-request debug logs.
func Sampled(ctx context.Context) *zap.Logger {
	if z := Extract(ctx); z != nil && z.Sampled {
		return z.Logger()
	}
	return nopLogger
}
