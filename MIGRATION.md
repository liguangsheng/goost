# Migration Guide

## Unreleased

The low-value public packages `bytesconv`, `itertools`, `redact`,
`slogctx/slogctxotel`, and `zapctx/zapctxotel` have been removed.
Use the standard library or local application helpers for those narrow cases.
The `zapctx/zapctxgin` and `zapctx/zapctxgrpc` integrations are now separate
modules at the same import paths.

- `bytesconv`: use ordinary `[]byte(s)` / `string(b)` conversions unless an
  application-specific unsafe helper is justified.
- `itertools`: use `slices`, loops, or local helpers close to the call site.
- `redact`: keep masking policy in the application layer where field rules are
  explicit.
- `slogctx/slogctxotel` and `zapctx/zapctxotel`: inject trace fields with an
  application-owned hook.
- `zapctx/zapctxgin` and `zapctx/zapctxgrpc`: keep using the same import paths;
  module-aware tooling will resolve them as separate optional modules.

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
| `zapctx.OtelTraceInject` | Removed after v0.3.0; use an application-owned hook. |

```go
import (
    "github.com/liguangsheng/goost/zapctx"
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
)

engine.Use(zapctxgin.Middleware(zap.L()))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
zapctx.L(ctx).Info("handled")
```

If you still need trace IDs on request loggers, pass a local hook to
`zapctxgin.Middleware` that reads your tracing context and appends fields to
`zapctx.Extract(ctx)`.

### slogctx integrations

| Before | After |
| --- | --- |
| `slogctx.OtelTraceInject` | Removed after v0.3.0; use an application-owned hook. |

```go
import (
    "github.com/liguangsheng/goost/slogctx"
)

// Add trace attrs from your application's tracing layer, if needed.
slogctx.Sampled(ctx).Info("sampled request")
```

If you still need trace IDs on context-bound slog loggers, enrich
`slogctx.Extract(ctx)` in application code.
