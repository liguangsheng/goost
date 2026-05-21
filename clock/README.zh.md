# clock

提供一个 `Clock` 接口，以及 `Real` 和 `Mock` 两种实现。适合需要控制时间、
但不想在测试中真实 sleep 的代码。

```go
import "github.com/liguangsheng/goost/clock"

func usesClock(c clock.Clock) {
    deadline := c.Now().Add(time.Second)
    <-c.After(time.Until(deadline))
    // ...
}

// 生产环境
usesClock(clock.Real())

// 测试
m := clock.NewMock(time.Unix(0, 0))
go usesClock(m)
m.Advance(time.Second) // 所有 deadline <= now 的 After/Sleep 等待者都会触发
```

`Clock` 覆盖五个调度原语，都可以由 `Mock` 驱动：

| `Clock` 方法 | 标准库对应项 |
| --- | --- |
| `Now()` | `time.Now` |
| `After(d)` | `time.After` |
| `Sleep(d)` | `time.Sleep` |
| `AfterFunc(d, fn)` | `time.AfterFunc` |
| `NewTicker(d)` | `time.NewTicker` |

由 `Mock` 驱动的 ticker 只有在调用 `Advance` 时才 tick。错过的 tick 会丢弃
（通道容量为 1），与 `time.Ticker` 一致：

```go
m := clock.NewMock(time.Unix(0, 0))
tk := m.NewTicker(time.Second)
defer tk.Stop()

m.Advance(time.Second)
fmt.Println(<-tk.C()) // 1970-01-01 00:00:01 UTC
```

`Mock.Now` 的签名是 `func() time.Time`，这正是 `ratelimit.Bucket.SetClock`、
`circuitbreaker.Config.Now` 等字段已经接受的形式。可以不改这些包的 API，
直接用 mock 驱动它们：

```go
m := clock.NewMock(time.Unix(0, 0))
b := ratelimit.NewBucket(10, 1)
b.SetClock(m.Now)
```
