package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

func TestMarkers_MatchFaces(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		markers := Markers{
			{MarkerUID: "m1", X: 0.1, Y: 0.1, W: 0.2, H: 0.2},
			{MarkerUID: "m2", X: 0.6, Y: 0.6, W: 0.2, H: 0.2},
		}
		detected := face.Faces{
			{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)},
			{Rows: 100, Cols: 100, Area: face.NewArea("face", 70, 70, 20)},
		}

		result := markers.MatchFaces(detected)

		require.Len(t, result, 2)
		assert.Equal(t, 0, result["m1"])
		assert.Equal(t, 1, result["m2"])
	})
	t.Run("NoMarkers", func(t *testing.T) {
		assert.Empty(t, Markers(nil).MatchFaces(face.Faces{{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)}}))
	})
	t.Run("NoDetections", func(t *testing.T) {
		assert.Empty(t, Markers{{MarkerUID: "m1", X: 0.1, Y: 0.1, W: 0.2, H: 0.2}}.MatchFaces(nil))
	})
	t.Run("PrefersConfidence", func(t *testing.T) {
		// A tie on overlap is broken by detection score rather than by the order the detector
		// emitted them, since containment puts several candidates at the top of the list.
		markers := Markers{{MarkerUID: "m1", X: 0.4, Y: 0.4, W: 0.1, H: 0.1}}
		detected := face.Faces{
			{Rows: 100, Cols: 100, Score: 12, Area: face.NewArea("face", 45, 45, 10)},
			{Rows: 100, Cols: 100, Score: 92, Area: face.NewArea("face", 45, 45, 10)},
		}

		result := markers.MatchFaces(detected)

		require.Contains(t, result, "m1")
		assert.Equal(t, 1, result["m1"], "the more confident detection must claim the marker")
	})
	t.Run("OneDetectionPerMarker", func(t *testing.T) {
		contested := Markers{
			{MarkerUID: "m1", X: 0.40, Y: 0.40, W: 0.10, H: 0.10},
			{MarkerUID: "m2", X: 0.42, Y: 0.42, W: 0.10, H: 0.10},
		}
		one := face.Faces{{Rows: 100, Cols: 100, Score: 80, Area: face.NewArea("face", 45, 45, 10)}}

		assert.Equal(t, map[string]int{"m1": 0}, contested.MatchFaces(one))
	})
}

// TestOversizedFaceMatch pins the bound that keeps a containing box from claiming a marker.
func TestOversizedFaceMatch(t *testing.T) {
	marker := crop.Area{Name: "face", X: 0.4, Y: 0.4, W: 0.1, H: 0.1}

	t.Run("SameSize", func(t *testing.T) {
		assert.False(t, OversizedFaceMatch(crop.Area{X: 0.4, Y: 0.4, W: 0.1, H: 0.1}, marker))
	})
	t.Run("SlightlyLarger", func(t *testing.T) {
		// Ordinary detector-to-detector drift must still match.
		assert.False(t, OversizedFaceMatch(crop.Area{X: 0.38, Y: 0.38, W: 0.14, H: 0.14}, marker))
	})
	t.Run("HeadAndShoulders", func(t *testing.T) {
		assert.True(t, OversizedFaceMatch(crop.Area{X: 0.3, Y: 0.3, W: 0.3, H: 0.3}, marker))
	})
	t.Run("EmptyMarker", func(t *testing.T) {
		// Nothing to compare against, so nothing is rejected on size.
		assert.False(t, OversizedFaceMatch(crop.Area{X: 0.3, Y: 0.3, W: 0.3, H: 0.3}, crop.Area{}))
	})
}

func TestMarkers_MatchFacesBestFit(t *testing.T) {
	t.Run("PrefersTheFittingBox", func(t *testing.T) {
		// A loose box containing the marker scores 100 against the marker's own area, while the
		// box that fits it scores less, so only intersection over union picks the fitting one.
		markers := Markers{{MarkerUID: "m1", X: 0.40, Y: 0.40, W: 0.10, H: 0.10}}
		detected := face.Faces{
			{Rows: 100, Cols: 100, Score: 95, Area: face.NewArea("face", 45, 45, 18)},
			{Rows: 100, Cols: 100, Score: 60, Area: face.NewArea("face", 46, 46, 10)},
		}

		assert.Equal(t, 0, markers.MatchFaces(detected)["m1"], "MatchFaces keeps ranking by marker overlap")
		assert.Equal(t, 1, markers.MatchFacesBestFit(detected, nil)["m1"])
	})
	t.Run("SkipsTaken", func(t *testing.T) {
		markers := Markers{{MarkerUID: "m1", X: 0.1, Y: 0.1, W: 0.2, H: 0.2}}
		detected := face.Faces{{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)}}

		assert.Empty(t, markers.MatchFacesBestFit(detected, map[int]bool{0: true}))
		assert.Equal(t, map[string]int{"m1": 0}, markers.MatchFacesBestFit(detected, map[int]bool{}))
	})
	t.Run("NoMarkers", func(t *testing.T) {
		assert.Empty(t, Markers(nil).MatchFacesBestFit(face.Faces{{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)}}, nil))
	})
}

func TestFaceMatchIoU(t *testing.T) {
	t.Run("Same", func(t *testing.T) {
		a := crop.Area{X: 0.1, Y: 0.1, W: 0.2, H: 0.2}
		assert.InDelta(t, 1.0, faceMatchIoU(a, a), 0.0001)
	})
	t.Run("Contained", func(t *testing.T) {
		assert.InDelta(t, 0.25, faceMatchIoU(crop.Area{X: 0, Y: 0, W: 0.4, H: 0.4}, crop.Area{X: 0.1, Y: 0.1, W: 0.2, H: 0.2}), 0.0001)
	})
	t.Run("Disjoint", func(t *testing.T) {
		assert.Zero(t, faceMatchIoU(crop.Area{X: 0, Y: 0, W: 0.1, H: 0.1}, crop.Area{X: 0.5, Y: 0.5, W: 0.1, H: 0.1}))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Zero(t, faceMatchIoU(crop.Area{}, crop.Area{}))
	})
}
