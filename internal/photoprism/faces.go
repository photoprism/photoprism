package photoprism

import (
	"fmt"
	"runtime/debug"

	"github.com/photoprism/photoprism/internal/entity"

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

// StartDefault starts face clustering and matching with default options.
func (w *Faces) StartDefault() (err error) {
	return w.Start(FacesOptions{
		Force: false,
	})
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

	if err := mutex.FacesWorker.Start(); err != nil {
		return err
	}

	defer mutex.FacesWorker.Stop()

	// Repair invalid marker face and subject references.
	if removed, err := query.FixMarkerReferences(); err != nil {
		log.Errorf("faces: %s (fix references)", err)
	} else if removed > 0 {
		log.Infof("faces: fixed %d marker references", removed)
	} else {
		log.Debugf("faces: no invalid marker references")
	}

	// Create known marker subjects if needed.
	if affected, err := query.CreateMarkerSubjects(); err != nil {
		log.Errorf("faces: %s (create subjects)", err)
	} else if affected > 0 {
		log.Infof("faces: added %d known marker subjects", affected)
	} else {
		log.Debugf("faces: marker subjects already exist")
	}

	// Resolve collisions of different subject's faces.
	if c, r, err := query.ResolveFaceCollisions(); err != nil {
		log.Errorf("faces: %s (resolve collisions)", err)
	} else if c > 0 {
		log.Infof("faces: resolved %d / %d collisions", r, c)
	} else {
		log.Debugf("faces: no collisions detected")
	}

	// Optimize existing face clusters.
	if res, err := w.Optimize(); err != nil {
		return err
	} else if res.Merged > 0 {
		log.Infof("faces: merged %d clusters", res.Merged)
	} else {
		log.Debugf("faces: no clusters could be merged")
	}

	var added entity.Faces

	// Cluster existing face embeddings.
	if added, err = w.Cluster(opt); err != nil {
		log.Errorf("faces: %s (cluster)", err)
	} else if n := len(added); n > 0 {
		log.Infof("faces: added %d new faces", n)
	} else {
		log.Debugf("faces: found no new faces")
	}

	// Match markers with faces and subjects.
	matches, err := w.Match(opt)

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
	mutex.FacesWorker.Cancel()
}

// Canceled tests if face clustering and matching should be stopped.
func (w *Faces) Canceled() bool {
	return mutex.FacesWorker.Canceled() || mutex.MainWorker.Canceled() || mutex.MetaWorker.Canceled()
}

// Disabled tests if facial recognition is disabled.
func (w *Faces) Disabled() bool {
	return !w.conf.Settings().Features.People
}
