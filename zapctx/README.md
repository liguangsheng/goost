# zapctx
zap logger from context.

# example

## simple
```go
package main

func init() {
	zapctx.BetterDefault()
}

func SomeFunction(ctx context.Context) {
    logger := zapctx.L(ctx)
    
	logger.Info("some log") 
	// {"level":"info","time":"2019-10-11T10:28:18.492+0800","caller":"_playground/main.go:64","msg":"some log","hello":"world"}
	
	logger.Info("some log") 
	// {"level":"info","time":"2019-10-11T10:28:18.492+0800","caller":"_playground/main.go:64","msg":"some log","hello":"world"}
	
	sampledLogger := zapctx.Sampled(ctx)
	sampledLogger.Info("some log") 
	// nothing, because Sampled is false
}

func main() {
	originContext := context.Background()
	newCtx := zapctx.ToContext(originContext, zap.L())
	_ctx := zapctx.Extract(newCtx)
	_ctx.AddFields(zap.String("hello", "world"))
	_ctx.Sampled = false

	SomeFunction(newCtx)
}
```

## gin middleware

```go
engine := gin.Default()
engine.Use(zapctx.GinMiddleware(zap.L(), zapctx.OpenTraceInject))
```

## grpc middleware
```go
grpc.NewServer(
	grpc.StatsHandler(statsHandler),
	grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		zapctx.UnaryServerInterceptor(zap.L(), zapctx.OpenTraceInject),
	)))
```

# TODO

- payload middleware