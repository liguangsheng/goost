# zapctx

Carry a `*zap.Logger` (and accumulated structured fields) through
`context.Context` so handlers can log with request-scoped data without
threading the logger through every signature.

## Quickstart

```go
func init() {
    if err := zapctx.BetterDefault(); err != nil {
        log.Fatal(err)
    }
}

func handler(ctx context.Context) {
    zapctx.L(ctx).Info("hello") // includes any fields attached upstream
}

func main() {
    ctx := zapctx.ToContext(context.Background(), zap.L())
    zapctx.Extract(ctx).AddFields(zap.String("hello", "world"))
    handler(ctx)
}
```

`zapctx.Sampled(ctx)` returns a no-op logger unless the request is marked
sampled (see `OpenTraceInject`), useful for verbose per-request logs.

## Middleware

```go
// gin
engine.Use(zapctx.GinMiddleware(zap.L()))

// grpc
grpc.NewServer(grpc.UnaryInterceptor(
    zapctx.UnaryServerInterceptor(zap.L()),
))
```

## OpenTelemetry trace injection

```go
engine.Use(zapctx.GinMiddleware(zap.L(), zapctx.OtelTraceInject))
```

`OtelTraceInject` reads `trace.SpanContextFromContext(ctx)` and adds
`trace.traceid` / `trace.spanid` / `trace.sampled` fields to the bound
logger; it also forwards the sample flag so `zapctx.Sampled(ctx)` can
gate verbose per-request logs.
