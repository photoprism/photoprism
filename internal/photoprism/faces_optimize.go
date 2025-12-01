package photoprism

import (
	"errors"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
)

// FacesOptimizeResult represents the outcome of Faces.Optimize().
type FacesOptimizeResult struct {
	Merged int
}

// Optimize optimizes the face lookup table.
func (w *Faces) Optimize() (result FacesOptimizeResult, err error) {
	return w.OptimizeFor("")
}

// OptimizeFor optimizes the face lookup table for the given subject UID (or all when empty).
func (w *Faces) OptimizeFor(subjUID string) (result FacesOptimizeResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("face recognition is disabled")
	}

	// Iterative merging of manually added face clusters.
	for i := 0; i <= 10; i++ {
		var n int
		var c = result.Merged
		var merge entity.Faces
		var faces entity.Faces

		// Fetch manually added faces from the database.
		if faces, err = query.ManuallyAddedFaces(false, false, subjUID); err != nil {
			return result, err
		} else if n = len(faces) - 1; n < 1 {
			// Need at least 2 faces to optimize.
			break
		}

		log.Debugf("faces: optimize for %s itr %d n %d", subjUID, i, n)

		// Find and merge matching faces.
		for j := 0; j <= n; j++ {
			if len(merge) == 0 {
				merge = entity.Faces{faces[j]}
			} else if faces[j].SubjUID != merge[len(merge)-1].SubjUID || j == n {
				if len(merge) < 2 {
					// Nothing to merge.
				} else if _, mergeErr := query.MergeFaces(merge, false); mergeErr != nil {
					if errors.Is(mergeErr, query.ErrRetainedManualClusters) {
						subject := entity.SubjNames.Log(merge[0].SubjUID)
						clusterIDs := strings.Join(merge.IDs(), ", ")

						event.SystemWarn([]string{
							"faces",
							"optimize",
							"retained manual clusters after merge",
							"subject %s, iteration %d, cluster %d, count %d, ids %s",
						}, subject, i, j, len(merge), clusterIDs)

						log.Debugf("faces: retained manual clusters after merge: kept %d candidate cluster(s) [%s] for subject %s (merge) itr %d cluster %d", len(merge), clusterIDs, subject, i, j)
					} else {
						log.Errorf("%s (merge) itr %d cluster %d count %d", mergeErr, i, j, len(merge))
					}
				} else {
					// not exactly right, potentially overcounting
					// see https://github.com/photoprism/photoprism/issues/3124#issuecomment-2558299360
					result.Merged += len(merge)
				}

				merge = nil
			} else if ok, dist := merge[0].Match(face.Embeddings{faces[j].Embedding()}); ok {
				log.Debugf("faces: can merge %s with %s, subject %s, dist %f", merge[0].ID, faces[j].ID, entity.SubjNames.Log(merge[0].SubjUID), dist)
				merge = append(merge, faces[j])
			} else if len(merge) == 1 {
				merge = nil
			}
		}

		// Done?
		if result.Merged <= c {
			break
		}
	}

	return result, nil
}
