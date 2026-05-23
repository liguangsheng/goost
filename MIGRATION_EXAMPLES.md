# Migration Examples

Small before/after examples for common migration paths.

The integration examples below are compile-backed by fixtures under
`testdata/migration` and checked by the root smoke tests.

## zapctx Gin Integration

Before v0.3.0, Gin middleware lived on the core `zapctx` package. Import the
nested module instead:

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

## zapctx gRPC Integration

Before v0.3.0, gRPC interceptors lived on the core `zapctx` package. Import
the nested module instead:

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

## Removed Narrow Helpers

For removed narrow helpers, prefer standard-library or application-owned code:

```go
s := string(b)
b := []byte(s)
```

Keep redaction policy in the application layer where field rules are explicit.
