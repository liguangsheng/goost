# keyedmutex

按 key 划分的 mutex：多个 goroutine 可以同时持有不同 key 的锁；
选择同一个 key 的 goroutine 会串行执行。

当竞争域是 key（例如用户 ID、资源路径），且 key 空间很大或生命周期很短时，
标准 `sync.Mutex` 不太合适；你不会想为每个可能的 key 预分配一个
`sync.Mutex`。

## 用法

```go
m := keyedmutex.New[string]()

m.Lock("user:42")
defer m.Unlock("user:42")
// ... 修改 user 42 ...
```

`TryLock` 提供非阻塞版本，`LockContext` 支持 context 感知等待，
`WithLock` 提供闭包形式：

```go
err := m.WithLock(ctx, "user:42", func() error {
    return updateUser(ctx, 42)
})
```

## 说明

- slot 会在第一次 `Lock` 时惰性分配，并在没有 goroutine 持有或等待它时释放。
  即使有数百万个一次性 key 反复出现，内部 map 也不会持续增长。
- 对未加锁的 key 调用 `Unlock` 会 panic，契约与 `sync.Mutex` 相同。
- `Len()` 返回当前已加锁或有等待者的 key 数量，便于诊断。
- 锁不可重入。同一个 goroutine 对同一个 `k` 调用两次 `Lock(k)` 会把自己锁死，
  与 `sync.Mutex` 相同。
