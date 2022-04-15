package entity

import (
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Strings is a simple string map that should not be accessed by multiple goroutines.
type Strings map[string]string

// StringMap is a string (reverse) lookup map that can be accessed by multiple goroutines.
type StringMap struct {
	sync.RWMutex
	m Strings
	r Strings
}

// NewStringMap creates a new string (reverse) lookup map.
func NewStringMap(s Strings) *StringMap {
	if s == nil {
		return &StringMap{m: make(Strings, 64), r: make(Strings, 64)}
	} else {
		m := &StringMap{m: s, r: make(Strings, len(s))}

		for k := range s {
			m.r[strings.ToLower(s[k])] = k
		}

		return m
	}
}

// Get returns a string from the map, empty if not found.
func (s *StringMap) Get(key string) string {
	if key == "" {
		return ""
	}

	s.RLock()
	defer s.RUnlock()

	return s.m[key]
}

// Key returns a string from the map, empty if not found.
func (s *StringMap) Key(val string) string {
	if val == "" {
		return ""
	}

	s.RLock()
	defer s.RUnlock()

	return s.r[strings.ToLower(val)]
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

	return s.m[key] == val && s.r[strings.ToLower(val)] == key
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
	s.r[strings.ToLower(val)] = key
}

// Unset removes a string from the map.
func (s *StringMap) Unset(key string) {
	if key == "" {
		return
	}

	s.Lock()
	defer s.Unlock()

	if v := s.m[key]; v == "" {
		// Should never happen.
	} else if v = strings.ToLower(v); s.r[v] == key {
		delete(s.r, v)
	}

	delete(s.m, key)
}
