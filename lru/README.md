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

### Notes

- `Get` evicts entries whose expiration has passed.
- `Peek` returns a value without updating recency.
- `Size` / `Clear` are safe to call from multiple goroutines.
