package lru

import "sync"

// Builder constructs a Cache. Use New to obtain one.
type Builder[K comparable, V any] struct {
	cap       int
	locker    sync.Locker
	evictHook EvictHook[K, V]
}

// New creates a Builder with sensible defaults: capacity 10000 and concurrency-safe.
func New[K comparable, V any]() *Builder[K, V] {
	return &Builder[K, V]{
		cap:    10000,
		locker: &sync.Mutex{},
	}
}

// Cap sets the maximum number of entries. A value of 0 disables eviction.
func (b *Builder[K, V]) Cap(n int) *Builder[K, V] {
	b.cap = n
	return b
}

// Safe toggles internal locking. When false, callers must serialize access.
func (b *Builder[K, V]) Safe(safe bool) *Builder[K, V] {
	if safe {
		b.locker = &sync.Mutex{}
	} else {
		b.locker = noopLocker{}
	}
	return b
}

// Evict installs a hook that is called whenever an entry is evicted due to
// capacity. It is not called by Remove or Clear.
func (b *Builder[K, V]) Evict(fn EvictHook[K, V]) *Builder[K, V] {
	b.evictHook = fn
	return b
}

// Build returns the configured Cache.
func (b *Builder[K, V]) Build() *Cache[K, V] {
	return newCache(b.cap, b.locker, b.evictHook)
}
