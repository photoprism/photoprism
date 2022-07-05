package server

import (
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
)

var (
	extMutex   sync.Mutex
	extensions atomic.Value
)

// Extension represents a named package extension with callbacks.
type Extension struct {
	name string
	init func(router *gin.Engine, conf *config.Config) error
}

// Register registers a new package extension.
func Register(name string, init func(router *gin.Engine, conf *config.Config) error) {
	extMutex.Lock()
	h, _ := extensions.Load().([]Extension)
	extensions.Store(append(h, Extension{name, init}))
	extMutex.Unlock()
}

// Extensions returns all registered package extensions.
func Extensions() (ext []Extension) {
	extMutex.Lock()
	ext, _ = extensions.Load().([]Extension)
	extMutex.Unlock()
	return ext
}
