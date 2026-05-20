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

func Test_OpenTraceInjectNoContext(t *testing.T) {
	// Must not panic when ctx has neither ZapContext nor span.
	ctx := OpenTraceInject(context.Background())
	assert.NotNil(t, ctx)
}

func Test_OpenTraceInjectWithLogger(t *testing.T) {
	ctx := ToContext(context.Background(), zap.NewNop())
	// No span attached — should be a no-op.
	out := OpenTraceInject(ctx)
	assert.Equal(t, ctx, out)
}

func Test_BetterDefault(t *testing.T) {
	assert.NoError(t, BetterDefault())
}
