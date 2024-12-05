package defaultmap

import "sync"

// Map is a concurrent-safe map that generates default values
// for non-existing keys.
type Map[K comparable, V any] struct {
	data        map[K]V
	lock        sync.RWMutex
	constructor func(k K) V
}

// Make creates a new DefaultMap instance.
func Make[K comparable, V any](constructor func(K) V) *Map[K, V] {
	return &Map[K, V]{
		data:        make(map[K]V),
		constructor: constructor,
	}
}

// Get returns the value of the specified key in the map. If the key
// does not exist, it generates a default value using the default
// value function and stores it in the map.
func (m *Map[K, V]) Get(key K) V {
	m.lock.RLock()
	val, ok := m.data[key]
	m.lock.RUnlock()

	if ok {
		return val
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if val2, ok2 := m.data[key]; ok2 {
		return val2
	}
	val = m.constructor(key)
	m.data[key] = val
	return val
}

// Set stores the specified value for the specified key in the map.
func (m *Map[K, V]) Set(key K, value V) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[key] = value
}

// Delete removes the specified key and its value from the map.
func (m *Map[K, V]) Delete(key K) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.data, key)
}
