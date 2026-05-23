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
- 任务 panic 会被恢复，池仍然可用；可用 `WithPanicHandler` 观察恢复到的 panic 值。
- `Close` 停止接受新任务，并等待 worker 完成已经接受的任务，包括已排队任务。
- `Stats` 会报告 worker、容量、正在执行的任务、排队任务、队列容量、已完成任务、panic 数量和关闭状态。

## Shutdown 语义

`Close` 是幂等的。第一次调用会关闭新提交入口、关闭内部 work queue，并等待 worker
完成所有已经接受的任务。close 开始后，`Schedule` 和 `ScheduleTimeout` 都会返回
`ErrPoolClosed`。close 前已经接受的任务仍会执行；close 后提交的任务不会执行。

`Stats()` 返回调用时刻的只读 snapshot。`Completed` 和 `Panics` 是累计 counter。
`Workers`、`InFlight`、`Queued` 和 `Closed` 是调用时刻的当前值。`Capacity` 和
`QueueCapacity` 是复制到 snapshot 中的配置值，便于作为 metrics label 或日志字段。
