package photoprism

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
)

// Places represents a geo data worker.
type Places struct {
	conf *config.Config
}

// NewPlaces returns a new Places worker.
func NewPlaces(conf *config.Config) *Places {
	instance := &Places{
		conf: conf,
	}

	return instance
}

// Start runs the Places worker.
func (w *Places) Start() (updated []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("places: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.MainWorker.Start(); err != nil {
		// Already running.
		log.Warnf("places: %s (start)", err.Error())
		return []string{}, err
	} else if !w.conf.Sponsor() && !w.conf.Test() {
		// Only for sponsors as this puts load on our API.
		log.Warnf("places: only sponsors may fetch updated location infos")
		return []string{}, err
	}

	defer mutex.MainWorker.Stop()

	// Fetch cell IDs from index.
	cells, err := query.CellIDs()

	if err != nil {
		return []string{}, err
	} else if len(cells) == 0 {
		log.Warnf("places: found no locations")
		return []string{}, nil
	}

	log.Infof("places: updating %s", english.Plural(len(cells), "location", "locations"))

	updated = make([]string, 0, len(cells))

	// Update known locations.
	for _, id := range cells {
		if w.Canceled() {
			return updated, nil
		} else if id == "" || id == entity.UnknownID {
			// Skip unknown places.
			continue
		}

		c := entity.Cell{ID: id}

		// Fetch updated information from backend API.
		if err = c.Refresh(entity.GeoApi); err != nil {
			log.Errorf("places: %s", err)
		} else {
			updated = append(updated, id)
		}

		// Short break.
		time.Sleep(25 * time.Millisecond)
	}

	return updated, err
}

// Canceled tests if the worker should be stopped.
func (w *Places) Canceled() bool {
	return mutex.MainWorker.Canceled() || mutex.MetaWorker.Canceled()
}

// Cancel stops the current operation.
func (w *Places) Cancel() {
	mutex.MainWorker.Cancel()
}
