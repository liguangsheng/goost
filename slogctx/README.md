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

```go
otelInjected := slogctxotel.TraceInject(ctx)
// adds trace.traceid / trace.spanid / trace.sampled attrs when a valid
// OpenTelemetry span is in ctx
```

The OpenTelemetry hook lives in `github.com/liguangsheng/goost/slogctx/slogctxotel`
so core `slogctx` stays usable without compiling OpenTelemetry.
