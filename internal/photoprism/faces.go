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

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	// Remove invalid reference IDs from markers table.
	if removed, err := query.RemoveInvalidMarkerReferences(); err != nil {
		log.Errorf("faces: %s (remove invalid references)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d invalid references", removed)
	} else {
		log.Debugf("faces: no invalid references")
	}

	// Optimize existing face clusters.
	if res, err := w.Optimize(); err != nil {
		return err
	} else if res.Merged > 0 {
		log.Infof("faces: merged %d clusters", res.Merged)
	} else {
		log.Debugf("faces: no clusters could be merged")
	}

	// Add known marker subjects.
	if affected, err := query.AddMarkerSubjects(); err != nil {
		log.Errorf("faces: %s (match markers with subjects)", err)
	} else if affected > 0 {
		log.Infof("faces: added %d known marker subjects", affected)
	} else {
		log.Debugf("faces: no subjects were missing")
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
	mutex.MainWorker.Cancel()
}

// Disabled tests if facial recognition is disabled.
func (w *Faces) Disabled() bool {
	return !(w.conf.Experimental() && w.conf.Settings().Features.People)
}
