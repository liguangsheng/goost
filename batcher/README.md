# batcher

DataLoader-style coalescing of concurrent per-key requests into one
batch call. Solves a problem [`x/sync/singleflight`][sf] does not:

| | same key | different keys |
| --- | --- | --- |
| singleflight | dedupe | — |
| **batcher**  | dedupe within window | **coalesce into one batch call** |

[sf]: https://pkg.go.dev/golang.org/x/sync/singleflight

## Usage

```go
func loadUsers(ctx context.Context, ids []int) (map[int]*User, error) {
    // SELECT * FROM users WHERE id IN (?)
    return q.Users(ctx, ids)
}

b := batcher.New(loadUsers).
    MaxBatch(100).            // flush early once the batch holds 100 keys
    MaxWait(5*time.Millisecond). // flush at the latest 5ms after the first key
    Build()

u, err := b.Load(ctx, 42)
```

Concurrent calls within the window collapse to one `loadUsers` call.
Duplicate keys are deduped automatically — wrapping `batcher` in a
singleflight is unnecessary.

`LoadMany(ctx, keys)` fans out keys in parallel, then returns per-key
results and per-key errors:

```go
vals, errs := b.LoadMany(ctx, []int{1, 2, 99})
// vals: {1: u1, 2: u2}
// errs: {99: batcher.ErrNotFound}
```

## Tuning

`Stats()` exposes batch behavior for tuning `MaxBatch` / `MaxWait`:

```go
s := b.Stats()
// s.Batches      — total loadFn invocations
// s.Loads        — total Load calls
// s.Coalesced    — Load calls that joined an existing window
// s.MaxBatchSize — largest batch seen
// s.PendingKeys  — unique keys waiting in the current open window
// s.InFlight     — loadFn calls currently running
// s.MaxBatch     — configured batch-size cap
// s.MaxWait      — configured maximum wait for an open window
```

The returned value is a point-in-time, read-only snapshot. `Batches`, `Loads`,
`Coalesced`, and `MaxBatchSize` are cumulative counters. `PendingKeys` and
`InFlight` are current values at the time of the call. `MaxBatch` and `MaxWait`
are configuration values copied into the snapshot for metrics labels and logs.

If `Coalesced` stays near zero, your windows are too short for the
arrival rate; raise `MaxWait`. If `MaxBatchSize` constantly hits
`MaxBatch`, raise the cap or your downstream is the bottleneck.

## Notes

- A panic inside `loadFn` is recovered and surfaced as an error to
  every caller in the batch.
- If `loadFn` returns a map missing some keys, those callers receive
  `ErrNotFound`. Other batchers conflate "missing" with "zero value";
  this one does not.
- The context passed to `loadFn` is the one given to `Builder.Context`
  (default `context.Background()`). Individual `Load(ctx, ...)` callers
  use their own `ctx` for waiting only — canceling does not abort the
  in-flight batch.
