# keyedmutex

Per-key mutex: many goroutines can hold locks on different keys
concurrently; any two that pick the same key serialize.

Standard `sync.Mutex` doesn't fit when the contention domain is keyed
(per user ID, per resource path) and the key space is large or
short-lived — you don't want to preallocate one `sync.Mutex` per
possible key.

## Usage

```go
m := keyedmutex.New[string]()

m.Lock("user:42")
defer m.Unlock("user:42")
// ... mutate user 42 ...
```

`TryLock` for the non-blocking variant, `LockContext` for ctx-aware
waits, `WithLock` for the closure form:

```go
err := m.WithLock(ctx, "user:42", func() error {
    return updateUser(ctx, 42)
})
```

## Notes

- Slots are allocated lazily on first `Lock` and freed once no
  goroutine holds or waits on them. A churn of millions of distinct
  one-shot keys does **not** grow the internal map.
- `Unlock` on a key that isn't locked panics — same contract as
  `sync.Mutex`.
- `Len()` returns the number of keys currently locked or with
  waiters; useful for diagnostics.
- Locks are not reentrant. A goroutine that calls `Lock(k)` twice
  for the same `k` deadlocks itself — same as `sync.Mutex`.
