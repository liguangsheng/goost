# defaultmap

A concurrent map that constructs default values on demand. Modeled after
Python's `collections.defaultdict`.

```go
counts := defaultmap.Make(func(k string) int { return 0 })
counts.Set("foo", counts.Get("foo")+1)

// Get returns the lazily constructed zero value:
n := counts.Get("bar") // 0
```

The constructor must not call back into the same map on the same key;
that will deadlock. The constructor runs while a write lock is held, so
keep it cheap.
