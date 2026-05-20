// Package keyedmutex provides a per-key Mutex: many goroutines can hold
// locks on different keys concurrently, while any two goroutines that
// pick the same key serialize.
//
// Typical use cases:
//
//   - Serializing writes per user / per resource ID without spinning
//     up one sync.Mutex per ID up front
//   - Avoiding double-fetch when a cache key is being recomputed
//     (combine with an LRU/TTL cache)
//   - Cheaply coordinating goroutines whose contention domain is
//     known only at runtime
//
// The implementation keeps a slot in an internal map only while the
// key is locked or has waiters; once the last waiter leaves, the slot
// is removed so an idle map of many one-shot keys does not grow
// unbounded.
//
// Locks are channel-backed, so LockContext can wait with a context and
// return ctx.Err() instead of blocking forever.
package keyedmutex

import (
	"context"
	"fmt"
	"sync"
)

// Mutex is a per-key mutex keyed by K. The zero value is not usable;
// call New.
type Mutex[K comparable] struct {
	mu    sync.Mutex
	slots map[K]*slot
}

type slot struct {
	// tokens has capacity 1. The slot is unlocked when a token is
	// buffered; Lock drains the token, Unlock puts it back.
	tokens chan struct{}
	// refs is the number of goroutines that have called acquire and
	// not yet release. Guarded by Mutex.mu.
	refs int
}

// New returns a Mutex keyed by K.
func New[K comparable]() *Mutex[K] {
	return &Mutex[K]{slots: make(map[K]*slot)}
}

// Lock acquires the lock for key, blocking until it is available.
func (m *Mutex[K]) Lock(key K) {
	s := m.acquire(key)
	<-s.tokens
}

// TryLock attempts to acquire the lock for key without blocking. It
// returns true on success. The corresponding Unlock must be called
// only when TryLock returned true.
func (m *Mutex[K]) TryLock(key K) bool {
	s := m.acquire(key)
	select {
	case <-s.tokens:
		return true
	default:
		m.release(key, s)
		return false
	}
}

// LockContext acquires the lock for key or returns ctx.Err() if ctx
// is canceled first.
func (m *Mutex[K]) LockContext(ctx context.Context, key K) error {
	s := m.acquire(key)
	select {
	case <-s.tokens:
		return nil
	case <-ctx.Done():
		m.release(key, s)
		return ctx.Err()
	}
}

// Unlock releases the lock for key. It panics if key is not currently
// locked.
func (m *Mutex[K]) Unlock(key K) {
	m.mu.Lock()
	s, ok := m.slots[key]
	if !ok {
		m.mu.Unlock()
		panic(fmt.Sprintf("keyedmutex: Unlock of unlocked key %v", key))
	}
	m.mu.Unlock()

	select {
	case s.tokens <- struct{}{}:
	default:
		panic(fmt.Sprintf("keyedmutex: Unlock of unlocked key %v", key))
	}
	m.release(key, s)
}

// WithLock holds the lock for key while fn runs. If ctx is canceled
// before the lock is acquired, fn is not called and ctx.Err() is
// returned. The lock is released even if fn panics.
func (m *Mutex[K]) WithLock(ctx context.Context, key K, fn func() error) error {
	if err := m.LockContext(ctx, key); err != nil {
		return err
	}
	defer m.Unlock(key)
	return fn()
}

// Len returns the number of keys currently locked or with waiters.
// Useful for diagnostics. Returns zero when the map is fully idle.
func (m *Mutex[K]) Len() int {
	m.mu.Lock()
	n := len(m.slots)
	m.mu.Unlock()
	return n
}

// acquire returns a slot for key, creating one if necessary, and
// increments its reference count. Callers must pair acquire with
// release.
func (m *Mutex[K]) acquire(key K) *slot {
	m.mu.Lock()
	s, ok := m.slots[key]
	if !ok {
		s = newSlot()
		m.slots[key] = s
	}
	s.refs++
	m.mu.Unlock()
	return s
}

// release decrements refs and deletes the slot when no one else holds
// or waits on it.
func (m *Mutex[K]) release(key K, s *slot) {
	m.mu.Lock()
	s.refs--
	if s.refs == 0 {
		delete(m.slots, key)
	}
	m.mu.Unlock()
}

func newSlot() *slot {
	s := &slot{tokens: make(chan struct{}, 1)}
	s.tokens <- struct{}{}
	return s
}
