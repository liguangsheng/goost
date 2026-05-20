# lru

Generic LRU cache with optional per-entry expiration and an evict hook.

```go
c := lru.New[string, string]().Cap(1000).Build()
c.Set("hello", "world")
c.SetWithDuration("session", "abc", time.Minute)

if v, ok := c.Get("hello"); ok {
    fmt.Println(v) // world
}
```

### Builder options

- `Cap(n)` — max entries. `0` disables eviction.
- `Safe(bool)` — toggle internal locking. Default `true`.
- `Evict(fn)` — hook called on capacity-driven eviction.
- `Shards(n, hashFn)` — partition the cache across `n` shards; use
  `BuildSharded()` instead of `Build()` to reduce lock contention.

### Notes

- `Get` evicts entries whose expiration has passed.
- `Peek` returns a value without updating recency.
- `Size` / `Clear` are safe to call from multiple goroutines.

### Benchmark

Single-threaded, 13th Gen i5-13600KF, Linux, Go 1.25, capacity 1M:

| op                       | ns/op | B/op | allocs/op |
| ------------------------ | ----: | ---: | --------: |
| goostlru `Set` (safe)    | 481.1 | 109  | 3 |
| goostlru `Get` (safe)    | 175.6 |   7  | 0 |
| goostlru `Set` (unsafe)  | 478.7 | 110  | 3 |
| goostlru `Get` (unsafe)  | 171.0 |   7  | 0 |
| golang-lru/v2 `Set`      | 466.3 | 118  | 2 |
| golang-lru/v2 `Get`      | 221.7 |   7  | 0 |
| gcache lru `Set`         | 415.7 | 143  | 4 |
| gcache lru `Get`         | 275.2 |  23  | 1 |
| ristretto `Set`          | 126.9 | 104  | 2 |
| ristretto `Get`          |  67.7 |  11  | 1 |

`ristretto` wins because it batches behind buffered channels and is built
for sharded high-concurrency workloads. For a single-key path with strict
LRU semantics, goost stays competitive with `golang-lru/v2`.
