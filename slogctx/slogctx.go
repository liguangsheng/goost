// Package slogctx carries a *slog.Logger and accumulated attributes
// through context.Context. It mirrors the API of the zapctx package but
// targets the standard library log/slog.
package slogctx

import (
	"context"
	"io"
	"log/slog"
)

// HookFunc enriches a context with additional attrs before it is passed
// to the next layer.
type HookFunc func(ctx context.Context) context.Context

// SlogContext holds the logger and accumulated attributes associated with
// a context.Context.
type SlogContext struct {
	logger  *slog.Logger
	attrs   []slog.Attr
	Sampled bool
}

// Logger returns a logger with all accumulated attrs applied.
func (c *SlogContext) Logger() *slog.Logger {
	if len(c.attrs) == 0 {
		return c.logger
	}
	anys := make([]any, len(c.attrs))
	for i, a := range c.attrs {
		anys[i] = a
	}
	return c.logger.With(anys...)
}

// AddAttrs appends attrs to be included with every future log call.
func (c *SlogContext) AddAttrs(attrs ...slog.Attr) {
	c.attrs = append(c.attrs, attrs...)
}

type ctxKey struct{}

var nopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// ToContext attaches logger to ctx so it can be retrieved with L/Extract.
func ToContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, &SlogContext{logger: logger})
}

// Extract returns the SlogContext stored on ctx, or nil if absent.
func Extract(ctx context.Context) *SlogContext {
	if ctx == nil {
		return nil
	}
	if s, ok := ctx.Value(ctxKey{}).(*SlogContext); ok {
		return s
	}
	return nil
}

// L returns the context-bound logger or slog.Default if none is attached.
func L(ctx context.Context) *slog.Logger {
	if s := Extract(ctx); s != nil {
		return s.Logger()
	}
	return slog.Default()
}

// Sampled returns the context-bound logger only when the request is
// marked as sampled; otherwise a no-op logger is returned. Useful for
// verbose per-request debug logs gated by tracing decisions.
func Sampled(ctx context.Context) *slog.Logger {
	if s := Extract(ctx); s != nil && s.Sampled {
		return s.Logger()
	}
	return nopLogger
}
