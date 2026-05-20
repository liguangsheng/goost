# pool

A bounded goroutine pool with optional queueing.

```go
p, err := pool.NewPool(64, 0, 4) // size 64, no queue, 4 eager workers
if err != nil { /* ... */ }
defer p.Close()

if err := p.Schedule(func() { doWork() }); err != nil {
    // ErrPoolClosed
}

if err := p.ScheduleTimeout(50*time.Millisecond, func() { doWork() }); err != nil {
    // ErrScheduleTimeout if no worker becomes available in time
}
```

- `Schedule` blocks until a slot is available.
- `ScheduleTimeout` gives up after the deadline.
- Task panics are recovered and logged with the zap global; the pool stays usable.
- `Close` stops accepting new work and waits for in-flight workers to drain.
