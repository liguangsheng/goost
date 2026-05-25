# zapctxgrpc

gRPC interceptors for `github.com/liguangsheng/goost/zapctx`.

This is a nested module so gRPC and its transitive dependencies stay out of the
root `goost` module. Import it only in applications that use gRPC.

## Install

```sh
go get github.com/liguangsheng/goost/zapctx/zapctxgrpc
```

## Usage

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(zapctxgrpc.UnaryServerInterceptor(zap.L())),
)
```

`UnaryServerInterceptor` attaches a request-scoped zap logger to the RPC
context. Payload interceptors are available when bounded request/response body
logging is needed. Use `GRPCWithBody(true)` to include messages and
`GRPCWithMaxBody(n)` to cap the formatted request and response strings.

Payload logging can record request and response bodies. Avoid enabling it on
RPCs that may carry secrets or personal data unless those fields are redacted
before logging.
