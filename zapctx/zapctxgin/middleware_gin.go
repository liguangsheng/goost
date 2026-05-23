package zapctxgin

import (
	"github.com/gin-gonic/gin"
	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
)

// Middleware returns a Gin middleware that attaches a request-scoped zap
// logger to the request context.
func Middleware(logger *zap.Logger, hooks ...zapctx.HookFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newCtx := zapctx.ToContext(ctx.Request.Context(), logger)
		for _, hook := range hooks {
			newCtx = hook(newCtx)
		}
		ctx.Request = ctx.Request.WithContext(newCtx)
	}
}
