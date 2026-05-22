// Package lru implements a generic least-recently-used cache with optional
// per-entry expiration and an evict hook.
package lru

import (
	"container/list"
	"sync"
	"time"
)

// EvictHook is invoked when an entry is evicted because the cache is full.
type EvictHook[K comparable, V any] func(K, V)

type entry[K comparable, V any] struct {
	key      K
	value    V
	expireNs int64 // 0 means no expiration
}

// Snapshot is a point-in-time read-only view of cache size and capacity.
// Capacity is 0 when capacity-based eviction is disabled.
type Snapshot struct {
	Size     int
	Capacity int
	Shards   int
}

// Cache is a generic LRU cache. The zero value is not usable; build one with New.
type Cache[K comparable, V any] struct {
	access     map[K]*list.Element
	ll         *list.List
	maxEntries int
	evictHook  EvictHook[K, V]
	lock       sync.Locker
}

func newCache[K comparable, V any](maxEntries int, lock sync.Locker, hook EvictHook[K, V]) *Cache[K, V] {
	return &Cache[K, V]{
		access:     make(map[K]*list.Element, maxEntries),
		ll:         list.New(),
		maxEntries: maxEntries,
		evictHook:  hook,
		lock:       lock,
	}
}

// Set inserts or updates the value for key without expiration.
func (c *Cache[K, V]) Set(key K, value V) {
	c.lock.Lock()
	c.set(key, value, 0)
	c.lock.Unlock()
}

// SetWithExpire inserts or updates the value for key with an absolute expiration.
func (c *Cache[K, V]) SetWithExpire(key K, value V, expiredAt time.Time) {
	c.lock.Lock()
	c.set(key, value, expiredAt.UnixNano())
	c.lock.Unlock()
}

// SetWithDuration inserts or updates the value for key with a relative expiration.
func (c *Cache[K, V]) SetWithDuration(key K, value V, d time.Duration) {
	c.lock.Lock()
	c.set(key, value, time.Now().Add(d).UnixNano())
	c.lock.Unlock()
}

// Get returns the value for key and whether it was found. Expired entries are
// evicted on access.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ele, ok := c.access[key]
	if !ok {
		var zero V
		return zero, false
	}
	ent := ele.Value.(*entry[K, V])
	if ent.expireNs > 0 && ent.expireNs <= time.Now().UnixNano() {
		c.removeElement(ele)
		var zero V
		return zero, false
	}
	c.ll.MoveToFront(ele)
	return ent.value, true
}

// Peek returns the value for key without updating recency.
func (c *Cache[K, V]) Peek(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ele, ok := c.access[key]
	if !ok {
		var zero V
		return zero, false
	}
	ent := ele.Value.(*entry[K, V])
	if ent.expireNs > 0 && ent.expireNs <= time.Now().UnixNano() {
		c.removeElement(ele)
		var zero V
		return zero, false
	}
	return ent.value, true
}

// Remove deletes key from the cache.
func (c *Cache[K, V]) Remove(key K) {
	c.lock.Lock()
	if ele, ok := c.access[key]; ok {
		c.removeElement(ele)
	}
	c.lock.Unlock()
}

// Size returns the current number of entries.
func (c *Cache[K, V]) Size() int {
	c.lock.Lock()
	n := c.ll.Len()
	c.lock.Unlock()
	return n
}

// Snapshot returns a point-in-time read-only view of the cache.
func (c *Cache[K, V]) Snapshot() Snapshot {
	c.lock.Lock()
	s := Snapshot{Size: c.ll.Len(), Capacity: c.maxEntries, Shards: 1}
	c.lock.Unlock()
	return s
}

// Resize changes the maximum number of entries. If shrinking, the
// least-recently-used entries are evicted (firing EvictHook) until the
// cache fits. n=0 disables capacity-based eviction entirely.
func (c *Cache[K, V]) Resize(n int) {
	c.lock.Lock()
	c.maxEntries = n
	if n > 0 {
		for c.ll.Len() > n {
			c.removeOldest()
		}
	}
	c.lock.Unlock()
}

// Clear removes all entries.
func (c *Cache[K, V]) Clear() {
	c.lock.Lock()
	c.ll = list.New()
	c.access = make(map[K]*list.Element, c.maxEntries)
	c.lock.Unlock()
}

// Keys returns a snapshot of all live (non-expired) keys in
// most-recently-used order. The slice is owned by the caller.
func (c *Cache[K, V]) Keys() []K {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().UnixNano()
	out := make([]K, 0, c.ll.Len())
	for ele := c.ll.Front(); ele != nil; ele = ele.Next() {
		ent := ele.Value.(*entry[K, V])
		if ent.expireNs > 0 && ent.expireNs <= now {
			continue
		}
		out = append(out, ent.key)
	}
	return out
}

// Range calls fn for every live entry in most-recently-used order.
// Returning false stops iteration. fn must not call back into the Cache
// or it will deadlock.
func (c *Cache[K, V]) Range(fn func(K, V) bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().UnixNano()
	for ele := c.ll.Front(); ele != nil; ele = ele.Next() {
		ent := ele.Value.(*entry[K, V])
		if ent.expireNs > 0 && ent.expireNs <= now {
			continue
		}
		if !fn(ent.key, ent.value) {
			return
		}
	}
}

func (c *Cache[K, V]) set(key K, value V, expireNs int64) {
	if ele, ok := c.access[key]; ok {
		ent := ele.Value.(*entry[K, V])
		ent.value = value
		ent.expireNs = expireNs
		c.ll.MoveToFront(ele)
		return
	}

	ent := &entry[K, V]{key: key, value: value, expireNs: expireNs}
	c.access[key] = c.ll.PushFront(ent)

	if c.maxEntries > 0 && c.ll.Len() > c.maxEntries {
		c.removeOldest()
	}
}

func (c *Cache[K, V]) removeOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	if c.evictHook != nil {
		ent := ele.Value.(*entry[K, V])
		c.evictHook(ent.key, ent.value)
	}
	c.removeElement(ele)
}

func (c *Cache[K, V]) removeElement(ele *list.Element) {
	ent := ele.Value.(*entry[K, V])
	c.ll.Remove(ele)
	delete(c.access, ent.key)
}
