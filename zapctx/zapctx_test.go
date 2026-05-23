package zapctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func Test_ToAndExtract(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := ToContext(context.Background(), logger)
	assert.NotNil(t, Extract(ctx))
	assert.Equal(t, logger, Extract(ctx).logger)
}

func Test_ExtractMissing(t *testing.T) {
	var nilCtx context.Context
	assert.Nil(t, Extract(nilCtx))
	assert.Nil(t, Extract(context.Background()))
}

func Test_L_FallbackToGlobal(t *testing.T) {
	assert.Equal(t, zap.L(), L(context.Background()))
}

func Test_LWithAddedFields(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	ctx := ToContext(context.Background(), logger)
	Extract(ctx).AddFields(zap.String("request_id", "abc"), zap.Int("attempt", 2))

	L(ctx).Info("handled")

	entry := logs.All()[0]
	assert.Equal(t, "handled", entry.Message)
	assert.Equal(t, "abc", entry.ContextMap()["request_id"])
	assert.Equal(t, int64(2), entry.ContextMap()["attempt"])
}

func Test_SampledUsesLoggerOnlyWhenSampled(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	ctx := ToContext(context.Background(), logger)
	Extract(ctx).AddFields(zap.String("request_id", "abc"))

	Sampled(ctx).Debug("not sampled")
	assert.Empty(t, logs.All())

	Extract(ctx).Sampled = true
	Sampled(ctx).Debug("sampled")

	entry := logs.All()[0]
	assert.Equal(t, "sampled", entry.Message)
	assert.Equal(t, "abc", entry.ContextMap()["request_id"])
}

func Test_BetterDefault(t *testing.T) {
	assert.NoError(t, BetterDefault())
}
