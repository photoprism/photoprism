package photoprism

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clusters"
)

// Match matches markers with faces and subjects.
func (w *Faces) Match() (recognized, unknown int64, err error) {
	if w.Disabled() {
		return 0, 0, nil
	}

	if faces, err := query.Faces(false); err != nil {
		return recognized, unknown, err
	} else {
		limit := 500
		offset := 0

		for {
			markers, err := query.Markers(limit, offset, entity.MarkerFace, true, false)

			if err != nil {
				return recognized, unknown, err
			}

			if len(markers) == 0 {
				break
			}

			for _, marker := range markers {
				if mutex.MainWorker.Canceled() {
					return recognized, unknown, fmt.Errorf("worker canceled")
				}

				// Pointer to the matching face.
				var f *entity.Face

				// Distance to the matching face.
				var d float64

				// Find the closest face match for marker.
				for _, e := range marker.Embeddings() {
					for i, match := range faces {
						if dist := clusters.EuclideanDistance(e, match.Embedding()); f == nil || dist < d {
							f = &faces[i]
							d = dist
						}
					}
				}

				// No match?
				if f == nil {
					continue
				}

				// Too distant?
				if d > (f.Radius + face.ClusterRadius) {
					continue
				}

				if updated, err := marker.SetFace(f); err != nil {
					log.Errorf("faces: %s", err)
				} else if updated {
					recognized++
				}

				if marker.SubjectUID == "" {
					unknown++
				}
			}

			offset += limit

			time.Sleep(50 * time.Millisecond)
		}
	}

	// Update remaining markers based on current matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return recognized, unknown, err
	} else {
		recognized += m
	}

	// Reset invalid marker data.
	if err := query.TidyMarkers(); err != nil {
		return recognized, unknown, err
	}

	return recognized, unknown, nil
}
