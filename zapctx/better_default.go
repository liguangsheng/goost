package zapctx

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func BetterDefault() {
	level := zapcore.DebugLevel
	env := strings.ToUpper(os.Getenv("APP_ENV"))
	if env == "production" || env == "prod" {
		level = zapcore.InfoLevel
	}

	logger, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: true,
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
	}.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
}
