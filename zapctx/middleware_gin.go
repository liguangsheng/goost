package zapctx

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GinMiddleware(logger *zap.Logger, hooks ...HookFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newCtx := ToContext(ctx.Request.Context(), logger)
		for _, hook := range hooks {
			newCtx = hook(newCtx)
		}
		ctx.Request = ctx.Request.WithContext(newCtx)
	}
}
