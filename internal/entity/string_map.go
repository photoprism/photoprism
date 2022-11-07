package entity

import (
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
)

// Strings is a simple string map that should not be accessed by multiple goroutines.
type Strings map[string]string

type MultiStrings map[string][]string

// StringMap is a string (reverse) lookup map that can be accessed by multiple goroutines.
type StringMap struct {
	sync.RWMutex
	m Strings
	r MultiStrings
}

// NewStringMap creates a new string (reverse) lookup map.
func NewStringMap(s Strings) *StringMap {
	if s == nil {
		return &StringMap{m: make(Strings, 64), r: make(MultiStrings, 64)}
	} else {
		m := &StringMap{m: s, r: make(MultiStrings, len(s))}

		for k := range s {
			m.r[strings.ToLower(s[k])] = []string{k}
		}

		return m
	}
}

// Get returns the matching value or an empty string if it was not found.
func (s *StringMap) Get(key string) string {
	if key == "" {
		return ""
	}

	s.RLock()
	defer s.RUnlock()

	return s.m[key]
}

// Has checks whether a value has been set for the specified key.
func (s *StringMap) Has(key string) bool {
	if key == "" {
		return false
	}

	s.RLock()
	defer s.RUnlock()

	_, ok := s.m[key]

	return ok
}

// Missing checks if the key is unknown.
func (s *StringMap) Missing(key string) bool {
	return !s.Has(key)
}

// Key returns the last added key that matches the specified value.
func (s *StringMap) Key(val string) string {
	keys := s.Keys(val)

	if l := len(keys); l > 0 {
		return keys[l-1]
	}

	return ""
}

// Keys returns all keys that match the specified value.
func (s *StringMap) Keys(val string) []string {
	if val == "" {
		return []string{}
	}

	s.RLock()
	defer s.RUnlock()

	return s.r[strings.ToLower(val)]
}

// HasValue checks if the specified value exists for any key.
func (s *StringMap) HasValue(val string) bool {
	if val == "" {
		return false
	}

	s.RLock()
	defer s.RUnlock()

	_, ok := s.r[strings.ToLower(val)]

	return ok
}

// Log returns a string sanitized for logging and using the key as fallback value.
func (s *StringMap) Log(key string) (val string) {
	if key == "" {
		return "<unknown>"
	}

	if val = s.Get(key); val != "" {
		return clean.Log(val)
	} else {
		return clean.Log(key)
	}
}

// Unchanged verifies if the key/value pair is unchanged.
func (s *StringMap) Unchanged(key string, val string) bool {
	if key == "" {
		return true
	}

	s.RLock()
	defer s.RUnlock()

	return s.m[key] == val && list.Contains(s.r[strings.ToLower(val)], key)
}

// Set adds a string to the map.
func (s *StringMap) Set(key string, val string) {
	if s.Unchanged(key, val) {
		return
	} else if val == "" {
		s.Unset(key)
		return
	}

	s.Lock()
	defer s.Unlock()

	s.m[key] = val

	// Update reverse lookup map.
	s.r[strings.ToLower(val)] = list.Add(s.r[strings.ToLower(val)], key)
}

// Unset removes a string from the map.
func (s *StringMap) Unset(key string) {
	if key == "" {
		return
	}

	s.Lock()
	defer s.Unlock()

	// Update reverse lookup map.
	if v := strings.ToLower(s.m[key]); v != "" {
		if keys := list.Remove(s.r[v], key); len(keys) == 0 {
			delete(s.r, v)
		} else {
			s.r[v] = keys
		}
	}

	delete(s.m, key)
}
