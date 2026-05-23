# ratelimit

Two concurrency-safe rate limiters.

## Token bucket

Burst-capable. Tokens accrue at a constant rate up to a cap; each request
consumes tokens.

```go
b := ratelimit.NewBucket(100, 200) // 100/s, burst 200

if b.Allow() {
    handle()
}

// or block until allowed (respecting ctx)
if err := b.Wait(ctx, 1); err == nil {
    handle()
}

snap := b.Snapshot()
metrics.RecordTokens(snap.Tokens, snap.Burst)
```

## Leaky bucket

Smooth pacing: at most one request per interval, no burst.

```go
l := ratelimit.NewLeaky(50 * time.Millisecond) // 20 req/s pacing

if l.Allow() {
    handle()
}
if err := l.Wait(ctx); err == nil {
    handle()
}

snap := l.Snapshot()
metrics.RecordLimiterDelay(snap.AvailableIn)
```

Both limiters expose `Snapshot()` for read-only metrics/logging. The returned
value is a point-in-time snapshot. Bucket `Rate` and `Burst`, and leaky
`Interval`, are configuration values copied into the snapshot for metrics labels
and logs. Bucket `Tokens` / `LastRefill`, and leaky `Next` / `AvailableIn`, are
current values at the time of the call.

`Wait` methods respect context cancellation while waiting. A canceled wait
returns `ctx.Err()` and does not reserve or consume a future token or leaky-bucket
slot; later callers can still use the limiter normally when capacity becomes
available.

Both limiters expose `SetClock(fn func() time.Time)` for deterministic
tests. Pair them with [`clock.Mock`](../clock):

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
b.Allow()              // true, spends the initial burst token
b.Allow()              // false until enough time passes
m.Advance(time.Second) // refills
b.Allow()              // true
```
