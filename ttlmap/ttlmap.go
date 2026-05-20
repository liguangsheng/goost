// Package ttlmap provides a concurrency-safe map with per-entry expiration.
//
// Unlike lru.Cache, TTLMap has no capacity bound; it is suited to caches
// whose size is governed by expiration alone. Expired entries are lazily
// removed on access and periodically swept by a background goroutine.
package ttlmap

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value    V
	expireNs int64
}

// TTLMap is a concurrent map with per-key expiration.
type TTLMap[K comparable, V any] struct {
	mu       sync.RWMutex
	data     map[K]entry[V]
	stop     chan struct{}
	stopOnce sync.Once
}

// New creates a TTLMap and starts a sweep goroutine that runs every
// sweepEvery (use 0 to disable background sweeping). Call Close to stop it.
func New[K comparable, V any](sweepEvery time.Duration) *TTLMap[K, V] {
	m := &TTLMap[K, V]{
		data: make(map[K]entry[V]),
		stop: make(chan struct{}),
	}
	if sweepEvery > 0 {
		go m.sweepLoop(sweepEvery)
	}
	return m
}

// Set inserts or updates key with the given TTL. A non-positive ttl means
// the entry never expires.
func (m *TTLMap[K, V]) Set(key K, value V, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}
	m.mu.Lock()
	m.data[key] = entry[V]{value: value, expireNs: exp}
	m.mu.Unlock()
}

// Get returns the value for key and whether it was found and not expired.
// An expired entry is removed on access.
func (m *TTLMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	e, ok := m.data[key]
	m.mu.RUnlock()
	if !ok {
		var zero V
		return zero, false
	}
	if e.expireNs > 0 && e.expireNs <= time.Now().UnixNano() {
		m.mu.Lock()
		// recheck under the write lock to avoid removing a refresh
		if cur, ok := m.data[key]; ok && cur.expireNs == e.expireNs {
			delete(m.data, key)
		}
		m.mu.Unlock()
		var zero V
		return zero, false
	}
	return e.value, true
}

// Delete removes key.
func (m *TTLMap[K, V]) Delete(key K) {
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()
}

// Len reports the current number of entries, including ones that may be
// expired but not yet swept.
func (m *TTLMap[K, V]) Len() int {
	m.mu.RLock()
	n := len(m.data)
	m.mu.RUnlock()
	return n
}

// Close stops the sweep goroutine. Subsequent calls are no-ops. The map
// remains usable; entries are still expired on access.
func (m *TTLMap[K, V]) Close() {
	m.stopOnce.Do(func() { close(m.stop) })
}

func (m *TTLMap[K, V]) sweepLoop(every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-m.stop:
			return
		case now := <-t.C:
			m.sweep(now.UnixNano())
		}
	}
}

func (m *TTLMap[K, V]) sweep(nowNs int64) {
	m.mu.Lock()
	for k, e := range m.data {
		if e.expireNs > 0 && e.expireNs <= nowNs {
			delete(m.data, k)
		}
	}
	m.mu.Unlock()
}
