# slogctx

面向 `log/slog` 的 `zapctx`。通过 `context.Context` 携带 `*slog.Logger`
和累积的 attrs。

```go
ctx := slogctx.ToContext(context.Background(), slog.Default())
slogctx.Extract(ctx).AddAttrs(slog.String("request_id", id))
slogctx.L(ctx).Info("hello") // 包含 request_id
```

除非请求被标记为 sampled，`slogctx.Sampled(ctx)` 会返回 no-op logger：

```go
sc := slogctx.Extract(ctx)
sc.Sampled = true
slogctx.Sampled(ctx).Debug("verbose trace") // 会带上累积 attrs 输出
```

## 与 zapctx 的共享模型

`slogctx` 复用核心 `zapctx` 概念，但使用标准库 `log/slog` 类型：

| Concept | slogctx | zapctx equivalent |
| --- | --- | --- |
| Attach logger | `ToContext(ctx, *slog.Logger)` | `zapctx.ToContext(ctx, *zap.Logger)` |
| Extract state | `Extract(ctx)` | `zapctx.Extract(ctx)` |
| Add request data | `AddAttrs(...)` | `AddFields(...)` |
| Log normally | `L(ctx)` | `zapctx.L(ctx)` / `zapctx.S(ctx)` |
| Sample-gated logs | `Sampled(ctx)` | `zapctx.Sampled(ctx)` |

`slogctx` 不包含 framework integrations；如果未来增加集成，也应留在 nested modules 中。
