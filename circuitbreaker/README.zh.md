# circuitbreaker

三态熔断器（closed / open / half-open），用于快速短路对故障依赖的调用。

```go
b := circuitbreaker.New(circuitbreaker.Config{
    FailureThreshold:  5,
    CooldownPeriod:    30 * time.Second,
    HalfOpenSuccesses: 1,
})

err := b.Do(ctx, func(ctx context.Context) error {
    return callDownstream(ctx)
})
if errors.Is(err, circuitbreaker.ErrOpen) {
    return fallback()
}

snap := b.Snapshot()
metrics.RecordBreakerState(snap.State.String(), snap.Failures, snap.CooldownRemaining)
```

### 行为

- **closed**：所有调用放行；连续失败达到 `FailureThreshold` 后熔断。
- **open**：调用立即返回 `ErrOpen`；经过 `CooldownPeriod` 后进入 half-open。
- **half-open**：同一时间只允许一个探测调用。连续成功
  `HalfOpenSuccesses` 次后关闭熔断器；任何失败都会重新打开。

### 定制

- `IsFailure` 可把 `context.Canceled` 这类预期错误排除在失败计数之外。
- `OnStateChange` 会在每次状态切换时触发，可用于指标或日志。
- `Snapshot` 会返回当前状态、失败计数、打开时间和剩余冷却时间，可用于指标或日志。
- `Now` 可替换时钟，方便确定性测试；可配合 [`clock.Mock`](../clock)：

```go
m := clock.NewMock(time.Unix(0, 0))
b := circuitbreaker.New(circuitbreaker.Config{Now: m.Now, ...})
m.Advance(time.Minute) // 推进冷却时间
```
