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
```

Both limiters expose `SetClock(fn func() time.Time)` for deterministic
tests. Pair them with [`clock.Mock`](../clock):

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
b.Allow()              // false after the burst is spent
m.Advance(time.Second) // refills
b.Allow()              // true
```
