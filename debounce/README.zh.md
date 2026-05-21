# debounce

把一连串 `Trigger(v)` 调用合并成一次从 `C()` 发出的事件。每次
`Trigger` 后，debouncer 会等待一个没有新触发的安静窗口，然后转发
**最新**值。中间值会被丢弃。

```go
d := debounce.New[string](300 * time.Millisecond)
defer d.Stop()

go func() {
    for v := range d.C() {
        applyConfig(v)
    }
}()

// 例如保存文件时连续触发的 file watcher：
d.Trigger(load("config.yaml"))
d.Trigger(load("config.yaml"))
d.Trigger(load("config.yaml"))
// 300ms 后，applyConfig 只会用最新值执行一次
```

## 不是限流器

`ratelimit` 限制调用速率；`debounce` 则延迟发送，直到输入流安静下来。
如果想要“每秒最多 5 次”，用 `ratelimit`。如果想要“等到 X 时间内没有新事件，
然后执行”，用 `debounce`。

## 背压

输出通道容量为 1，并采用 **latest-wins** 策略：如果消费者落后，
新的发送会覆盖缓冲区里的旧值。消费者最终一定能看到**最新**发送值，
但可能错过中间值。

## 便于测试

注入 `clock.Clock` 可让测试中的时间确定可控：

```go
m := clock.NewMock(time.Unix(0, 0))
d := debounce.New[int](100 * time.Millisecond).WithClock(m)

d.Trigger(42)
m.Advance(100 * time.Millisecond) // 此时触发发送
v := <-d.C()
```
