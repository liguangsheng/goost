# pool

有界 goroutine 池，支持可选排队。

```go
p, err := pool.NewPool(64, 0, 4) // 大小 64，无队列，预启动 4 个 worker
if err != nil { /* ... */ }
defer p.Close()

if err := p.Schedule(func() { doWork() }); err != nil {
    // ErrPoolClosed
}

if err := p.ScheduleTimeout(50*time.Millisecond, func() { doWork() }); err != nil {
    // 没有 worker 在期限内可用时返回 ErrScheduleTimeout
}
```

- `Schedule` 会阻塞直到有可用 slot。
- `ScheduleTimeout` 会在截止时间后放弃。
- 任务 panic 会被恢复，并用 zap 全局 logger 记录；池仍然可用。
- `Close` 停止接受新任务，并等待正在执行的 worker 退出。
