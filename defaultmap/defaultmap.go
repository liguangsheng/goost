// Package defaultmap provides a concurrency-safe map that lazily constructs
// a value for any key that has not yet been set.
package defaultmap

import "sync"

// Map is a concurrent map that calls a constructor when Get sees a missing key.
//
// The constructor must not call back into the same Map on the same key, or it
// will deadlock; it should also be cheap, as it runs while a write lock is held.
type Map[K comparable, V any] struct {
	data        map[K]V
	lock        sync.RWMutex
	constructor func(K) V
}

// Make creates a Map that uses constructor to produce values for missing keys.
func Make[K comparable, V any](constructor func(K) V) *Map[K, V] {
	return &Map[K, V]{
		data:        make(map[K]V),
		constructor: constructor,
	}
}

// Get returns the value for key, constructing and storing it if absent.
func (m *Map[K, V]) Get(key K) V {
	v, _ := m.GetOrInit(key)
	return v
}

// GetOrInit returns the value for key. If the key was absent, the
// constructor is called, the value is stored, and loaded is false.
// If the key was already present, loaded is true.
func (m *Map[K, V]) GetOrInit(key K) (value V, loaded bool) {
	m.lock.RLock()
	v, ok := m.data[key]
	m.lock.RUnlock()
	if ok {
		return v, true
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	if got, ok := m.data[key]; ok {
		return got, true
	}
	v = m.constructor(key)
	m.data[key] = v
	return v, false
}

// LoadOrStore is the sync.Map-style entry point that stores value only
// if the key was absent. The constructor is NOT invoked.
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if v, ok := m.data[key]; ok {
		return v, true
	}
	m.data[key] = value
	return value, false
}

// Set replaces the value for key.
func (m *Map[K, V]) Set(key K, value V) {
	m.lock.Lock()
	m.data[key] = value
	m.lock.Unlock()
}

// Has reports whether key has been set or initialized.
func (m *Map[K, V]) Has(key K) bool {
	m.lock.RLock()
	_, ok := m.data[key]
	m.lock.RUnlock()
	return ok
}

// Delete removes key and its value.
func (m *Map[K, V]) Delete(key K) {
	m.lock.Lock()
	delete(m.data, key)
	m.lock.Unlock()
}

// Len returns the current number of entries.
func (m *Map[K, V]) Len() int {
	m.lock.RLock()
	n := len(m.data)
	m.lock.RUnlock()
	return n
}

// Range calls fn for every key/value pair. Returning false stops iteration.
// fn must not call back into the Map; doing so will deadlock.
func (m *Map[K, V]) Range(fn func(K, V) bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for k, v := range m.data {
		if !fn(k, v) {
			return
		}
	}
}
