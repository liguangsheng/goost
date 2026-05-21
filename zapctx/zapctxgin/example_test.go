package zapctxgin

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
)

func ExampleMiddleware() {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(Middleware(zap.NewNop()))
	engine.GET("/", func(c *gin.Context) {
		fmt.Println(zapctx.Extract(c.Request.Context()) != nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	engine.ServeHTTP(httptest.NewRecorder(), req)

	// Output:
	// true
}
