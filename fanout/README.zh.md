# fanout

进程内广播器：每次 `Publish(v)` 都会发送给所有当前订阅者。适合事件总线、
日志 tee、配置重载信号，以及任何需要在同一进程内把一个生产者输出分发给
N 个独立消费者的场景。

```go
b := fanout.New[Event]().Buffer(64).Build()

sub := b.Subscribe()
defer sub.Close()

go func() {
    for ev := range sub.C() {
        handle(ev)
    }
}()

b.Publish(Event{...})
```

## 背压：丢弃，不阻塞

如果广播器在最慢订阅者上阻塞 `Publish`，一个卡住的消费者就能停住整个系统。
`fanout` 选择在单个订阅者缓冲区已满时丢弃消息并计数。发布者保持快速；
慢消费者只会自己落后。

```go
sub.Drops()      // 这个 sub 错过的消息数
b.Stats().Drops  // 所有 sub 的累计丢弃数（包括仍存活和已关闭的 sub）
```

`Stats()` 也会报告当前订阅者数、每个订阅者的配置缓冲区大小、活跃订阅者中
当前排队的消息总数，以及 broadcaster 是否已关闭。

返回值是调用时刻的只读 snapshot。`Publishes` 和 `Drops` 是累计 counter。
`Subscribers`、`Queued` 和 `Closed` 是调用时刻的当前值。`Buffer` 是复制到
snapshot 中的配置值，表示每订阅者缓冲区大小，便于作为 metrics label 或日志字段。

如果不能容忍丢弃，请把 `Buffer` 设得足够大，以吸收预期突发量。

## 生命周期

- `Sub.Close()` 取消订阅；`Sub.C()` 的通道会关闭，因此
  `for v := range sub.C()` 会干净退出。
- `Broadcaster.Close()` 关闭所有订阅者通道，并让之后的 `Publish` 变成 no-op。
- 两种 `Close` 调用都是幂等的。
- 在 `Broadcaster.Close` 后创建的订阅会返回一个已经关闭的通道，
  因此可以无条件接入 `for range` 循环。

## 说明

- 订阅者只能看到它们调用 `Subscribe` **之后**发布的值；不会回放历史值。
- 每个订阅者的缓冲区大小在 `Subscribe` 时从 Builder 的 `Buffer(n)` 读取。
  之后再改 builder 不会影响已有订阅者。
