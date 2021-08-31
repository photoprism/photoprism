package photoprism

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
)

// FacesOptimizeResult represents the outcome of Faces.Optimize().
type FacesOptimizeResult struct {
	Merged int
}

// Optimize optimizes the face lookup table.
func (w *Faces) Optimize() (result FacesOptimizeResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("facial recognition is disabled")
	}

	faces, err := query.ManuallyAddedFaces()

	if err != nil {
		return result, err
	}

	// Max face index.
	n := len(faces) - 1

	// Need at least 2 faces to optimize.
	if n < 1 {
		return result, nil
	}

	var merge entity.Faces

	for i := 0; i <= n; i++ {
		if len(merge) == 0 {
			merge = entity.Faces{faces[i]}
		} else if faces[i].SubjectUID != merge[len(merge)-1].SubjectUID || i == n {
			if len(merge) < 2 {
				// Nothing to merge.
			} else if _, err := query.MergeFaces(merge); err != nil {
				log.Errorf("%s (merge)", err)
			} else {
				result.Merged += len(merge)
			}

			merge = nil
		} else if ok, dist := merge[0].Match(entity.Embeddings{faces[i].Embedding()}); ok {
			log.Debugf("faces: can merge %s with %s, subject %s, dist %f", merge[0].ID, faces[i].ID, merge[0].SubjectUID, dist)
			merge = append(merge, faces[i])
		} else if len(merge) == 1 {
			merge = nil
		}
	}

	return result, nil
}
