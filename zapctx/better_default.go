package zapctx

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BetterDefault installs a zap global logger configured for either
// development (debug level, dev encoder defaults) or production
// (info level), based on the APP_ENV environment variable.
//
// It returns an error rather than panicking so callers can decide how
// to handle failure.
func BetterDefault() error {
	env := strings.ToLower(os.Getenv("APP_ENV"))
	production := env == "production" || env == "prod"

	level := zapcore.DebugLevel
	if production {
		level = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: !production,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("zapctx: build logger: %w", err)
	}
	zap.ReplaceGlobals(logger)
	return nil
}
