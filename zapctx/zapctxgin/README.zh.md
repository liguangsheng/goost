# zapctxgin

用于 `github.com/liguangsheng/goost/zapctx` 的 Gin middleware。

这是一个 nested module，因此 Gin 及其传递依赖不会进入 root `goost` module。只有使用 Gin 的应用才需要导入它。

## 安装

```sh
go get github.com/liguangsheng/goost/zapctx/zapctxgin
```

## 用法

```go
engine.Use(zapctxgin.Middleware(zap.L()))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
```

`Middleware` 会把请求级 zap logger 放入 request context。`PayloadMiddleware` 会记录有大小限制的 HTTP payload 摘要，并支持 skipper 与 sampling options。

Payload logging 可能记录 request 和 response body。对可能包含 secret 或个人数据的路由，除非这些字段会在写日志前被脱敏，否则不要启用。
