# shutdown

Coordinate graceful shutdown hooks driven by OS signals.

```go
func StartServer() {
    lis, _ := net.Listen("tcp", "127.0.0.1:0")
    srv := grpc.NewServer()
    shutdown.Add(srv.GracefulStop)
    _ = srv.Serve(lis)
}

func main() {
    go StartServer()

    // Block until SIGINT/SIGTERM, then run hooks and return.
    shutdown.Wait(context.Background())
}
```

For tests or libraries that need an isolated registry, use `NewManager`:

```go
m := shutdown.NewManager(syscall.SIGUSR1)
m.Add(func() { /* ... */ })
sig := m.Wait(ctx) // nil when ctx is canceled
```

Hooks run in registration order. `Cleanup` is idempotent and also available
directly when you want to trigger shutdown without waiting for a signal.

Hook panics are recovered so later hooks still run. Use `WithTimeout(d)` to
abandon a slow hook and continue to the next one; use `WithName(name)` to label
timeout and panic log lines.
