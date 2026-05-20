package zapctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

func Test_BetterDefault(t *testing.T) {
	assert.NoError(t, BetterDefault())
}
