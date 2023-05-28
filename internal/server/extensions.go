package server

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
)

var (
	extInit    sync.Once
	extMutex   sync.Mutex
	extensions atomic.Value
)

// Register registers a new package extension.
func Register(name string, init func(router *gin.Engine, conf *config.Config) error) {
	extMutex.Lock()
	h, _ := extensions.Load().(Extensions)
	extensions.Store(append(h, Extension{name: name, init: init}))
	extMutex.Unlock()
}

// Ext returns all registered package extensions.
func Ext() (ext Extensions) {
	extMutex.Lock()
	ext, _ = extensions.Load().(Extensions)
	extMutex.Unlock()
	return ext
}

// Extension represents a named package extension with callbacks.
type Extension struct {
	name string
	init func(router *gin.Engine, conf *config.Config) error
}

// Extensions represents a list of package extensions.
type Extensions []Extension

// Init initializes the registered extensions.
func (ext Extensions) Init(router *gin.Engine, conf *config.Config) {
	extInit.Do(func() {
		for _, e := range ext {
			start := time.Now()

			if err := e.init(router, conf); err != nil {
				log.Warnf("server: %s in %s extension [%s]", err, clean.Log(e.name), time.Since(start))
			} else {
				log.Tracef("server: %s extension loaded [%s]", clean.Log(e.name), time.Since(start))
			}
		}
	})
}
