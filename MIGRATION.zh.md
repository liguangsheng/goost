# 迁移指南

## v0.3.0

v0.3.0 将可选日志集成从核心 context 包中移出。只导入 `zapctx` 或
`slogctx` 的代码不再需要编译 Gin、gRPC 或 OpenTelemetry 依赖。

### zapctx 集成

| 之前 | 之后 |
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

`zapctxotel.OtelTraceInject` 仍作为 `zapctxotel.TraceInject` 的迁移别名保留。

### slogctx 集成

| 之前 | 之后 |
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

`slogctxotel.OtelTraceInject` 仍作为 `slogctxotel.TraceInject` 的迁移别名保留。
