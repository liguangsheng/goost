# circuitbreaker

Three-state breaker (closed / open / half-open) to short-circuit calls to
a failing dependency.

```go
b := circuitbreaker.New(circuitbreaker.Config{
    FailureThreshold:  5,
    CooldownPeriod:    30 * time.Second,
    HalfOpenSuccesses: 1,
})

err := b.Do(ctx, func(ctx context.Context) error {
    return callDownstream(ctx)
})
if errors.Is(err, circuitbreaker.ErrOpen) {
    return fallback()
}

snap := b.Snapshot()
metrics.RecordBreakerState(snap.State.String(), snap.Failures, snap.CooldownRemaining)
```

### Behavior

- **closed**: all calls pass through; consecutive failures up to
  `FailureThreshold` trip the breaker.
- **open**: calls return `ErrOpen` immediately; after `CooldownPeriod`,
  the breaker becomes half-open.
- **half-open**: a single probe is allowed at a time. After
  `HalfOpenSuccesses` consecutive successes the breaker closes; any
  failure re-opens it.

### Customization

- `IsFailure` lets you exclude expected errors like `context.Canceled`
  from the failure count.
- `OnStateChange` fires on every transition for metrics/logging.
- `Snapshot` returns the current state, failure counters, open time, and
  cooldown remaining for metrics/logging.
- `Now` overrides the clock for deterministic tests; pair with
  [`clock.Mock`](../clock):

```go
m := clock.NewMock(time.Unix(0, 0))
b := circuitbreaker.New(circuitbreaker.Config{
    FailureThreshold: 1,
    CooldownPeriod:   time.Minute,
    Now:              m.Now,
})
m.Advance(time.Minute) // moves cooldown forward
```
