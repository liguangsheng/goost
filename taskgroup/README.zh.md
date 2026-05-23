# taskgroup

`golang.org/x/sync/errgroup` 的小型替代实现，额外提供两点：

- 并发限制
- 任务 panic 会被恢复并以错误形式暴露

```go
g := taskgroup.New(ctx).WithLimit(8)
for _, item := range items {
    item := item
    g.Go(func(ctx context.Context) error {
        return process(ctx, item)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

第一个非 nil 错误会取消 group 的 context，让兄弟任务可以提前退出；
后续错误会被丢弃。任务 panic 会走同一条路径：panic 会被恢复，转换成带
`taskgroup: panic:` 前缀的错误，取消兄弟任务，并在它是第一个失败时由 `Wait`
返回。即使全部成功，`Wait` 返回前也会取消 context；如果存在任务错误或恢复后的
panic 错误，`Cause()` 会返回第一个失败原因。

如果每个任务都会返回值，可使用 `Results[T]`：

```go
g := taskgroup.NewResults[string](ctx).WithLimit(4)
for _, item := range items {
    item := item
    g.Run(func(ctx context.Context) (string, error) {
        return fetch(ctx, item)
    })
}
values, err := g.Wait() // values 按完成顺序返回
```

与 `Group` 一样，`Results[T]` 会在第一个任务错误或恢复后的 panic 出现时取消
context，返回取消前已经完成的 values；如果存在任务错误或 panic 错误，`Cause()`
会返回第一个失败原因。
