package photoprism

import (
	"fmt"
	"runtime/debug"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
)

// Faces represents a worker for face clustering and matching.
type Faces struct {
	conf *config.Config
}

// NewFaces returns a new Faces worker.
func NewFaces(conf *config.Config) *Faces {
	instance := &Faces{
		conf: conf,
	}

	return instance
}

// Start face clustering and matching.
func (w *Faces) Start(opt FacesOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)\nstack: %s", r, debug.Stack())
			log.Errorf("faces: %s", err)
		}
	}()

	if w.Disabled() {
		return fmt.Errorf("facial recognition is disabled")
	}

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	// Skip clustering if index contains no new face markers and force option isn't set.
	if n := query.CountNewFaceMarkers(); n < 1 && !opt.Force {
		log.Debugf("faces: no new samples")

		var updated int64

		// Adds and reference known marker subjects.
		if affected, err := query.AddMarkerSubjects(); err != nil {
			log.Errorf("faces: %s (match markers with subjects)", err)
		} else {
			updated += affected
		}

		// Match markers with known faces.
		if affected, err := query.MatchFaceMarkers(); err != nil {
			return err
		} else {
			updated += affected
		}

		// Log result.
		if updated > 0 {
			log.Infof("faces: %d markers updated", updated)
		} else {
			log.Debug("faces: no changes")
		}

		// Remove invalid ids from marker table.
		if err := query.CleanInvalidMarkerReferences(); err != nil {
			log.Errorf("faces: %s (clean)", err)
		}

		return nil
	} else {
		log.Infof("faces: %d new samples", n)
	}

	var clustersAdded, clustersRemoved int64

	// Cluster existing face embeddings.
	if clustersAdded, clustersRemoved, err = w.Cluster(opt); err != nil {
		log.Errorf("faces: %s (cluster)", err)
	}

	// Log face clustering results.
	if (clustersAdded - clustersRemoved) != 0 {
		log.Infof("faces: %d clusters added, %d removed", clustersAdded, clustersRemoved)
	} else {
		log.Debugf("faces: %d clusters added, %d removed", clustersAdded, clustersRemoved)
	}

	// Remove invalid marker references.
	if err = query.CleanInvalidMarkerReferences(); err != nil {
		log.Errorf("faces: %s (clean)", err)
	}

	// Match markers with faces and subjects.
	matches, err := w.Match()

	if err != nil {
		log.Errorf("faces: %s (match)", err)
	}

	// Log face matching results.
	if matches.Updated > 0 {
		log.Infof("faces: %d markers updated, %d faces recognized, %d unknown", matches.Updated, matches.Recognized, matches.Unknown)
	} else {
		log.Debugf("faces: %d markers updated, %d faces recognized, %d unknown", matches.Updated, matches.Recognized, matches.Unknown)
	}

	return nil
}

// Cancel stops the current operation.
func (w *Faces) Cancel() {
	mutex.MainWorker.Cancel()
}

// Disabled tests if facial recognition is disabled.
func (w *Faces) Disabled() bool {
	return !(w.conf.Experimental() && w.conf.Settings().Features.People)
}
