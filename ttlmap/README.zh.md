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

`Len` 报告当前存储的条目数，包括已经过期但尚未被读取或扫描清理的条目。

后台扫描禁用时，或需要在统计大小前主动限制陈旧条目时，可使用
`PurgeExpired` 按需删除已过期条目。它会返回删除数量，并为每个删除条目触发
`OnExpire`。`OnExpire` 只会在 `Get`、后台扫描或 `PurgeExpired` 观察到 TTL
过期时触发；`Delete` 和 `Close` 不会调用它。
