# slogctx

面向 `log/slog` 的 `zapctx`。通过 `context.Context` 携带 `*slog.Logger`
和累积的 attrs。

```go
ctx := slogctx.ToContext(context.Background(), slog.Default())
slogctx.Extract(ctx).AddAttrs(slog.String("request_id", id))
slogctx.L(ctx).Info("hello") // 包含 request_id
```

除非请求被标记为 sampled，`slogctx.Sampled(ctx)` 会返回 no-op logger。
