package workers

import (
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Groom represents a groom worker.
type Groom struct {
	conf *config.Config
}

// NewGroom returns a new groom worker.
func NewGroom(conf *config.Config) *Groom {
	return &Groom{conf: conf}
}

// logError logs an error message if err is not nil.
func (worker *Groom) logError(err error) {
	if err != nil {
		log.Errorf("groom: %s", err.Error())
	}
}

// logWarn logs a warning message if err is not nil.
func (worker *Groom) logWarn(err error) {
	if err != nil {
		log.Warnf("groom: %s", err.Error())
	}
}

// Start starts the groom worker.
func (worker *Groom) Start() (err error) {
	if err := mutex.GroomWorker.Start(); err != nil {
		worker.logWarn(err)
		return err
	}

	defer mutex.GroomWorker.Stop()

	return err
}
