package photoprism

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
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

// Start runs the Places worker to update location information.
func (w *Places) Start(force bool) (updated []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("index: %s (update locations)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Check if a worker is already running.
	if err = mutex.IndexWorker.Start(); err != nil {
		log.Warnf("index: %s (update locations)", err.Error())
		return []string{}, err
	}

	defer mutex.IndexWorker.Stop()

	// Update existing location information?
	if force {
		cells, queryErr := query.CellIDs()

		// Check if query failed.
		if queryErr != nil {
			return []string{}, queryErr
		} else if len(cells) == 0 {
			log.Warnf("index: found no locations to update")
			return []string{}, nil
		}

		// List of updated cells.
		updated = make([]string, 0, len(cells))

		log.Infof("index: fetching location details")

		// Update known locations.
		for i, cell := range cells {
			if i%10 == 0 {
				log.Infof("index: fetched %s, %d remaining",
					english.Plural(i, "location", "locations"),
					len(cells)-i)
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
			if queryErr = c.Refresh(entity.GeoApi); queryErr != nil {
				log.Warnf("index: %s", queryErr)
			} else {
				// Append if successful.
				updated = append(updated, cell.ID)
			}

			// Wait 20ms before fetching the next location.
			time.Sleep(20 * time.Millisecond)
		}
	}

	// Remove unused entries from the places table.
	if err = query.PurgePlaces(); err != nil {
		log.Errorf("index: %s (purge places)", err)
	}

	// Update location-related photo metadata in the index.
	if _, err = w.UpdatePhotos(force); err != nil {
		log.Errorf("index: %s (update photos)", err)
	}

	// Update photo counts in places.
	if err = entity.UpdatePlacesCounts(); err != nil {
		log.Errorf("index: %s (update counts)", err)
	}

	return updated, err
}

// UpdatePhotos updates all location-related photo metadata in the index.
func (w *Places) UpdatePhotos(force bool) (affected int, err error) {
	start := time.Now()

	var u []string

	q := query.UnscopedDb()

	// Only select photos with unresolved details or force an update of all photos with location?
	if force {
		q = q.Raw(`SELECT photo_uid FROM photos WHERE place_id <> 'zz' OR photo_lat <> 0 OR photo_lng <> 0 ORDER BY id`)
	} else {
		q = q.Raw(`SELECT photo_uid FROM photos WHERE (place_id = 'zz' OR cell_id = 'zz') AND (photo_lat <> 0 OR photo_lng <> 0) ORDER BY id`)
	}

	// Run select query.
	if err = q.Pluck("photo_uid", &u).Error; err != nil {
		return affected, err
	}

	// Get number of results.
	n := len(u)

	// Return if no results.
	if n == 0 {
		log.Debugf("index: found no photos with location [%s]", time.Since(start))
		return affected, err
	}

	log.Infof("index: updating references, titles, and keywords")

	for i := 0; i < n; i++ {
		if i%10 == 0 {
			log.Infof("index: updated %s, %s remaining",
				english.Plural(i, "photo", "photos"),
				english.Plural(n-i, "photo", "photos"))
		}

		var model entity.Photo

		model, err = query.PhotoByUID(u[i])

		if err != nil {
			log.Errorf("index: %s while loading %s", err, model.String())
			continue
		} else if model.NoLatLng() {
			log.Debugf("index: photo %s has no location", model.String())
			continue
		}

		if err = model.SaveLocation(); err != nil {
			log.Errorf("index: %s while updating %s", err, model.String())
		} else {
			affected++
		}
	}

	return affected, err
}

// Canceled tests if the worker should be stopped.
func (w *Places) Canceled() bool {
	return mutex.IndexWorker.Canceled() || mutex.MetaWorker.Canceled()
}

// Cancel stops the current operation.
func (w *Places) Cancel() {
	mutex.IndexWorker.Cancel()
}
