# zapctxgrpc

用于 `github.com/liguangsheng/goost/zapctx` 的 gRPC interceptors。

这是一个 nested module，因此 gRPC 及其传递依赖不会进入 root `goost` module。只有使用 gRPC 的应用才需要导入它。

## 安装

```sh
go get github.com/liguangsheng/goost/zapctx/zapctxgrpc
```

## 用法

```go
server := grpc.NewServer(
    grpc.UnaryInterceptor(zapctxgrpc.UnaryServerInterceptor(zap.L())),
)
```

`UnaryServerInterceptor` 会把请求级 zap logger 放入 RPC context。需要记录有大小限制的 request/response body 时，可以使用 payload interceptors。使用 `GRPCWithBody(true)` 启用 message 记录，并用 `GRPCWithMaxBody(n)` 限制格式化后的 request 和 response 字符串长度。

Payload logging 可能记录 request 和 response body。对可能包含 secret 或个人数据的 RPC 启用前，请先阅读 [../../SECURITY.zh.md](../../SECURITY.zh.md)。
