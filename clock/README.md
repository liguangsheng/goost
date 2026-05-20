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

`Mock.Now` matches `func() time.Time`, the signature already accepted by
`backoff.Backoff.Rand`, `ratelimit.Bucket.SetClock`, `circuitbreaker.Config.Now`,
and similar fields. Drive those packages from a mock without touching their
APIs:

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
```
