# batcher

DataLoader 风格的批处理器：把并发的按 key 请求合并成一次批量调用。
它解决了 [`x/sync/singleflight`][sf] 不处理的问题：

| | 相同 key | 不同 key |
| --- | --- | --- |
| singleflight | 去重 | - |
| **batcher** | 窗口内去重 | **合并成一次批量调用** |

[sf]: https://pkg.go.dev/golang.org/x/sync/singleflight

## 用法

```go
func loadUsers(ctx context.Context, ids []int) (map[int]*User, error) {
    // SELECT * FROM users WHERE id IN (?)
    return q.Users(ctx, ids)
}

b := batcher.New(loadUsers).
    MaxBatch(100).            // 批次达到 100 个 key 时提前刷新
    MaxWait(5*time.Millisecond). // 第一个 key 到达后最多等待 5ms
    Build()

u, err := b.Load(ctx, 42)
```

窗口内的并发调用会折叠成一次 `loadUsers` 调用。重复 key 会自动去重，
因此没有必要再在 `batcher` 外层包一层 singleflight。

`LoadMany(ctx, keys)` 会并行分发多个 key，并返回按 key 划分的结果和错误：

```go
vals, errs := b.LoadMany(ctx, []int{1, 2, 99})
// vals: {1: u1, 2: u2}
// errs: {99: batcher.ErrNotFound}
```

## 调优

`Stats()` 暴露批处理行为，便于调节 `MaxBatch` / `MaxWait`：

```go
s := b.Stats()
// s.Batches      - loadFn 调用总数
// s.Loads        - Load 调用总数
// s.Coalesced    - 加入已有窗口的 Load 调用数
// s.MaxBatchSize - 见过的最大批次大小
```

如果 `Coalesced` 长期接近 0，说明窗口相对于到达速率太短，可以提高
`MaxWait`。如果 `MaxBatchSize` 经常打满 `MaxBatch`，可以提高上限，
或者下游已经成为瓶颈。

## 说明

- `loadFn` 内部 panic 会被恢复，并以错误形式返回给该批次的所有调用者。
- 如果 `loadFn` 返回的 map 缺少某些 key，对应调用者会收到
  `ErrNotFound`。有些批处理器会把“缺失”和“零值”混在一起，这里不会。
- 传给 `loadFn` 的 context 来自 `Builder.Context`，默认是
  `context.Background()`。单个 `Load(ctx, ...)` 调用者自己的 `ctx`
  只用于等待；取消它不会中止已经在飞行中的批次。
