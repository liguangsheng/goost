# slogctx

`zapctx` for `log/slog`. Carry a `*slog.Logger` and accumulated attrs
through `context.Context`.

```go
ctx := slogctx.ToContext(context.Background(), slog.Default())
slogctx.Extract(ctx).AddAttrs(slog.String("request_id", id))
slogctx.L(ctx).Info("hello") // includes request_id
```

`slogctx.Sampled(ctx)` returns a no-op logger unless the request is
marked sampled.
