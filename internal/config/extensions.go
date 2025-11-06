package config

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Stage identifies when an extension should run during the configuration lifecycle.
type Stage string

const (
	// StageBoot runs extensions before a database connection is established so they
	// can adjust low-level settings such as storage paths or database credentials.
	StageBoot Stage = "boot"
	// StageInit runs extensions after the database has been opened and runtime
	// services are available.
	StageInit Stage = "init"
)

var (
	extInit        sync.Once
	extMutex       sync.Mutex
	extensions     atomic.Value
	bootInit       sync.Once
	bootMutex      sync.Mutex
	bootExtensions atomic.Value
)

// Extensions represents a list of package extensions.
type Extensions []Extension

// Extension represents a named package extension with callbacks.
type Extension struct {
	name         string
	init         func(c *Config) error
	clientValues func(c *Config, t ClientType) Values
}

// Register registers a new package extension so it runs during the specified stage.
// StageBoot handlers execute before the database connection is opened, whereas
// StageInit handlers run afterwards. Unknown stages are ignored with a warning.
func Register(stage Stage, name string, initConfig func(c *Config) error, clientConfig func(c *Config, t ClientType) Values) {
	switch stage {
	case StageBoot:
		// BootRegister registers a package extension that should run before the
		// database connection is established. Use this for hooks that may influence
		// DB settings or other early configuration.
		bootMutex.Lock()
		defer bootMutex.Unlock()
		n, _ := bootExtensions.Load().(Extensions)
		bootExtensions.Store(append(n, Extension{name: name, init: initConfig, clientValues: clientConfig}))
	case StageInit:
		extMutex.Lock()
		defer extMutex.Unlock()
		n, _ := extensions.Load().(Extensions)
		extensions.Store(append(n, Extension{name: name, init: initConfig, clientValues: clientConfig}))
	default:
		log.Warnf("config: invalid extension stage %s", clean.Log(string(stage)))
	}
}

// Ext returns all registered extensions for the specified stage.
func Ext(stage Stage) (ext Extensions) {
	switch stage {
	case StageBoot:
		bootMutex.Lock()
		defer bootMutex.Unlock()
		ext, _ = bootExtensions.Load().(Extensions)
		return ext
	case StageInit:
		extMutex.Lock()
		defer extMutex.Unlock()
		ext, _ = extensions.Load().(Extensions)
		return ext
	default:
		log.Warnf("config: invalid extension stage %s", clean.Log(string(stage)))
		return ext
	}
}

// Boot calls the registered extensions before a database connection is
// established. Each extension is executed at most once per process.
func (ext Extensions) Boot(c *Config) {
	bootInit.Do(func() {
		for _, e := range ext {
			start := time.Now()

			if err := e.init(c); err != nil {
				log.Warnf("config: %s when loading %s boot extension", err, clean.Log(e.name))
			} else {
				log.Tracef("config: %s boot extension loaded [%s]", clean.Log(e.name), time.Since(start))
			}
		}
	})
}

// Init calls the registered extensions once a database connection has been
// established. Each extension is executed at most once per process.
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
