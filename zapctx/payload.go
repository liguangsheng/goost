package zapctx

import (
	"bytes"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PayloadOption configures the gin payload-logging middleware.
type PayloadOption func(*payloadConfig)

type payloadConfig struct {
	maxBody     int
	sampleEvery int64 // 1 = every request; N = every Nth
	skip        func(*gin.Context) bool
}

// WithMaxBody caps the number of bytes logged per request and response.
// 0 means do not log the body. Defaults to 4096.
func WithMaxBody(n int) PayloadOption { return func(c *payloadConfig) { c.maxBody = n } }

// WithSampling logs every n-th request (n >= 1). Defaults to 1 (every request).
func WithSampling(n int) PayloadOption {
	return func(c *payloadConfig) {
		if n < 1 {
			n = 1
		}
		c.sampleEvery = int64(n)
	}
}

// WithSkipper skips requests for which fn returns true. Common usage:
// skip health-check endpoints.
func WithSkipper(fn func(*gin.Context) bool) PayloadOption {
	return func(c *payloadConfig) { c.skip = fn }
}

// PayloadGinMiddleware logs request/response bodies, status, latency, and
// trace fields attached by upstream hooks. Bodies are truncated by maxBody
// to keep log lines bounded. Bodies are kept fully readable by handlers.
func PayloadGinMiddleware(logger *zap.Logger, opts ...PayloadOption) gin.HandlerFunc {
	cfg := &payloadConfig{maxBody: 4096, sampleEvery: 1}
	for _, o := range opts {
		o(cfg)
	}
	var counter atomic.Int64

	return func(c *gin.Context) {
		if cfg.skip != nil && cfg.skip(c) {
			c.Next()
			return
		}
		if cfg.sampleEvery > 1 && counter.Add(1)%cfg.sampleEvery != 0 {
			c.Next()
			return
		}

		start := time.Now()
		reqBody := readAndRestore(c.Request, cfg.maxBody)

		rw := &capturingWriter{ResponseWriter: c.Writer, buf: &bytes.Buffer{}, max: cfg.maxBody}
		c.Writer = rw
		c.Next()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		}
		if cfg.maxBody > 0 {
			fields = append(fields,
				zap.ByteString("request_body", reqBody),
				zap.ByteString("response_body", rw.captured()),
			)
		}
		L(c.Request.Context()).With(fields...).Info("http")
		// reuse logger to keep linter happy and allow future filtering by name
		_ = logger
	}
}

func readAndRestore(req *http.Request, max int) []byte {
	if max <= 0 || req.Body == nil {
		return nil
	}
	all, err := io.ReadAll(req.Body)
	if err != nil {
		return nil
	}
	_ = req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(all))
	if len(all) > max {
		return all[:max]
	}
	return all
}

type capturingWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
	max int
}

func (w *capturingWriter) Write(p []byte) (int, error) {
	if w.max > 0 {
		remaining := w.max - w.buf.Len()
		if remaining > 0 {
			if remaining >= len(p) {
				w.buf.Write(p)
			} else {
				w.buf.Write(p[:remaining])
			}
		}
	}
	return w.ResponseWriter.Write(p)
}

func (w *capturingWriter) captured() []byte { return w.buf.Bytes() }
