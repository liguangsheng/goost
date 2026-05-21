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
后续错误会被丢弃。
