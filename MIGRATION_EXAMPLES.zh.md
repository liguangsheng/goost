# 迁移示例

常见迁移路径的小型 before/after 示例。

下面的 integration 示例由 `testdata/migration` 下的 fixture 做编译校验，并由
root smoke tests 检查。

## zapctx Gin 集成

v0.3.0 之前，Gin middleware 位于核心 `zapctx` 包。现在请导入 nested module：

```go
import (
    "github.com/liguangsheng/goost/zapctx"
    "github.com/liguangsheng/goost/zapctx/zapctxgin"
    "go.uber.org/zap"
)

engine.Use(zapctxgin.Middleware(zap.L()))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
zapctx.L(ctx).Info("handled")
```

## zapctx gRPC 集成

v0.3.0 之前，gRPC interceptors 位于核心 `zapctx` 包。现在请导入 nested module：

```go
import (
    "github.com/liguangsheng/goost/zapctx/zapctxgrpc"
    "go.uber.org/zap"
    "google.golang.org/grpc"
)

server := grpc.NewServer(
    grpc.UnaryInterceptor(zapctxgrpc.UnaryServerInterceptor(zap.L())),
)
```

## 已移除的窄场景 helper

对于已移除的窄场景 helper，优先使用标准库或应用内代码：

```go
s := string(b)
b := []byte(s)
```

脱敏策略应保留在应用层，让字段规则保持明确。
