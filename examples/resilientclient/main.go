// resilientclient demonstrates httpx + ratelimit + circuitbreaker: outbound
// requests are rate-limited, retried on transient 5xx responses, and protected
// by a circuit breaker after repeated downstream failures.
//
// Run from examples/: go run ./resilientclient
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/circuitbreaker"
	"github.com/liguangsheng/goost/httpx"
	"github.com/liguangsheng/goost/ratelimit"
)

func main() {
	var calls atomic.Int64
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		switch calls.Add(1) {
		case 1, 2:
			http.Error(w, "try again", http.StatusBadGateway)
		default:
			_, _ = io.WriteString(w, "ok")
		}
	}))
	defer downstream.Close()

	client := httpx.New(httpx.Options{
		Retry: &httpx.RetryPolicy{
			MaxAttempts: 3,
			Backoff:     &backoff.Backoff{Initial: 10 * time.Millisecond, Max: 20 * time.Millisecond},
		},
		Limiter: ratelimit.NewBucket(20, 1),
		Breaker: circuitbreaker.New(circuitbreaker.Config{
			FailureThreshold: 3,
			CooldownPeriod:   time.Second,
		}),
	})

	resp, err := client.Get(downstream.URL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status=%d body=%s calls=%d\n", resp.StatusCode, body, calls.Load())
}
