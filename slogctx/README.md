# slogctx

`zapctx` for `log/slog`. Carry a `*slog.Logger` and accumulated attrs
through `context.Context`.

```go
ctx := slogctx.ToContext(context.Background(), slog.Default())
slogctx.Extract(ctx).AddAttrs(slog.String("request_id", id))
slogctx.L(ctx).Info("hello") // includes request_id
```

`slogctx.Sampled(ctx)` returns a no-op logger unless the request is
marked sampled:

```go
sc := slogctx.Extract(ctx)
sc.Sampled = true
slogctx.Sampled(ctx).Debug("verbose trace") // emitted with accumulated attrs
```

## Shared Model with zapctx

`slogctx` mirrors the core `zapctx` concepts while using standard library
`log/slog` types:

| Concept | slogctx | zapctx equivalent |
| --- | --- | --- |
| Attach logger | `ToContext(ctx, *slog.Logger)` | `zapctx.ToContext(ctx, *zap.Logger)` |
| Extract state | `Extract(ctx)` | `zapctx.Extract(ctx)` |
| Add request data | `AddAttrs(...)` | `AddFields(...)` |
| Log normally | `L(ctx)` | `zapctx.L(ctx)` / `zapctx.S(ctx)` |
| Sample-gated logs | `Sampled(ctx)` | `zapctx.Sampled(ctx)` |

There are no framework integrations in `slogctx`; integrations should stay in
nested modules if they are added later.
