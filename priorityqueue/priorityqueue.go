// Package priorityqueue is a thin, generic façade over container/heap.
// container/heap is fast, but requires the caller to implement the
// five-method heap.Interface on every type. This package does that
// once and exposes a small queue API:
//
//	q := priorityqueue.New[Job](func(a, b Job) bool {
//	    return a.Priority < b.Priority // min-heap by Priority
//	})
//	q.Push(Job{Priority: 3})
//	q.Push(Job{Priority: 1})
//	top, _ := q.Pop() // {Priority: 1}
//
// The queue is NOT safe for concurrent use; wrap with sync.Mutex if
// multiple goroutines need access.
package priorityqueue

import "container/heap"

// PriorityQueue orders items by less: less(a, b) == true means a
// should come out before b. A min-heap is the typical idiom.
type PriorityQueue[T any] struct {
	h *heapImpl[T]
}

// New returns an empty PriorityQueue.
func New[T any](less func(a, b T) bool) *PriorityQueue[T] {
	if less == nil {
		panic("priorityqueue: less func must not be nil")
	}
	return &PriorityQueue[T]{h: &heapImpl[T]{less: less}}
}

// NewWithCapacity preallocates the underlying storage.
func NewWithCapacity[T any](less func(a, b T) bool, cap int) *PriorityQueue[T] {
	if less == nil {
		panic("priorityqueue: less func must not be nil")
	}
	if cap < 0 {
		cap = 0
	}
	return &PriorityQueue[T]{h: &heapImpl[T]{
		less:  less,
		items: make([]T, 0, cap),
	}}
}

// Push adds v to the queue.
func (q *PriorityQueue[T]) Push(v T) { heap.Push(q.h, v) }

// Pop removes and returns the highest-priority item. ok is false if
// the queue is empty.
func (q *PriorityQueue[T]) Pop() (v T, ok bool) {
	if len(q.h.items) == 0 {
		return v, false
	}
	return heap.Pop(q.h).(T), true
}

// Peek returns the highest-priority item without removing it. ok is
// false if the queue is empty.
func (q *PriorityQueue[T]) Peek() (v T, ok bool) {
	if len(q.h.items) == 0 {
		return v, false
	}
	return q.h.items[0], true
}

// Len returns the number of items currently queued.
func (q *PriorityQueue[T]) Len() int { return len(q.h.items) }

// Clear removes every item. The underlying capacity is preserved.
func (q *PriorityQueue[T]) Clear() { q.h.items = q.h.items[:0] }

// Drain pops every item into a slice in priority order and empties
// the queue.
func (q *PriorityQueue[T]) Drain() []T {
	out := make([]T, 0, q.Len())
	for q.Len() > 0 {
		v, _ := q.Pop()
		out = append(out, v)
	}
	return out
}

// heapImpl satisfies heap.Interface for PriorityQueue[T].
type heapImpl[T any] struct {
	items []T
	less  func(a, b T) bool
}

func (h *heapImpl[T]) Len() int           { return len(h.items) }
func (h *heapImpl[T]) Less(i, j int) bool { return h.less(h.items[i], h.items[j]) }
func (h *heapImpl[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }

func (h *heapImpl[T]) Push(x any) {
	h.items = append(h.items, x.(T))
}

func (h *heapImpl[T]) Pop() any {
	old := h.items
	n := len(old)
	x := old[n-1]
	h.items = old[:n-1]
	return x
}
