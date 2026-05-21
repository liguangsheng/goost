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
sampled, useful for verbose per-request logs.

## Middleware

```go
// gin
engine.Use(zapctxgin.Middleware(zap.L()))

// grpc
grpc.NewServer(grpc.UnaryInterceptor(
    zapctxgrpc.UnaryServerInterceptor(zap.L()),
))
```

The framework integrations live in subpackages so core `zapctx` stays usable
without compiling Gin or gRPC:

```go
import (
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
    "github.com/liguangsheng/goost/zapctx/zapctxgrpc"
)
```

## OpenTelemetry trace injection

```go
engine.Use(zapctxgin.Middleware(zap.L(), zapctxotel.TraceInject))
```

The OpenTelemetry hook lives in a subpackage so core `zapctx` stays usable
without compiling OpenTelemetry:

```go
import "github.com/liguangsheng/goost/zapctx/zapctxotel"
```

`zapctxotel.TraceInject` reads `trace.SpanContextFromContext(ctx)` and adds
`trace.traceid` / `trace.spanid` / `trace.sampled` fields to the bound
logger; it also forwards the sample flag so `zapctx.Sampled(ctx)` can
gate verbose per-request logs.
