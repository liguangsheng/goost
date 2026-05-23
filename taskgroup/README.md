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
exit early; subsequent errors are dropped. A task panic follows the same path:
it is recovered, converted to an error with a `taskgroup: panic:` prefix,
cancels siblings, and is returned by `Wait` if it is the first failure. `Wait`
cancels the context before returning even on success, and `Cause()` reports the
first task error or recovered panic when there was one.

Use `Results[T]` when each task returns a value:

```go
g := taskgroup.NewResults[string](ctx).WithLimit(4)
for _, item := range items {
    item := item
    g.Run(func(ctx context.Context) (string, error) {
        return fetch(ctx, item)
    })
}
values, err := g.Wait() // values are in completion order
```

Like `Group`, `Results[T]` cancels its context on the first task error or
recovered panic, returns values that completed before cancellation, and
`Cause()` reports the first task error or panic error when there was one.
