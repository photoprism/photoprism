package photoprism

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
}

// NewSync returns a new service sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{conf: conf}
}

// Start starts the sync worker.
func (c *Sync) Start() (err error) {
	if err := mutex.Sync.Start(); err != nil {
		event.Error(fmt.Sprintf("import: %s", err.Error()))
		return err
	}

	defer mutex.Sync.Stop()

	log.Info("sync: start")

	return err
}
