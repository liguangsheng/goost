# defaultmap

按需构造默认值的并发 map。模型类似 Python 的 `collections.defaultdict`。

```go
counts := defaultmap.Make(func(k string) int { return 0 })
counts.Set("foo", counts.Get("foo")+1)

// Get 返回惰性构造的零值：
n := counts.Get("bar") // 0
```

构造函数不能在同一个 key 上回调同一个 map，否则会死锁。构造函数会在持有写锁时执行，
所以应保持轻量。
