# httpx

由 goost 构建块组装出的 `*http.Client`：支持 backoff 重试、可选限流、
可选熔断器和可选请求日志。

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
                "method", e.Method,
                "host", e.Host,
                "path", e.Path,
                "status", e.StatusCode,
                "delay", e.Delay,
                "error", e.Err)
        },
    },
    Limiter: ratelimit.NewBucket(50, 100),       // 50 req/s，突发 100
    Breaker: circuitbreaker.New(circuitbreaker.Config{
        FailureThreshold: 5,
        CooldownPeriod:   30 * time.Second,
    }),
    Logger: slog.Default(),
})

resp, err := c.Get("https://api.example.com/users")
```

默认重试策略会重试传输错误、HTTP 429 和所有 5xx。可用
`RetryPolicy.RetryOn` 覆盖。传给 `Post` 的 body 会被缓冲，以便重试时回放。
`RetryPolicy.OnRetry` 只会在确实将发起下一次尝试时运行。Retry event 会包含
脱敏请求元数据：method、scheme、host 和 path，但不包含 query string 或 body。

设置 `Logger` 后，`httpx` 会在重试完成后为每个请求记录一行摘要。日志包含
method、scheme、host、path、status、attempts、duration 和 error。查询字符串
和 body 会被刻意省略。
