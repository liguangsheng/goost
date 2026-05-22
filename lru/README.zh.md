# lru

泛型 LRU 缓存，支持可选的单条目过期时间和驱逐 hook。

```go
c := lru.New[string, string]().Cap(1000).Build()
c.Set("hello", "world")
c.SetWithDuration("session", "abc", time.Minute)

if v, ok := c.Get("hello"); ok {
    fmt.Println(v) // world
}
```

### Builder 选项

- `Cap(n)`：最大条目数。`0` 表示不按容量驱逐。
- `Safe(bool)`：开关内部锁。默认 `true`。
- `Evict(fn)`：容量驱逐时调用的 hook。
- `Shards(n, hashFn)`：把缓存分成 `n` 个 shard；使用 `BuildSharded()`
  而不是 `Build()`，以降低锁竞争。

### 说明

- `Get` 会驱逐已经过期的条目。
- `Peek` 返回值但不更新最近使用顺序。
- `Snapshot` 会报告当前大小、配置容量和 shard 数，便于指标和日志采集，且不会改变最近使用顺序。
- `Size` / `Clear` 可以安全地从多个 goroutine 调用。

### 基准

单线程，13th Gen i5-13600KF，Linux，Go 1.25，容量 1M：

| op | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| goostlru `Set` (safe) | 481.1 | 109 | 3 |
| goostlru `Get` (safe) | 175.6 | 7 | 0 |
| goostlru `Set` (unsafe) | 478.7 | 110 | 3 |
| goostlru `Get` (unsafe) | 171.0 | 7 | 0 |
| golang-lru/v2 `Set` | 466.3 | 118 | 2 |
| golang-lru/v2 `Get` | 221.7 | 7 | 0 |
| gcache lru `Set` | 415.7 | 143 | 4 |
| gcache lru `Get` | 275.2 | 23 | 1 |
| ristretto `Set` | 126.9 | 104 | 2 |
| ristretto `Get` | 67.7 | 11 | 1 |

`ristretto` 更快，因为它在缓冲通道后批处理，并且面向分片高并发工作负载。
对于单 key 路径和严格 LRU 语义，goost 仍然能与 `golang-lru/v2` 保持竞争力。
