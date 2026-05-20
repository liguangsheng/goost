package redact

import (
	"log/slog"

	"go.uber.org/zap"
)

// ZapString returns a zap.Field whose value is Mask(value, left, right).
func ZapString(key, value string, keepLeft, keepRight int) zap.Field {
	return zap.String(key, Mask(value, keepLeft, keepRight))
}

// SlogString returns an slog.Attr whose value is Mask(value, left, right).
func SlogString(key, value string, keepLeft, keepRight int) slog.Attr {
	return slog.String(key, Mask(value, keepLeft, keepRight))
}
