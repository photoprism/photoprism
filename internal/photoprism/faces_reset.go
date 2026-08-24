package photoprism

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// runFacesReindex delegates face-only indexing to the supplied Index instance; tests may override it.
var runFacesReindex = func(index *Index, opt IndexOptions) (fs.Done, int, error) {
	if index == nil {
		return nil, 0, fmt.Errorf("faces: index service unavailable")
	}

	found, updated := index.Start(opt)
	return found, updated, nil
}

// Reset removes automatically added face clusters, marker matches, and dangling subjects.
func (w *Faces) Reset() (err error) {
	// Remove automatically added subject and face references from the markers table.
	if removed, err := query.ResetFaceMarkerMatches(); err != nil {
		return fmt.Errorf("faces: %s (reset markers)", err)
	} else {
		log.Infof("faces: removed %d face matches", removed)
	}

	// Remove automatically added face clusters from the index.
	if removed, err := query.RemoveAutoFaceClusters(); err != nil {
		return fmt.Errorf("faces: %s (reset faces)", err)
	} else {
		log.Infof("faces: removed %d face clusters", removed)
	}

	// Remove dangling marker subjects.
	if removed, err := query.RemoveOrphanSubjects(); err != nil {
		return fmt.Errorf("faces: %s (reset subjects)", err)
	} else {
		log.Infof("faces: removed %d dangling subjects", removed)
	}

	return nil
}

// ResetAndReindex resets face data and regenerates markers with the specified detector, or resets
// only when none is named.
//
// The detector is what a caller has to name, because every one of them runs on the same runtime:
// naming the runtime would not say which model places the landmarks, and those decide the crop.
func (w *Faces) ResetAndReindex(detector string, index *Index) error {
	name := strings.TrimSpace(detector)

	if name != "" && !face.KnownDetectorName(name) {
		return fmt.Errorf("faces: unsupported face detector %q", detector)
	}

	regenerate := name != "" && face.ParseDetectorName(name) != face.DetectorNone

	if regenerate && w.conf == nil {
		return fmt.Errorf("faces: configuration not available")
	}

	if regenerate {
		w.conf.Options().FaceDetector = face.ParseDetectorName(name)

		// Checked before anything is removed: a request to regenerate that cannot be met would
		// otherwise delete every person and face and rebuild nothing.
		if w.conf.FaceDetector() == face.DetectorNone {
			return fmt.Errorf("faces: face detector %s cannot be used, so markers cannot be regenerated", clean.Log(name))
		}
	}

	if err := w.Reset(); err != nil {
		return err
	}

	if !regenerate {
		return nil
	}

	if err := w.conf.ConfigureFaceDetector(0); err != nil {
		return err
	}

	convert := w.conf.Settings().Index.Convert && w.conf.SidecarWritable()
	opt := IndexOptionsFacesOnly(w.conf)
	opt.Convert = convert

	found, updated, err := runFacesReindex(index, opt)
	if err != nil {
		return err
	}

	log.Infof("faces: regenerated %s with detector %s (%s scanned)",
		english.Plural(updated, "file", "files"), clean.Log(w.conf.FaceDetector()), english.Plural(len(found), "file", "files"))

	return nil
}
