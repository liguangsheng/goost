# debounce

Coalesce a burst of `Trigger(v)` calls into a single emit on `C()`.
After a `Trigger`, the debouncer waits a quiet window with no further
trigger before forwarding the **most recent** value. Intermediate
values are discarded.

```go
d := debounce.New[string](300 * time.Millisecond)
defer d.Stop()

go func() {
    for v := range d.C() {
        applyConfig(v)
    }
}()

// e.g. a file-watcher that fires repeatedly during a save:
d.Trigger(load("config.yaml"))
d.Trigger(load("config.yaml"))
d.Trigger(load("config.yaml"))
// 300ms later, applyConfig runs once with the latest value
```

## Not a rate limiter

`ratelimit` throttles call rate; `debounce` defers emission until the
input stream is quiet. If you want "at most 5 per second", use
`ratelimit`. If you want "wait until no more events arrive for X, then
act", use `debounce`.

## Backpressure

The output channel has buffer 1 and a **latest-wins** policy: if the
consumer falls behind, a fresh emit overwrites the stale buffered
value. The consumer is guaranteed to eventually see the **latest**
emitted value but may miss intermediate ones.

## Test-friendly

Inject `clock.Clock` to make timing deterministic in tests:

```go
m := clock.NewMock(time.Unix(0, 0))
d := debounce.New[int](100 * time.Millisecond).WithClock(m)

d.Trigger(42)
m.Advance(100 * time.Millisecond) // emit happens at this point
v := <-d.C()
```
