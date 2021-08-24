package photoprism

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
)

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
}

// Match matches markers with faces and subjects.
func (w *Faces) Match(opt FacesOptions) (result FacesMatchResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("facial recognition is disabled")
	}

	// Skip matching if index contains no new face markers, and force option isn't set.
	if opt.Force {
		log.Infof("faces: forced matching")
	} else if n := query.CountUnmatchedFaceMarkers(); n > 0 {
		log.Infof("faces: %d unmatched markers", n)
	} else {
		result.Recognized, err = query.MatchFaceMarkers()
		return result, err
	}

	faces, err := query.Faces(false, "")

	if err != nil {
		return result, err
	}

	limit := 100
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
			if w.Canceled() {
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

	// Update remaining markers based on previous matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return result, err
	} else {
		result.Recognized += m
	}

	// Update face match timestamps.
	for _, m := range faces {
		if err := m.UpdateMatchTime(); err != nil {
			log.Warnf("faces: %s (update match time)", err)
		}
	}

	return result, nil
}
