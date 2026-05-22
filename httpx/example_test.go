package httpx_test

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/httpx"
)

func ExampleNew_retryAndLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client := httpx.New(httpx.Options{
		Timeout: 5 * time.Second,
		Retry: &httpx.RetryPolicy{
			MaxAttempts: 3,
			Backoff: &backoff.Backoff{
				Initial: 100 * time.Millisecond,
				Max:     time.Second,
			},
			OnRetry: func(e httpx.RetryEvent) {
				logger.Info("retrying request",
					"attempt", e.Attempt,
					"max_attempts", e.MaxAttempts,
					"status", e.StatusCode,
					"delay", e.Delay,
					"error", e.Err)
			},
			OnGiveUp: func(e httpx.RetryEvent) {
				logger.Warn("request retries exhausted",
					"attempt", e.Attempt,
					"max_attempts", e.MaxAttempts,
					"status", e.StatusCode,
					"error", e.Err)
			},
		},
		Logger: logger,
	})

	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/users", nil)
	if err != nil {
		panic(err)
	}
	_ = client
	_ = req
}
