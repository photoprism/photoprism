package config

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

var (
	extInit    sync.Once
	extMutex   sync.Mutex
	extensions atomic.Value
)

// Register registers a new package extension.
func Register(name string, initConfig func(c *Config) error, clientConfig func(c *Config, t ClientType) Map) {
	extMutex.Lock()
	n, _ := extensions.Load().(Extensions)
	extensions.Store(append(n, Extension{name: name, init: initConfig, clientValues: clientConfig}))
	extMutex.Unlock()
}

// Ext returns all registered package extensions.
func Ext() (ext Extensions) {
	extMutex.Lock()
	ext, _ = extensions.Load().(Extensions)
	extMutex.Unlock()
	return ext
}

// Extensions represents a list of package extensions.
type Extensions []Extension

// Extension represents a named package extension with callbacks.
type Extension struct {
	name         string
	init         func(c *Config) error
	clientValues func(c *Config, t ClientType) Map
}

// Init initializes the registered extensions.
func (ext Extensions) Init(c *Config) {
	extInit.Do(func() {
		for _, e := range ext {
			start := time.Now()

			if err := e.init(c); err != nil {
				log.Warnf("config: %s while initializing the %s extension [%s]", err, clean.Log(e.name), time.Since(start))
			} else {
				log.Tracef("config: %s extension loaded [%s]", clean.Log(e.name), time.Since(start))
			}
		}
	})
}
