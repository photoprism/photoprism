package ordered

import "iter"

// Map maintains key/value pairs and preserves the order in which keys were
// inserted. Internally it keeps a regular Go map for O(1) lookups plus the
// intrusive list defined in list.go to make iteration stable.
type Map[K comparable, V any] struct {
	kv map[K]*Element[K, V]
	ll list[K, V]
}

// NewMap returns an empty ordered map ready for use. The zero value works as
// well, but this helper is clearer at call sites.
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		kv: make(map[K]*Element[K, V]),
	}
}

// NewMapWithCapacity creates a map with enough pre-allocated space to
// hold the specified number of elements.
func NewMapWithCapacity[K comparable, V any](capacity int) *Map[K, V] {
	return &Map[K, V]{
		kv: make(map[K]*Element[K, V], capacity),
	}
}

// NewMapWithElements returns a map pre-populated with the passed elements.
// Elements are copied into the new map to avoid aliasing external list nodes.
func NewMapWithElements[K comparable, V any](els ...*Element[K, V]) *Map[K, V] {
	om := NewMapWithCapacity[K, V](len(els))
	for _, el := range els {
		om.Set(el.Key, el.Value)
	}
	return om
}

// Get returns the value for a key. If the key does not exist, the second return
// parameter will be false and the value will be the zero value for V.
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	v, ok := m.kv[key]
	if ok {
		value = v.Value
	}

	return
}

// Set will set (or replace) a value for a key. If the key was new, then true
// will be returned. The returned value will be false if the value was replaced
// (even if the value was the same).
func (m *Map[K, V]) Set(key K, value V) bool {
	_, alreadyExist := m.kv[key]
	if alreadyExist {
		m.kv[key].Value = value
		return false
	}

	element := m.ll.PushBack(key, value)
	m.kv[key] = element
	return true
}

// ReplaceKey replaces an existing key with a new key while preserving the
// element's position in the iteration order. This function returns true if the
// operation succeeds, or false if originalKey is not found OR newKey already
// exists.
func (m *Map[K, V]) ReplaceKey(originalKey, newKey K) bool {
	element, originalExists := m.kv[originalKey]
	_, newKeyExists := m.kv[newKey]
	if originalExists && !newKeyExists {
		delete(m.kv, originalKey)
		m.kv[newKey] = element
		element.Key = newKey
		return true
	}
	return false
}

// GetOrDefault returns the value for a key. If the key does not exist, returns
// the default value instead.
func (m *Map[K, V]) GetOrDefault(key K, defaultValue V) V {
	if value, ok := m.kv[key]; ok {
		return value.Value
	}

	return defaultValue
}

// GetElement returns the element for a key. If the key does not exist, the
// pointer will be nil.
func (m *Map[K, V]) GetElement(key K) *Element[K, V] {
	element, ok := m.kv[key]
	if ok {
		return element
	}

	return nil
}

// Len returns the number of elements in the map.
func (m *Map[K, V]) Len() int {
	return len(m.kv)
}

// AllFromFront returns a lazy iterator that yields all elements in the map
// starting at the front (oldest Set element). It stops early if the consumer
// returns false.
func (m *Map[K, V]) AllFromFront() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Key, el.Value) {
				return
			}
		}
	}
}

// AllFromBack returns a lazy iterator that yields all elements in the map
// starting at the back (most recent Set element).
func (m *Map[K, V]) AllFromBack() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for el := m.Back(); el != nil; el = el.Prev() {
			if !yield(el.Key, el.Value) {
				return
			}
		}
	}
}

// Keys returns an iterator that yields all keys in insertion order. Use
// slices.Collect(m.Keys()) if a materialized slice is required.
func (m *Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(key K) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Key) {
				return
			}
		}
	}
}

// Values returns an iterator that yields all values in insertion order. Use
// slices.Collect(m.Values()) when you need a slice copy.
func (m *Map[K, V]) Values() iter.Seq[V] {
	return func(yield func(value V) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Value) {
				return
			}
		}
	}
}

// Delete removes a key from the map and unlinks the corresponding element from
// the order list. It returns true when the key existed.
func (m *Map[K, V]) Delete(key K) (didDelete bool) {
	element, ok := m.kv[key]
	if ok {
		m.ll.Remove(element)
		delete(m.kv, key)
	}

	return ok
}

// Front will return the element that is the first (oldest Set element). If
// there are no elements this will return nil.
func (m *Map[K, V]) Front() *Element[K, V] {
	return m.ll.Front()
}

// Back will return the element that is the last (most recent Set element). If
// there are no elements this will return nil.
func (m *Map[K, V]) Back() *Element[K, V] {
	return m.ll.Back()
}

// Copy returns a new Map with the same elements. Callers must avoid concurrent
// writes during the copy or the snapshot may be inconsistent.
func (m *Map[K, V]) Copy() *Map[K, V] {
	m2 := NewMapWithCapacity[K, V](m.Len())
	for el := m.Front(); el != nil; el = el.Next() {
		m2.Set(el.Key, el.Value)
	}
	return m2
}

// Has checks if a key exists in the map.
func (m *Map[K, V]) Has(key K) bool {
	_, exists := m.kv[key]
	return exists
}
