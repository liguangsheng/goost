// httpserver demonstrates zapctx + gin + OpenTelemetry hooks: every
// request gets a logger pre-loaded with trace IDs, and PayloadGinMiddleware
// logs the response body and timing.
//
// Run:  go run ./examples/httpserver
// Then: curl -X POST -d hello localhost:8080/echo
package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/liguangsheng/goost/shutdown"
	"github.com/liguangsheng/goost/zapctx"
	"go.uber.org/zap"
)

func main() {
	if err := zapctx.BetterDefault(); err != nil {
		panic(err)
	}
	logger := zap.L()

	e := gin.New()
	e.Use(zapctx.GinMiddleware(logger, zapctx.OtelTraceInject))
	e.Use(zapctx.PayloadGinMiddleware(logger,
		zapctx.WithMaxBody(1024),
		zapctx.WithSkipper(func(c *gin.Context) bool {
			return c.Request.URL.Path == "/healthz"
		}),
	))

	e.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	e.POST("/echo", func(c *gin.Context) {
		body, _ := c.GetRawData()
		zapctx.L(c.Request.Context()).Info("handled echo", zap.Int("bytes", len(body)))
		c.Data(http.StatusOK, "text/plain", body)
	})

	srv := &http.Server{Addr: ":8080", Handler: e}
	shutdown.Add(func() {
		_ = srv.Shutdown(context.Background())
	}, shutdown.WithName("http"), shutdown.WithTimeout(5))

	logger.Info("listening", zap.String("addr", srv.Addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	shutdown.Wait(context.Background())
}
