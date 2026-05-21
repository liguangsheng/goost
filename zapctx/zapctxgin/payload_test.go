package zapctxgin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newObservedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.InfoLevel)
	return zap.New(core), logs
}

func Test_PayloadGinMiddleware_LogsBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, logs := newObservedLogger()

	e := gin.New()
	e.Use(Middleware(logger))
	e.Use(PayloadMiddleware(logger))
	e.POST("/echo", func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.Data(http.StatusCreated, "text/plain", body)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString("hello"))
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "hello", w.Body.String())
	if assert.Equal(t, 1, logs.Len()) {
		entry := logs.All()[0]
		fields := entry.ContextMap()
		assert.Equal(t, int64(http.StatusCreated), fields["status"])
		assert.Equal(t, "hello", fields["request_body"])
		assert.Equal(t, "hello", fields["response_body"])
	}
}

func Test_PayloadGinMiddleware_Sampling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, logs := newObservedLogger()

	e := gin.New()
	e.Use(Middleware(logger))
	e.Use(PayloadMiddleware(logger, WithSampling(3), WithMaxBody(0)))
	e.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })

	for range 9 {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		e.ServeHTTP(w, req)
	}

	// 9 requests at sample 3 -> 3 log lines.
	assert.Equal(t, 3, logs.Len())
}

func Test_PayloadGinMiddleware_Skipper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, logs := newObservedLogger()

	e := gin.New()
	e.Use(Middleware(logger))
	e.Use(PayloadMiddleware(logger, WithSkipper(func(c *gin.Context) bool {
		return c.Request.URL.Path == "/health"
	})))
	e.GET("/health", func(c *gin.Context) { c.String(200, "ok") })
	e.GET("/data", func(c *gin.Context) { c.String(200, "ok") })

	for _, path := range []string{"/health", "/data"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		e.ServeHTTP(w, req)
	}
	assert.Equal(t, 1, logs.Len(), "/health must be skipped")
}
