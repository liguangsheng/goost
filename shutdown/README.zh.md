# shutdown

协调由操作系统信号触发的优雅关闭 hook。

```go
func StartServer() {
    lis, _ := net.Listen("tcp", "127.0.0.1:0")
    srv := grpc.NewServer()
    shutdown.Add(srv.GracefulStop)
    _ = srv.Serve(lis)
}

func main() {
    go StartServer()

    // 阻塞直到 SIGINT/SIGTERM，然后运行 hook 并返回。
    shutdown.Wait(context.Background())
}
```

测试或库代码如果需要隔离的 registry，可使用 `NewManager`：

```go
m := shutdown.NewManager(syscall.SIGUSR1)
m.Add(func() { /* ... */ })
m.Wait(ctx)
```

hook panic 会被恢复，因此后续 hook 仍会继续运行。
