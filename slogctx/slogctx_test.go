package slogctx

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newCapturingLogger() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), buf
}

func Test_ToAndExtract(t *testing.T) {
	logger, _ := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	assert.NotNil(t, Extract(ctx))
}

func Test_ExtractMissing(t *testing.T) {
	var nilCtx context.Context
	assert.Nil(t, Extract(nilCtx))
	assert.Nil(t, Extract(context.Background()))
}

func Test_L_FallbackToDefault(t *testing.T) {
	assert.Equal(t, slog.Default(), L(context.Background()))
}

func Test_LWithAddedAttrs(t *testing.T) {
	logger, buf := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	Extract(ctx).AddAttrs(slog.String("k", "v"))
	L(ctx).Info("hello")
	assert.True(t, strings.Contains(buf.String(), "k=v"))
}

func Test_SampledNoOpWhenNotSampled(t *testing.T) {
	logger, buf := newCapturingLogger()
	ctx := ToContext(context.Background(), logger)
	Sampled(ctx).Info("nope")
	assert.Empty(t, buf.String())
}
