package photoprism

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
)

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
}

// Match matches markers with faces and subjects.
func (w *Faces) Match() (result FacesMatchResult, err error) {
	if w.Disabled() {
		return result, nil
	}

	faces, err := query.Faces(false, "")

	if err != nil {
		return result, err
	}

	limit := 500
	offset := 0

	for {
		markers, err := query.Markers(limit, offset, entity.MarkerFace, true, false)

		if err != nil {
			return result, err
		}

		if len(markers) == 0 {
			break
		}

		for _, marker := range markers {
			if mutex.MainWorker.Canceled() {
				return result, fmt.Errorf("worker canceled")
			}

			// Pointer to the matching face.
			var f *entity.Face

			// Distance to the matching face.
			var d float64

			// Find the closest face match for marker.
			for i, m := range faces {
				if ok, dist := m.Match(marker.Embeddings()); ok && (f == nil || dist < d) {
					f = &faces[i]
					d = dist
				}
			}

			// No matching face?
			if f == nil {
				if updated, err := marker.ClearFace(); err != nil {
					log.Warnf("faces: %s (clear match)", err)
				} else if updated {
					result.Updated++
				}

				continue
			}

			// Assign matching face to marker.
			updated, err := marker.SetFace(f)

			if err != nil {
				log.Warnf("faces: %s (match)", err)
				continue
			}

			if updated {
				result.Updated++
			}

			if marker.SubjectUID != "" {
				result.Recognized++
			} else {
				result.Unknown++
			}
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	// Update remaining markers based on current matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return result, err
	} else {
		result.Recognized += m
	}

	// Reset invalid marker data.
	if err := query.CleanInvalidMarkerReferences(); err != nil {
		return result, err
	}

	return result, nil
}
