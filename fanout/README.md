# fanout

In-process broadcaster: each `Publish(v)` is delivered to every current
subscriber. Useful for event buses, log tees, config-reload signals, or
anything where one producer's output needs to reach N independent
consumers in the same process.

```go
b := fanout.New[Event]().Buffer(64).Build()

sub := b.Subscribe()
defer sub.Close()

go func() {
    for ev := range sub.C() {
        handle(ev)
    }
}()

b.Publish(Event{...})
```

## Backpressure: drop, don't block

A broadcaster that blocks `Publish` on the slowest subscriber lets one
stuck consumer halt the whole system. `fanout` instead drops messages
on per-subscriber-full and counts the drops. Publishers stay fast;
slow consumers fall behind on their own.

```go
sub.Drops()      // messages this sub missed
b.Stats().Drops  // aggregate across all subs (alive and closed)
```

`Stats()` also reports current subscribers, configured per-subscriber
buffer size, queued messages across active subscribers, and whether the
broadcaster is closed.

The returned value is a point-in-time, read-only snapshot. `Publishes` and
`Drops` are cumulative counters. `Subscribers`, `Queued`, and `Closed` are
current values at the time of the call. `Buffer` belongs to configuration values:
the per-subscriber buffer size copied into the snapshot for metrics labels and logs.

If you cannot tolerate drops, set `Buffer` high enough to absorb your
expected burst.

## Lifecycle

- `Sub.Close()` unsubscribes; the channel `Sub.C()` is closed so
  `for v := range sub.C()` exits cleanly.
- `Broadcaster.Close()` closes every subscriber's channel and turns
  future `Publish` into a no-op.
- Both `Close` calls are idempotent.
- A subscription created after `Broadcaster.Close` returns a
  pre-closed channel (so it's safe to plug into a `for range` loop
  unconditionally).

## Notes

- Subscribers see values published **after** their `Subscribe`
  call. There is no replay of past values.
- Each subscriber's buffer is set at `Subscribe` time from the
  Builder's `Buffer(n)`. Changing the builder later does not affect
  existing subscribers.
