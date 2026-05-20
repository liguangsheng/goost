package lru

import "sync"

// Builder constructs a Cache. Use New to obtain one.
type Builder[K comparable, V any] struct {
	cap       int
	locker    sync.Locker
	evictHook EvictHook[K, V]
	shards    int
	hash      HashFunc[K]
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

// Shards configures the builder to produce a ShardedCache with at least n
// shards (rounded up to the next power of two). hash maps keys to shards;
// for string keys use StringHash. Without a call to Shards, BuildSharded
// panics; for string keys, BuildSharded uses StringHash by default.
func (b *Builder[K, V]) Shards(n int, hash HashFunc[K]) *Builder[K, V] {
	b.shards = n
	b.hash = hash
	return b
}

// Build returns the configured Cache.
func (b *Builder[K, V]) Build() *Cache[K, V] {
	return newCache(b.cap, b.locker, b.evictHook)
}

// BuildSharded returns a ShardedCache. The total capacity is distributed
// across shards (capacity per shard = ceil(cap / shards)).
func (b *Builder[K, V]) BuildSharded() *ShardedCache[K, V] {
	if b.hash == nil {
		panic("lru: BuildSharded requires Shards(n, hashFn) to be called with a non-nil hash")
	}
	shardCount := nextPow2(b.shards)
	if shardCount < 1 {
		shardCount = 1
	}
	perShardCap := 0
	if b.cap > 0 {
		perShardCap = (b.cap + shardCount - 1) / shardCount
	}
	shards := make([]*Cache[K, V], shardCount)
	for i := range shards {
		shards[i] = newCache(perShardCap, newLocker(b.locker), b.evictHook)
	}
	return &ShardedCache[K, V]{
		shards: shards,
		mask:   uint64(shardCount - 1),
		hash:   b.hash,
	}
}

// newLocker returns a fresh locker matching template's safety semantics.
// We can't share a single sync.Locker across shards or contention defeats
// the point of sharding.
func newLocker(template sync.Locker) sync.Locker {
	if _, ok := template.(noopLocker); ok {
		return noopLocker{}
	}
	return &sync.Mutex{}
}
