# priorityqueue

`container/heap` is fast but verbose: every type needs five methods.
`priorityqueue` does that once and exposes a small queue API on top.

```go
q := priorityqueue.New(func(a, b Job) bool {
    return a.Priority < b.Priority // min-heap by Priority
})

q.Push(Job{Priority: 3})
q.Push(Job{Priority: 1})
q.Push(Job{Priority: 2})

top, _ := q.Pop() // {Priority: 1}
```

For a max-heap, flip the comparator: `return a.Priority > b.Priority`.

## API

```go
New(less func(a, b T) bool) *PriorityQueue[T]
NewWithCapacity(less func(a, b T) bool, n int) *PriorityQueue[T]
Push(v T)
Pop() (T, bool)        // false when empty
Peek() (T, bool)       // false when empty
Len() int
Clear()
Drain() []T            // pop everything into a slice in priority order
```

## Notes

- Not safe for concurrent use. Wrap with `sync.Mutex` if needed.
- `Drain` is O(n log n). `Clear` is O(1) (keeps the underlying slice).
- Use `NewWithCapacity(less, n)` if you know the queue size up front
  to avoid grow-and-copy; negative capacity is treated as zero.
