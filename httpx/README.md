# httpx

`*http.Client` assembled from goost building blocks: retry with backoff,
optional rate limiting, optional circuit breaker, and optional request logging.

```go
c := httpx.New(httpx.Options{
    Timeout: 5 * time.Second,
    Retry: &httpx.RetryPolicy{
        MaxAttempts: 3,
        Backoff: &backoff.Backoff{
            Initial: 100 * time.Millisecond,
            Max:     2 * time.Second,
            Jitter:  0.2,
        },
        OnRetry: func(e httpx.RetryEvent) {
            slog.Info("retrying outbound request",
                "attempt", e.Attempt,
                "max_attempts", e.MaxAttempts,
                "status", e.StatusCode,
                "delay", e.Delay,
                "error", e.Err)
        },
    },
    Limiter: ratelimit.NewBucket(50, 100),       // 50 req/s, burst 100
    Breaker: circuitbreaker.New(circuitbreaker.Config{
        FailureThreshold: 5,
        CooldownPeriod:   30 * time.Second,
    }),
    Logger: slog.Default(),
})

resp, err := c.Get("https://api.example.com/users")
```

The default retry policy retries on transport errors, HTTP 429, and any
5xx. Override with `RetryPolicy.RetryOn`. Bodies passed to `Post` are
buffered so they can be replayed on retry. `RetryPolicy.OnRetry` runs only
when another attempt will be made.

When `Logger` is set, `httpx` logs one summary line per request after retries
finish. The log includes method, scheme, host, path, status, attempts,
duration, and error. Query strings and bodies are intentionally omitted.
