package photoprism

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity/query"
)

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
