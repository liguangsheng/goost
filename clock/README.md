# clock

A `Clock` interface plus `Real` and `Mock` implementations for tests that
need to control time without sleeping.

```go
import "github.com/liguangsheng/goost/clock"

func usesClock(c clock.Clock) {
    deadline := c.Now().Add(time.Second)
    <-c.After(time.Until(deadline))
    // ...
}

// production
usesClock(clock.Real())

// tests
m := clock.NewMock(time.Unix(0, 0))
go usesClock(m)
m.Advance(time.Second) // any After/Sleep waiter whose deadline ≤ now fires
```

`Clock` covers four scheduling primitives, all driveable by `Mock`:

| `Clock` method | stdlib counterpart |
| --- | --- |
| `Now()` | `time.Now` |
| `After(d)` | `time.After` |
| `Sleep(d)` | `time.Sleep` |
| `AfterFunc(d, fn)` | `time.AfterFunc` |
| `NewTicker(d)` | `time.NewTicker` |

A ticker driven by `Mock` only ticks when you `Advance`. Missed ticks
drop (channel cap 1) — same as `time.Ticker`:

```go
m := clock.NewMock(time.Unix(0, 0))
tk := m.NewTicker(time.Second)
defer tk.Stop()

m.Advance(time.Second)
fmt.Println(<-tk.C()) // 1970-01-01 00:00:01 UTC
```

`Mock.Now` matches `func() time.Time`, the signature already accepted by
`backoff.Backoff.Rand`, `ratelimit.Bucket.SetClock`, `circuitbreaker.Config.Now`,
and similar fields. Drive those packages from a mock without touching their
APIs:

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
```
