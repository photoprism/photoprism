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
	// Early extension registry for hooks that must run before DB connect.
	earlyExtInit    sync.Once
	earlyExtMutex   sync.Mutex
	earlyExtensions atomic.Value
)

// TODO: Provide a test-only reset for earlyExtensions and extensions if we ever
// need to reinitialize different early hooks across multiple test packages.
// sync.Once currently prevents re-running initializers within the same process.

// Register registers a new package extension.
func Register(name string, initConfig func(c *Config) error, clientConfig func(c *Config, t ClientType) Map) {
	extMutex.Lock()
	defer extMutex.Unlock()

	n, _ := extensions.Load().(Extensions)
	extensions.Store(append(n, Extension{name: name, init: initConfig, clientValues: clientConfig}))
}

// RegisterEarly registers a package extension that should run before the
// database connection is established. Use this for hooks that may influence
// DB settings or other early configuration.
func RegisterEarly(name string, initConfig func(c *Config) error, clientConfig func(c *Config, t ClientType) Map) {
	earlyExtMutex.Lock()
	defer earlyExtMutex.Unlock()

	n, _ := earlyExtensions.Load().(Extensions)
	earlyExtensions.Store(append(n, Extension{name: name, init: initConfig, clientValues: clientConfig}))
}

// Ext returns all registered package extensions.
func Ext() (ext Extensions) {
	extMutex.Lock()
	defer extMutex.Unlock()

	ext, _ = extensions.Load().(Extensions)

	return ext
}

// EarlyExt returns all registered early package extensions.
func EarlyExt() (ext Extensions) {
	earlyExtMutex.Lock()
	defer earlyExtMutex.Unlock()

	ext, _ = earlyExtensions.Load().(Extensions)

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
				log.Warnf("config: %s when loading %s extension", err, clean.Log(e.name))
			} else {
				log.Tracef("config: %s extension loaded [%s]", clean.Log(e.name), time.Since(start))
			}
		}
	})
}

// InitEarly initializes early registered extensions.
func (ext Extensions) InitEarly(c *Config) {
	earlyExtInit.Do(func() {
		for _, e := range ext {
			start := time.Now()

			if err := e.init(c); err != nil {
				log.Warnf("config: %s when loading early %s extension", err, clean.Log(e.name))
			} else {
				log.Tracef("config: early %s extension loaded [%s]", clean.Log(e.name), time.Since(start))
			}
		}
	})
}
