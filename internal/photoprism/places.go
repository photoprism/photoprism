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
			err = fmt.Errorf("index: %s (update places)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if err := mutex.MainWorker.Start(); err != nil {
		// A worker is already running.
		log.Warnf("index: %s (update places)", err.Error())
		return []string{}, err
	} else if !w.conf.Sponsor() && !w.conf.Test() {
		log.Errorf(config.MsgSponsorCommand)
		log.Errorf(config.MsgFundingInfo)
		return []string{}, err
	}

	defer mutex.MainWorker.Stop()

	// Fetch cell IDs from index.
	cells, err := query.CellIDs()

	// Error?
	if err != nil {
		return []string{}, err
	} else if len(cells) == 0 {
		log.Warnf("index: found no places to update")
		return []string{}, nil
	}

	// Drop and recreate places database table.
	if err = entity.RecreateTable(entity.Place{}); err != nil {
		return []string{}, fmt.Errorf("index: %s", err)
	}

	// List of updated cells.
	updated = make([]string, 0, len(cells))

	// Update known locations.
	for i, cell := range cells {
		if i%10 == 0 {
			log.Infof("index: updated %s, %d remaining", english.Plural(i, "place", "places"), len(cells)-i)
		}

		if w.Canceled() {
			return updated, nil
		} else if cell.ID == "" || cell.ID == entity.UnknownID {
			// Skip unknown places.
			continue
		}

		// Create cell from location and place ID.
		c := entity.Cell{ID: cell.ID, PlaceID: cell.PlaceID}

		// Fetch updated cell data from backend API.
		if err = c.Refresh(entity.GeoApi); err != nil {
			log.Warnf("index: %s", err)
		} else {
			// Append if successful.
			updated = append(updated, cell.ID)
		}

		// Short break.
		time.Sleep(40 * time.Millisecond)
	}

	// Find and fix bad place ids.
	log.Debug("places: updating references")
	if fixed, err := query.UpdatePlaceIDs(); err != nil {
		log.Errorf("places: %s (updated references)", err)
	} else if fixed > 0 {
		log.Infof("places: updated %d references", fixed)
	}

	// Update photo counts in places.
	if err := entity.UpdatePlacesCounts(); err != nil {
		log.Errorf("index: %s (update counts)", err)
	} else {
		log.Infof("index: updated counts")
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
