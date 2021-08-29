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
	k := len(faces) - 1

	// Need at least 2 faces to optimize.
	if k < 1 {
		return result, nil
	}

	var merge entity.Faces

	for i, f := range faces {
		if i == 0 {
			continue
		}

		// Previous face.
		prev := faces[i-1]

		// Collect faces to merge.
		if prev.SubjectUID == f.SubjectUID {
			if f.SampleRadius < 0.25 {
				f.SampleRadius = 0.25
			}

			if ok, dist := f.Match(entity.Embeddings{prev.Embedding()}); ok {
				log.Debugf("faces: found clusters to merge for subject %s, dist %f", f.SubjectUID, dist)

				if len(merge) == 0 {
					merge = entity.Faces{prev, f}
				} else {
					merge = append(merge, f)
				}
			}
		}

		// Merge matched faces.
		if prev.SubjectUID != f.SubjectUID || i == k {
			if len(merge) < 2 {
				// Nothing to merge.
			} else if _, err := query.MergeFaces(merge); err != nil {
				log.Errorf("faces: %s (merge)", err)
			} else {
				result.Merged += len(merge)
			}

			merge = nil
		}
	}

	return result, nil
}
