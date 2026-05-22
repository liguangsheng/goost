# ttlmap

A concurrent map with per-entry expiration. Unlike `lru.Cache`, there is no
capacity bound; size is governed by expiration alone.

```go
m := ttlmap.New[string, *Session](time.Minute) // sweep stale entries every minute
defer m.Close()

m.Set(token, sess, 10*time.Minute)

if s, ok := m.Get(token); ok {
    serve(s)
}
```

Pass `0` to `New` to disable background sweeping; expired entries are still
removed on access.

`Len` reports the number of stored entries, including expired entries that have
not yet been read or swept.

Use `PurgeExpired` to remove expired entries on demand when background sweeping
is disabled or when you want to bound stale entries before measuring size. It
returns the number of removed entries and fires `OnExpire` for each removal.
`OnExpire` fires only for TTL expiration observed by `Get`, background sweep,
or `PurgeExpired`; `Delete` and `Close` do not call it.
