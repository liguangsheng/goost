# ratelimit

两个并发安全的限流器。

## Token bucket

支持突发。token 以恒定速率增长到上限；每个请求消耗 token。

```go
b := ratelimit.NewBucket(100, 200) // 100/s，突发 200

if b.Allow() {
    handle()
}

// 或阻塞直到允许通过（尊重 ctx）
if err := b.Wait(ctx, 1); err == nil {
    handle()
}

snap := b.Snapshot()
metrics.RecordTokens(snap.Tokens, snap.Burst)
```

## Leaky bucket

平滑节奏：每个间隔最多一个请求，不支持突发。

```go
l := ratelimit.NewLeaky(50 * time.Millisecond) // 20 req/s 节奏

if l.Allow() {
    handle()
}
if err := l.Wait(ctx); err == nil {
    handle()
}

snap := l.Snapshot()
metrics.RecordLimiterDelay(snap.AvailableIn)
```

两个限流器都暴露 `Snapshot()`，用于只读指标或日志。返回值是调用时刻的
只读 snapshot。Bucket 的 `Rate` 和 `Burst`、leaky 的 `Interval` 是复制到 snapshot
中的配置值，便于作为 metrics label 或日志字段。Bucket 的 `Tokens` / `LastRefill`
以及 leaky 的 `Next` / `AvailableIn` 是调用时刻的当前值。

`Wait` 方法会在等待期间尊重 context cancellation。取消的 wait 会返回 `ctx.Err()`，
并且不会预留或消耗未来 token 或 leaky-bucket slot；容量恢复后，后续调用者仍可正常使用 limiter。

两个限流器都暴露 `SetClock(fn func() time.Time)`，方便确定性测试。
可配合 [`clock.Mock`](../clock)：

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
b.Allow()              // true，消耗初始突发 token
b.Allow()              // 时间推进前为 false
m.Advance(time.Second) // 补充 token
b.Allow()              // true
```
