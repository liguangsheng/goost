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
sig := m.Wait(ctx) // ctx 取消时返回 nil
```

hook 会按注册顺序运行。`Cleanup` 是幂等的；如果想不等待信号而直接触发关闭，
也可以直接调用它。

hook panic 会被恢复，因此后续 hook 仍会继续运行。可用 `WithTimeout(d)` 放弃慢
hook 并继续执行下一个；可用 `WithName(name)` 给 timeout 和 panic 日志加标签。
