# slogctx

面向 `log/slog` 的 `zapctx`。通过 `context.Context` 携带 `*slog.Logger`
和累积的 attrs。

```go
ctx := slogctx.ToContext(context.Background(), slog.Default())
slogctx.Extract(ctx).AddAttrs(slog.String("request_id", id))
slogctx.L(ctx).Info("hello") // 包含 request_id
```

除非请求被标记为 sampled，`slogctx.Sampled(ctx)` 会返回 no-op logger。

```go
otelInjected := slogctxotel.TraceInject(ctx)
// 当 ctx 中有有效 OpenTelemetry span 时，添加
// trace.traceid / trace.spanid / trace.sampled attrs
```

OpenTelemetry hook 位于 `github.com/liguangsheng/goost/slogctx/slogctxotel`，
因此核心 `slogctx` 无需编译 OpenTelemetry 也能使用。
