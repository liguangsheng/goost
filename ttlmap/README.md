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

`Close` stops the background sweep goroutine and is safe to call more than
once. It does not clear the map and does not make the map unusable: `Set`,
`Get`, `Delete`, `Len`, and `PurgeExpired` still work after `Close`. Expired
entries continue to be removed lazily by `Get` or explicitly by `PurgeExpired`.

`Len` reports the number of stored entries, including expired entries that have
not yet been read or swept.

Use `PurgeExpired` to remove expired entries on demand when background sweeping
is disabled or when you want to bound stale entries before measuring size. It
returns the number of removed entries and fires `OnExpire` for each removal.
`OnExpire` fires only for TTL expiration observed by `Get`, background sweep,
or `PurgeExpired`; `Delete` and `Close` do not call it.
