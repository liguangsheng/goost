# priorityqueue

`container/heap` 很快，但样板代码多：每个类型都要实现五个方法。
`priorityqueue` 只做一次这件事，并在上层提供一个小型队列 API。

```go
q := priorityqueue.New(func(a, b Job) bool {
    return a.Priority < b.Priority // 按 Priority 的最小堆
})

q.Push(Job{Priority: 3})
q.Push(Job{Priority: 1})
q.Push(Job{Priority: 2})

top, _ := q.Pop() // {Priority: 1}
```

最大堆只需反转比较器：`return a.Priority > b.Priority`。

## API

```go
Push(v T)
Pop() (T, bool)        // 空时 false
Peek() (T, bool)       // 空时 false
Len() int
Clear()
Drain() []T            // 按优先级顺序弹出全部元素
```

## 说明

- 非并发安全。需要并发使用时请自行包 `sync.Mutex`。
- `Drain` 是 O(n log n)。`Clear` 是 O(1)，会保留底层 slice。
- 如果预先知道队列大小，使用 `NewWithCapacity(less, n)` 可避免扩容和复制。
