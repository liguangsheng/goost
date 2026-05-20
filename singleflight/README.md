# singleflight

Generic thin wrapper around `golang.org/x/sync/singleflight`.

```go
type User struct{ /* ... */ }

g := singleflight.NewString[*User]()

// Hundreds of concurrent requests for the same id share one DB call.
u, err, shared := g.Do(id, func() (*User, error) {
    return db.LoadUser(id)
})
```

For non-string keys, supply a stringify function:

```go
g := singleflight.New[int, *User](strconv.Itoa)
g.Do(42, ...)
```
