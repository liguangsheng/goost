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

## Shared Model with slogctx

`zapctx` and `slogctx` use the same context logging model:

| Concept | zapctx | slogctx equivalent |
| --- | --- | --- |
| Attach logger | `ToContext(ctx, *zap.Logger)` | `slogctx.ToContext(ctx, *slog.Logger)` |
| Extract state | `Extract(ctx)` | `slogctx.Extract(ctx)` |
| Add request data | `AddFields(...)` | `AddAttrs(...)` |
| Log normally | `L(ctx)` / `S(ctx)` | `slogctx.L(ctx)` |
| Sample-gated logs | `Sampled(ctx)` | `slogctx.Sampled(ctx)` |

Framework integrations stay in nested modules; the core package only carries
logger state through `context.Context`.

## Middleware

```go
// gin
engine.Use(zapctxgin.Middleware(zap.L()))

// grpc
grpc.NewServer(grpc.UnaryInterceptor(
    zapctxgrpc.UnaryServerInterceptor(zap.L()),
))
```

The framework integrations live in optional modules so core `zapctx` stays
usable without compiling Gin or gRPC:

```go
import (
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
    "github.com/liguangsheng/goost/zapctx/zapctxgrpc"
)
```
