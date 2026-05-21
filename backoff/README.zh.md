# backoff

带可选抖动的指数退避，并提供感知 `context` 的 `Retry` 辅助函数。

```go
b := &backoff.Backoff{
    Initial: 200 * time.Millisecond,
    Max:     10 * time.Second,
    Factor:  2.0,
    Jitter:  0.2, // ±20% 随机抖动
}

err := backoff.Retry(ctx, b, 5, func(ctx context.Context) error {
    return doSomething(ctx)
})
```

在回调中返回 `backoff.Permanent(err)` 可立即中止重试。传入
`maxAttempts = 0` 表示无限重试，最终由 `ctx` 控制退出。

`Backoff.Rand` 接受一个返回 `[0, 1)` 的函数，测试中可以用它固定抖动。
如果消费 `Backoff.Next()` 返回的持续时间，也可以配合 [`clock`](../clock)
确定性推进时间。
