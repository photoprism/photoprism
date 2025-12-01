package ordered

import "sync"

// SyncMap wraps Map with an RWMutex so callers can share an ordered map across
// goroutines without managing locks at every call site.
type SyncMap[K comparable, V any] struct {
	Map[K, V]
	sync.RWMutex
}

// NewSyncMap returns an empty thread-safe ordered map.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{*NewMap[K, V](), sync.RWMutex{}}
}

// NewSyncMapWithCapacity returns a thread-safe map with space pre-allocated for
// capacity elements.
func NewSyncMapWithCapacity[K comparable, V any](capacity int) *SyncMap[K, V] {
	return &SyncMap[K, V]{*NewMapWithCapacity[K, V](capacity), sync.RWMutex{}}
}

// Get returns the value for key while holding a read lock.
func (m *SyncMap[K, V]) Get(key K) (value V, ok bool) {
	m.RLock()
	defer m.RUnlock()

	return m.Map.Get(key)
}

// Set stores the value for key while holding an exclusive lock.
func (m *SyncMap[K, V]) Set(key K, value V) bool {
	m.Lock()
	defer m.Unlock()

	return m.Map.Set(key, value)
}

// ReplaceKey safely forwards to Map.ReplaceKey using a write lock.
func (m *SyncMap[K, V]) ReplaceKey(originalKey, newKey K) bool {
	m.Lock()
	defer m.Unlock()

	return m.Map.ReplaceKey(originalKey, newKey)
}

// GetOrDefault is the concurrent-safe variant of Map.GetOrDefault.
func (m *SyncMap[K, V]) GetOrDefault(key K, defaultValue V) V {
	m.RLock()
	defer m.RUnlock()

	return m.Map.GetOrDefault(key, defaultValue)
}

// Len returns the number of elements while holding a read lock.
func (m *SyncMap[K, V]) Len() int {
	m.RLock()
	defer m.RUnlock()

	return m.Map.Len()
}

// Delete removes a key/value pair while holding an exclusive lock.
func (m *SyncMap[K, V]) Delete(key K) (didDelete bool) {
	m.Lock()
	defer m.Unlock()

	return m.Map.Delete(key)
}

// Copy takes a consistent snapshot by holding a read lock during the copy.
func (m *SyncMap[K, V]) Copy() *Map[K, V] {
	m.RLock()
	defer m.RUnlock()

	return m.Map.Copy()
}

// Has performs a constant-time key existence check behind a read lock.
func (m *SyncMap[K, V]) Has(key K) bool {
	m.RLock()
	defer m.RUnlock()

	return m.Map.Has(key)
}
