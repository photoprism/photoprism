package photoprism

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
)

// Faces represents a worker for face clustering and matching.
type Faces struct {
	conf      *config.Config
	vetoMu    sync.Mutex
	vetoCache map[string]time.Time
	// reported carries the last value logged for a recurring condition, so a worker that wakes
	// every few minutes states it once rather than every time. Guarded by vetoMu.
	reported map[string]int
}

const faceVetoTTL = 30 * time.Minute

// NewFaces returns a new Faces worker.
func NewFaces(conf *config.Config) *Faces {
	instance := &Faces{
		conf:      conf,
		vetoCache: make(map[string]time.Time),
	}

	return instance
}

// rememberVeto marks a marker as vetoed until the TTL expires.
func (w *Faces) rememberVeto(markerUID string) {
	if markerUID == "" {
		return
	}

	w.vetoMu.Lock()
	defer w.vetoMu.Unlock()

	w.pruneVetoLocked(time.Now())
	w.vetoCache[markerUID] = time.Now().Add(faceVetoTTL)
}

// clearVeto removes a marker UID from the veto cache.
func (w *Faces) clearVeto(markerUID string) {
	if markerUID == "" {
		return
	}

	w.vetoMu.Lock()
	delete(w.vetoCache, markerUID)
	w.vetoMu.Unlock()
}

// vetoed checks whether a marker UID is currently vetoed.
func (w *Faces) vetoed(markerUID string) bool {
	if markerUID == "" {
		return false
	}

	w.vetoMu.Lock()
	defer w.vetoMu.Unlock()

	w.pruneVetoLocked(time.Now())

	_, ok := w.vetoCache[markerUID]
	return ok
}

// pruneVetoLocked removes expired veto entries; caller must hold vetoMu.
func (w *Faces) pruneVetoLocked(now time.Time) {
	for uid, expires := range w.vetoCache {
		if expires.Before(now) {
			delete(w.vetoCache, uid)
		}
	}
}

// StartDefault starts face clustering and matching with default options.
func (w *Faces) StartDefault() (err error) {
	return w.Start(FacesOptions{
		Force: false,
	})
}

// Start face clustering and matching.
//
// The lock is checked here rather than in start, which a migration calls directly while holding
// it: clustering is the last thing a migration does, so refusing it there would leave every
// replacement cluster unbuilt.
func (w *Faces) Start(opt FacesOptions) (err error) {
	// A migration replaces every cluster in one transaction and ordinarily runs in another
	// process, where the worker activity below cannot see it.
	if held := w.conf.FacesLocked(); held != "" {
		log.Infof("faces: waiting for the %s to complete", held)
		return nil
	}

	// A migration that completed in that other process left its target in "options.yml", which
	// this one has not loaded: clustering would compare vectors of two different lengths.
	w.conf.CheckFaceModelSuperseded()

	if err = mutex.FacesWorker.Start(); err != nil {
		return err
	}

	defer mutex.FacesWorker.Stop()

	return w.start(opt)
}

// start performs face clustering and matching while the caller holds the faces worker lock.
func (w *Faces) start(opt FacesOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)\nstack: %s", r, debug.Stack())
			log.Errorf("faces: %s", err)
		}
	}()

	if w.Disabled() {
		return fmt.Errorf("face recognition is disabled")
	}

	// Clustering and matching compare stored vectors, so both are paused while the library
	// holds vectors the configured model cannot read. The reason is reported once when the
	// configuration is initialized; a worker that wakes every few minutes must not repeat it.
	if reason := face.EmbeddingsBlockedReason(); reason != "" {
		log.Debugf("faces: %s, so clustering and matching are paused", reason)
		return nil
	}

	var start time.Time

	// changed records whether this run moved anything the subject counts are computed from, so an
	// idle wake does not pay for the join that refreshes them; every step below that can reassign a
	// marker already reports how many it touched. A forced run recomputes regardless, because drift
	// arising outside the worker - an interrupted run, a photo turned private - moves no marker.
	changed := opt.Force

	// Remove orphan file markers.
	start = time.Now()
	if removed, err := query.RemoveOrphanMarkers(); err != nil {
		log.Errorf("faces: %s (remove orphan markers)", err)
	} else if removed > 0 {
		changed = true
		log.Infof("faces: removed %d orphan markers [%s]", removed, time.Since(start))
	} else {
		log.Debugf("faces: found no orphan markers [%s]", time.Since(start))
	}

	// Repair invalid marker face and subject references.
	start = time.Now()
	if removed, err := query.FixMarkerReferences(); err != nil {
		log.Errorf("markers: %s (fix references)", err)
	} else if removed > 0 {
		changed = true
		log.Infof("markers: fixed %d references [%s]", removed, time.Since(start))
	} else {
		log.Debugf("markers: found no invalid references [%s]", time.Since(start))
	}

	// Create known marker subjects if needed.
	start = time.Now()
	if affected, err := query.CreateMarkerSubjects(); err != nil {
		log.Errorf("markers: %s (create subjects)", err)
	} else if affected > 0 {
		changed = true
		log.Infof("markers: added %d known subjects [%s]", affected, time.Since(start))
	} else {
		log.Debugf("markers: found no missing subjects [%s]", time.Since(start))
	}

	// Resolve collisions of different subject's faces.
	start = time.Now()
	if c, r, err := query.ResolveFaceCollisions(); err != nil {
		log.Errorf("faces: %s (resolve ambiguous subjects)", err)
	} else if c > 0 {
		changed = true
		log.Infof("faces: resolved %d / %d ambiguous subjects [%s]", r, c, time.Since(start))
	} else {
		log.Debugf("faces: found no ambiguous subjects [%s]", time.Since(start))
	}

	// Optimize existing face clusters.
	start = time.Now()
	if res, err := w.Optimize(); err != nil {
		return err
	} else if res.Merged > 0 {
		changed = true
		log.Infof("faces: merged %d clusters [%s]", res.Merged, time.Since(start))
	} else {
		log.Debugf("faces: found no clusters to be merged [%s]", time.Since(start))
	}

	var added entity.Faces

	// Cluster existing face embeddings.
	start = time.Now()
	if added, err = w.Cluster(opt); err != nil {
		log.Errorf("faces: %s (cluster)", err)
	} else if n := len(added); n > 0 {
		log.Infof("faces: added %d new faces [%s]", n, time.Since(start))
	} else {
		log.Debugf("faces: found no new faces [%s]", time.Since(start))
	}

	// Match markers with faces and subjects.
	start = time.Now()
	matches, err := w.Match(opt)

	if err != nil {
		log.Errorf("faces: %s (match)", err)
	}

	// Log face matching results.
	if matches.Updated > 0 {
		changed = true

		log.Infof("faces: updated %s, recognized %s, %d unknown [%s]", english.Plural(int(matches.Updated), "marker", "markers"), english.Plural(int(matches.Recognized), "face", "faces"), matches.Unknown, time.Since(start))
	} else {
		log.Debugf("faces: updated %s, recognized %s, %d unknown [%s]", english.Plural(int(matches.Updated), "marker", "markers"), english.Plural(int(matches.Recognized), "face", "faces"), matches.Unknown, time.Since(start))
	}

	// Remove unused people.
	start = time.Now()
	if count, err := entity.DeleteOrphanPeople(); err != nil {
		log.Errorf("faces: %s (remove people)", err)
	} else if count > 0 {
		log.Debugf("faces: removed %d people [%s]", count, time.Since(start))
	}

	// Remove unused face clusters.
	start = time.Now()
	if count, err := entity.DeleteOrphanFaces(); err != nil {
		log.Errorf("faces: %s (remove clusters)", err)
	} else if count > 0 {
		log.Debugf("faces: removed %d clusters [%s]", count, time.Since(start))
	}

	// Refresh the subject counts this run invalidated.
	//
	// They are read by the people views, which order and filter on file_count, so a person whose
	// markers this run assigned correctly would otherwise sort last or be filtered out - which is
	// how a mistyped person stayed invisible long enough to look as though naming had failed.
	if changed {
		start = time.Now()

		if err = entity.UpdateSubjectCounts(true); err != nil {
			log.Errorf("faces: %s (update subject counts)", err)
		} else {
			log.Debugf("faces: updated subject counts [%s]", time.Since(start))
		}
	}

	entity.UpdateFaces.Store(false)

	return nil
}

// Cancel stops the current operation.
func (w *Faces) Cancel() {
	mutex.FacesWorker.Cancel()
}

// Canceled tests if face clustering and matching should be stopped.
func (w *Faces) Canceled() bool {
	return mutex.FacesWorker.Canceled() || mutex.IndexWorker.Canceled() || mutex.MetaWorker.Canceled()
}

// Disabled tests if face recognition is disabled.
func (w *Faces) Disabled() bool {
	return w.conf.DisableFaces()
}
