package lru

import (
	"hash/fnv"
	"time"
)

// HashFunc maps a key to a uint64 used for shard selection. The same key
// must always hash to the same bucket; the hash does not need to be
// cryptographic.
type HashFunc[K comparable] func(K) uint64

// StringHash is a convenient HashFunc for keys whose string representation
// is itself the key (i.e. `~string` types).
func StringHash[K ~string](k K) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(string(k)))
	return h.Sum64()
}

// ShardedCache partitions a Cache across N shards keyed by hash, reducing
// lock contention under high concurrency. Per-shard semantics match Cache:
// each shard maintains its own LRU order and its own capacity.
type ShardedCache[K comparable, V any] struct {
	shards []*Cache[K, V]
	mask   uint64 // shardCount-1; shardCount is a power of two
	hash   HashFunc[K]
}

func (c *ShardedCache[K, V]) shard(key K) *Cache[K, V] {
	return c.shards[c.hash(key)&c.mask]
}

// Set inserts or updates the value for key without expiration.
func (c *ShardedCache[K, V]) Set(key K, value V) {
	c.shard(key).Set(key, value)
}

// SetWithExpire inserts or updates the value with an absolute expiration.
func (c *ShardedCache[K, V]) SetWithExpire(key K, value V, expiredAt time.Time) {
	c.shard(key).SetWithExpire(key, value, expiredAt)
}

// SetWithDuration inserts or updates the value with a relative expiration.
func (c *ShardedCache[K, V]) SetWithDuration(key K, value V, d time.Duration) {
	c.shard(key).SetWithDuration(key, value, d)
}

// Get returns the value for key. Expired entries are evicted on access.
func (c *ShardedCache[K, V]) Get(key K) (V, bool) { return c.shard(key).Get(key) }

// Peek returns the value for key without updating recency.
func (c *ShardedCache[K, V]) Peek(key K) (V, bool) { return c.shard(key).Peek(key) }

// Remove deletes key from the cache.
func (c *ShardedCache[K, V]) Remove(key K) { c.shard(key).Remove(key) }

// Size returns the total number of entries across all shards.
func (c *ShardedCache[K, V]) Size() int {
	n := 0
	for _, s := range c.shards {
		n += s.Size()
	}
	return n
}

// Clear removes all entries from every shard.
func (c *ShardedCache[K, V]) Clear() {
	for _, s := range c.shards {
		s.Clear()
	}
}

// Keys returns a snapshot of all live keys across all shards. Order is
// shard-by-shard MRU-first; not a global ordering.
func (c *ShardedCache[K, V]) Keys() []K {
	out := make([]K, 0)
	for _, s := range c.shards {
		out = append(out, s.Keys()...)
	}
	return out
}

// Range calls fn for every live entry across all shards. Returning false
// stops iteration. Each shard's lock is held only while iterating that
// shard, so concurrent writes to other shards proceed normally.
func (c *ShardedCache[K, V]) Range(fn func(K, V) bool) {
	for _, s := range c.shards {
		cont := true
		s.Range(func(k K, v V) bool {
			if !fn(k, v) {
				cont = false
				return false
			}
			return true
		})
		if !cont {
			return
		}
	}
}

// nextPow2 returns the smallest power of two >= n.
func nextPow2(n int) int {
	if n <= 1 {
		return 1
	}
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}
