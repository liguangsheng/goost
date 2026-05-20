# taskgroup

A small alternative to `golang.org/x/sync/errgroup` with two extras:

- a concurrency limit
- task panics are recovered and surfaced as errors

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

The first non-nil error cancels the group's context so sibling tasks can
exit early; subsequent errors are dropped.
