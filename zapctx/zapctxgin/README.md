# zapctxgin

Gin middleware for `github.com/liguangsheng/goost/zapctx`.

This is a nested module so Gin and its transitive dependencies stay out of the
root `goost` module. Import it only in applications that use Gin.

## Install

```sh
go get github.com/liguangsheng/goost/zapctx/zapctxgin
```

## Usage

```go
engine.Use(zapctxgin.Middleware(zap.L()))
engine.Use(zapctxgin.PayloadMiddleware(zap.L(), zapctxgin.WithMaxBody(1024)))
```

`Middleware` attaches a request-scoped zap logger to the request context.
`PayloadMiddleware` logs bounded HTTP payload summaries with skipper and
sampling options.

Payload logging can record request and response bodies. Avoid enabling it on
routes that may carry secrets or personal data unless those fields are
redacted before logging.
