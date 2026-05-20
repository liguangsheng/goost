// Package fanout provides an in-process broadcaster: every value
// published is delivered to all current subscribers.
//
// The design choice that matters most is the backpressure policy. A
// fanout primitive that blocks Publish on the slowest subscriber lets
// one stuck consumer halt the whole system. fanout instead drops on
// full per-subscriber buffer and counts the drops; producers stay
// fast, slow consumers fall behind on their own.
//
// Subscribers see values published AFTER their Subscribe call. There
// is no replay of past messages.
//
// Example:
//
//	b := fanout.New[Event]().Buffer(64).Build()
//
//	sub := b.Subscribe()
//	defer sub.Close()
//
//	go func() {
//	    for ev := range sub.C() {
//	        handle(ev)
//	    }
//	}()
//
//	b.Publish(Event{...})
package fanout

import (
	"sync"
	"sync/atomic"
)

// Builder configures a Broadcaster. Use New to obtain one.
type Builder[T any] struct {
	buf int
}

// New starts a Builder. Default per-subscriber buffer is 16.
func New[T any]() *Builder[T] {
	return &Builder[T]{buf: 16}
}

// Buffer sets the per-subscriber channel buffer for new subscribers.
// Existing subscribers keep their original buffer.
func (b *Builder[T]) Buffer(n int) *Builder[T] {
	if n > 0 {
		b.buf = n
	}
	return b
}

// Build returns a ready Broadcaster.
func (b *Builder[T]) Build() *Broadcaster[T] {
	return &Broadcaster[T]{
		subs: make(map[*Sub[T]]struct{}),
		buf:  b.buf,
	}
}

// Broadcaster delivers each Publish value to every subscriber.
type Broadcaster[T any] struct {
	mu     sync.RWMutex
	subs   map[*Sub[T]]struct{}
	buf    int
	closed bool

	publishCount atomic.Int64
	dropCount    atomic.Int64
}

// Sub is a single subscription handle. Read from C(); call Close to
// unsubscribe.
type Sub[T any] struct {
	ch     chan T
	parent *Broadcaster[T]
	drops  atomic.Int64
}

// Subscribe registers a new subscriber. Values published after
// Subscribe returns are delivered to the returned subscription's
// channel; older values are not replayed.
//
// If the Broadcaster has been closed, the subscription's channel is
// returned already-closed.
func (b *Broadcaster[T]) Subscribe() *Sub[T] {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := &Sub[T]{
		ch:     make(chan T, b.buf),
		parent: b,
	}
	if b.closed {
		close(s.ch)
		return s
	}
	b.subs[s] = struct{}{}
	return s
}

// Publish sends v to every current subscriber. A subscriber whose
// buffer is full drops the message and increments its drop counter.
// Publish never blocks on a slow subscriber.
func (b *Broadcaster[T]) Publish(v T) {
	b.publishCount.Add(1)
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	for s := range b.subs {
		select {
		case s.ch <- v:
		default:
			s.drops.Add(1)
			b.dropCount.Add(1)
		}
	}
}

// Close terminates the Broadcaster. Every subscriber's channel is
// closed and future Publish calls become no-ops. Idempotent.
func (b *Broadcaster[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for s := range b.subs {
		close(s.ch)
	}
	b.subs = nil
}

// Len returns the number of active subscribers.
func (b *Broadcaster[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

// Stats is a snapshot of Broadcaster counters.
type Stats struct {
	Publishes   int64
	Drops       int64
	Subscribers int
}

// Stats returns a counter snapshot. Drops is the aggregate across all
// subscribers, current and past.
func (b *Broadcaster[T]) Stats() Stats {
	b.mu.RLock()
	n := len(b.subs)
	b.mu.RUnlock()
	return Stats{
		Publishes:   b.publishCount.Load(),
		Drops:       b.dropCount.Load(),
		Subscribers: n,
	}
}

// C returns the subscription's read-only channel. It closes when the
// Sub is closed (either via Sub.Close or Broadcaster.Close).
func (s *Sub[T]) C() <-chan T { return s.ch }

// Drops returns the number of messages this subscription missed
// because its buffer was full at Publish time.
func (s *Sub[T]) Drops() int64 { return s.drops.Load() }

// Close unsubscribes. The channel returned by C is closed. Idempotent.
func (s *Sub[T]) Close() {
	s.parent.mu.Lock()
	defer s.parent.mu.Unlock()
	if _, ok := s.parent.subs[s]; !ok {
		return
	}
	delete(s.parent.subs, s)
	close(s.ch)
}
