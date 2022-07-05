package config

import (
	"sync"
	"sync/atomic"
)

var (
	extMutex   sync.Mutex
	extensions atomic.Value
)

// Extension represents a named package extension with callbacks.
type Extension struct {
	name         string
	init         func(c *Config) error
	clientValues func(c *Config, t ClientType) Values
}

// Register registers a new package extension.
func Register(name string, initConfig func(c *Config) error, clientConfig func(c *Config, t ClientType) Values) {
	extMutex.Lock()
	n, _ := extensions.Load().([]Extension)
	extensions.Store(append(n, Extension{name, initConfig, clientConfig}))
	extMutex.Unlock()
}

// Extensions returns all registered package extensions.
func Extensions() (ext []Extension) {
	extMutex.Lock()
	ext, _ = extensions.Load().([]Extension)
	extMutex.Unlock()
	return ext
}
