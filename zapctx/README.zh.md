# zapctx

通过 `context.Context` 携带 `*zap.Logger` 和累积的结构化字段，让 handler
不必在每个函数签名中显式传递 logger，也能带上请求级数据写日志。

## 快速开始

```go
func init() {
    if err := zapctx.BetterDefault(); err != nil {
        log.Fatal(err)
    }
}

func handler(ctx context.Context) {
    zapctx.L(ctx).Info("hello") // 包含上游附加的所有字段
}

func main() {
    ctx := zapctx.ToContext(context.Background(), zap.L())
    zapctx.Extract(ctx).AddFields(zap.String("hello", "world"))
    handler(ctx)
}
```

除非请求被标记为 sampled，`zapctx.Sampled(ctx)` 会返回 no-op logger，
适合控制冗长的单请求日志。

## 中间件

```go
// gin
engine.Use(zapctxgin.Middleware(zap.L()))

// grpc
grpc.NewServer(grpc.UnaryInterceptor(
    zapctxgrpc.UnaryServerInterceptor(zap.L()),
))
```

框架集成位于子包中，因此核心 `zapctx` 无需编译 Gin 或 gRPC 也能使用：

```go
import (
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
    "github.com/liguangsheng/goost/zapctx/zapctxgrpc"
)
```

## OpenTelemetry trace 注入

```go
engine.Use(zapctxgin.Middleware(zap.L(), zapctxotel.TraceInject))
```

OpenTelemetry hook 位于子包中，因此核心 `zapctx` 无需编译 OpenTelemetry 也能使用：

```go
import "github.com/liguangsheng/goost/zapctx/zapctxotel"
```

`zapctxotel.TraceInject` 会读取 `trace.SpanContextFromContext(ctx)`，
并把 `trace.traceid` / `trace.spanid` / `trace.sampled` 字段添加到绑定的
logger；它也会转发 sample 标记，让 `zapctx.Sampled(ctx)` 可以控制冗长的
单请求日志。
