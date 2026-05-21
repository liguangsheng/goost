# ttlmap

带单条目过期时间的并发 map。与 `lru.Cache` 不同，它没有容量上限；
大小只由过期策略控制。

```go
m := ttlmap.New[string, *Session](time.Minute) // 每分钟扫描过期条目
defer m.Close()

m.Set(token, sess, 10*time.Minute)

if s, ok := m.Get(token); ok {
    serve(s)
}
```

传 `0` 给 `New` 可禁用后台扫描；过期条目仍会在访问时删除。
