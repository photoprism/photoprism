package entity

import (
	"sort"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

// FaceMatchOverlapMax bounds how much larger than the marker a detection claiming it may be.
//
// OverlapPercent divides by the marker's own surface, so a box that merely contains the marker
// scores a perfect 100 while the correctly fitting detection scores less. At low detection floors,
// a low-confidence head-and-shoulders box is exactly that shape.
const FaceMatchOverlapMax = 4

// faceMatchPair is a candidate assignment of a detected face to a marker.
type faceMatchPair struct {
	markerUID string
	detected  int
	overlap   float64
	score     int
}

// MatchFaces assigns each detected face to at most one marker, and returns the index of the
// detection each marker was given. Candidates rank by their overlap with the marker's own area.
func (m Markers) MatchFaces(detected face.Faces) map[string]int {
	return m.matchFaces(detected, nil, false)
}

// MatchFacesBestFit assigns each detected face not yet taken to at most one marker, ranking the
// candidates by intersection over union, so a loose box that contains a marker cannot outrank the
// one that fits it.
func (m Markers) MatchFacesBestFit(detected face.Faces, taken map[int]bool) map[string]int {
	return m.matchFaces(detected, taken, true)
}

// matchFaces pairs markers and detections above the overlap threshold, and assigns the best
// ranked pairs first.
func (m Markers) matchFaces(detected face.Faces, taken map[int]bool, bestFit bool) map[string]int {
	pairs := make([]faceMatchPair, 0)

	for _, marker := range m {
		area := marker.CropArea()

		for i := range detected {
			if taken[i] {
				continue
			}

			candidate := detected[i].CropArea()
			overlap := candidate.OverlapPercent(area)

			if overlap <= face.OverlapThresholdFloor || OversizedFaceMatch(candidate, area) {
				continue
			}

			pair := faceMatchPair{markerUID: marker.MarkerUID, detected: i, overlap: float64(overlap), score: detected[i].Score}

			if bestFit {
				pair.overlap = faceMatchIoU(candidate, area)
			}

			pairs = append(pairs, pair)
		}
	}

	// Confidence breaks a tie before detector output order does. Containment scores 100, so
	// several candidates reach the top and the order they were decoded in would decide.
	sort.Slice(pairs, func(i, j int) bool {
		switch {
		case pairs[i].overlap != pairs[j].overlap:
			return pairs[i].overlap > pairs[j].overlap
		case pairs[i].score != pairs[j].score:
			return pairs[i].score > pairs[j].score
		case pairs[i].markerUID != pairs[j].markerUID:
			return pairs[i].markerUID < pairs[j].markerUID
		default:
			return pairs[i].detected < pairs[j].detected
		}
	})

	result := make(map[string]int)
	used := make(map[int]bool)

	for _, pair := range pairs {
		if _, ok := result[pair.markerUID]; ok || used[pair.detected] {
			continue
		}

		result[pair.markerUID] = pair.detected
		used[pair.detected] = true
	}

	return result
}

// faceMatchIoU returns the intersection over union of two areas, or 0 if they have no surface.
func faceMatchIoU(a, b crop.Area) float64 {
	intersection := a.OverlapArea(b)
	union := float64(a.W)*float64(a.H) + float64(b.W)*float64(b.H) - intersection

	if union <= 0 {
		return 0
	}

	return intersection / union
}

// OversizedFaceMatch reports whether a detection is too much larger than the marker it overlaps to
// be the same face. Without it a box containing the marker outranks every other candidate, because
// the overlap is measured against the marker's surface alone.
func OversizedFaceMatch(candidate, area crop.Area) bool {
	markerSurface := float64(area.W) * float64(area.H)
	candidateSurface := float64(candidate.W) * float64(candidate.H)

	return markerSurface > 0 && candidateSurface > markerSurface*FaceMatchOverlapMax
}
