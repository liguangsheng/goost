# Migration Guide

## v0.3.0

v0.3.0 moves optional logging integrations out of the core context packages.
Code that imports only `zapctx` or `slogctx` no longer compiles Gin, gRPC, or
OpenTelemetry dependencies.

### zapctx integrations

| Before | After |
| --- | --- |
| `zapctx.GinMiddleware` | `zapctxgin.Middleware` |
| `zapctx.PayloadGinMiddleware` | `zapctxgin.PayloadMiddleware` |
| `zapctx.WithMaxBody` | `zapctxgin.WithMaxBody` |
| `zapctx.WithSampling` | `zapctxgin.WithSampling` |
| `zapctx.WithSkipper` | `zapctxgin.WithSkipper` |
| `zapctx.UnaryServerInterceptor` | `zapctxgrpc.UnaryServerInterceptor` |
| `zapctx.PayloadUnaryServerInterceptor` | `zapctxgrpc.PayloadUnaryServerInterceptor` |
| `zapctx.GRPCWithBody` | `zapctxgrpc.GRPCWithBody` |
| `zapctx.GRPCWithSampling` | `zapctxgrpc.GRPCWithSampling` |
| `zapctx.GRPCWithSkipper` | `zapctxgrpc.GRPCWithSkipper` |
| `zapctx.OtelTraceInject` | `zapctxotel.TraceInject` |

```go
import (
    "github.com/liguangsheng/goost/zapctx"
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
    "github.com/liguangsheng/goost/zapctx/zapctxotel"
)

engine.Use(zapctxgin.Middleware(zap.L(), zapctxotel.TraceInject))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
zapctx.L(ctx).Info("handled")
```

`zapctxotel.OtelTraceInject` remains available as a migration alias for
`zapctxotel.TraceInject`.

### slogctx integrations

| Before | After |
| --- | --- |
| `slogctx.OtelTraceInject` | `slogctxotel.TraceInject` |

```go
import (
    "github.com/liguangsheng/goost/slogctx"
    "github.com/liguangsheng/goost/slogctx/slogctxotel"
)

ctx = slogctxotel.TraceInject(ctx)
slogctx.Sampled(ctx).Info("sampled request")
```

`slogctxotel.OtelTraceInject` remains available as a migration alias for
`slogctxotel.TraceInject`.
