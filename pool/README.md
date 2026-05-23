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
- Task panics are recovered so the pool stays usable; use `WithPanicHandler`
  to observe recovered panic values.
- `Close` stops accepting new work and waits for workers to finish already
  accepted tasks, including queued tasks.
- `Stats` reports workers, capacity, in-flight tasks, queued tasks, completed
  tasks, queue capacity, panic count, and closed state.

## Shutdown Semantics

`Close` is idempotent. The first call closes the pool for new submissions,
closes the internal work queue, and waits for workers to finish every task that
was already accepted. Both `Schedule` and `ScheduleTimeout` return
`ErrPoolClosed` after close starts. Tasks accepted before close still run;
tasks submitted after close do not.

`Stats()` returns a point-in-time, read-only snapshot. `Completed` and `Panics`
are cumulative counters. `Workers`, `InFlight`, `Queued`, and `Closed` are
current values at the time of the call. `Capacity` and `QueueCapacity` are
configuration values copied into the snapshot for metrics labels and logs.
