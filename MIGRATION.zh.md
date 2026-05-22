# 迁移指南

## Unreleased

已移除低价值公开包：`bytesconv`、`itertools`、`redact`、
`slogctx/slogctxotel` 和 `zapctx/zapctxotel`。这些窄场景请改用标准库或
应用内 helper。`zapctx/zapctxgin` 和 `zapctx/zapctxgrpc` 现在是相同 import
path 下的独立 module。
`rotatingwriter` 现在会用 `0750` 权限创建新的日志目录，并在 umask 作用前用
`0600` 权限创建新的日志文件和备份文件。已有文件不会被 chmod。如果部署环境
要求 group-readable logs，请在 `rotatingwriter` 创建后由外部设置目录或文件权限，
或者提前用期望权限创建日志路径。

- `bytesconv`：除非应用明确需要 unsafe helper，否则使用普通
  `[]byte(s)` / `string(b)` 转换。
- `itertools`：使用 `slices`、直接循环，或靠近调用点的本地 helper。
- `redact`：脱敏策略保留在应用层，让字段规则更明确。
- `slogctx/slogctxotel` 和 `zapctx/zapctxotel`：如需 trace 字段，请在应用内
  hook 中注入。
- `zapctx/zapctxgin` 和 `zapctx/zapctxgrpc`：继续使用相同 import path；
  module-aware tooling 会把它们解析为独立可选 module。

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
| `zapctx.OtelTraceInject` | v0.3.0 后已移除；请使用应用内 hook。 |

```go
import (
    "github.com/liguangsheng/goost/zapctx"
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
)

engine.Use(zapctxgin.Middleware(zap.L()))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
zapctx.L(ctx).Info("handled")
```

如果仍需在请求 logger 中加入 trace ID，请给 `zapctxgin.Middleware` 传入本地
hook，在其中读取 tracing context，并向 `zapctx.Extract(ctx)` 追加字段。

### slogctx 集成

| 之前 | 之后 |
| --- | --- |
| `slogctx.OtelTraceInject` | v0.3.0 后已移除；请使用应用内 hook。 |

```go
import (
    "github.com/liguangsheng/goost/slogctx"
)

// 如有需要，从应用的 tracing 层添加 trace attrs。
slogctx.Sampled(ctx).Info("sampled request")
```

如果仍需在 context-bound slog logger 中加入 trace ID，请在应用代码中 enrich
`slogctx.Extract(ctx)`。
