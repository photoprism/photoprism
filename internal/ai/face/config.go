package face

import (
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

var (
	// CropSize is the face image crop size used when generating FaceNet embeddings.
	CropSize = crop.Sizes[crop.Tile160]
)

var (
	// OverlapThreshold defines the minimum face area overlap percentage required to treat detections as identical.
	OverlapThreshold = 42
	// OverlapThresholdFloor is the relaxed overlap threshold used to avoid rounding inconsistencies.
	OverlapThresholdFloor = OverlapThreshold - 1
	// ScoreThreshold is the base minimum face score accepted by the detector.
	ScoreThreshold = 9.0
	// ClusterScoreThreshold is the minimum score required for faces that contribute to automatic clustering.
	ClusterScoreThreshold = 15
	// SizeThreshold is the minimum detected face size, in pixels.
	SizeThreshold = 25
	// ClusterSizeThreshold is the minimum face size, in pixels, for faces considered when forming clusters.
	ClusterSizeThreshold = 50
	// ClusterDist is the similarity distance threshold that defines the cluster core.
	ClusterDist = 0.64
	// ClusterRadius is the maximum normalized distance for cluster samples.
	ClusterRadius = 0.42
	// MatchDist is the distance offset threshold used to match new faces with existing clusters.
	MatchDist = 0.42
	// CollisionDist is the minimum distance under which embeddings cannot be distinguished.
	CollisionDist = 0.1
	// ClusterCore is the minimum number of faces required to seed a cluster core.
	ClusterCore = 4
	// SampleThreshold is the number of faces required before automatic clustering begins.
	SampleThreshold = 2 * ClusterCore
	// Epsilon is the numeric tolerance used during cluster comparisons.
	Epsilon = 0.01
	// SkipChildren controls whether the clustering step omits faces from child samples by default.
	SkipChildren = true
	// IgnoreBackground determines whether background faces are ignored when generating matches.
	IgnoreBackground = true
)

var (
	// LandmarkQualityFloor is the minimum score accepted when both eyes are located by the landmark detector.
	LandmarkQualityFloor = float32(5.0)
	// LandmarkQualityScaleMin is the minimum face size eligible for the landmark-assisted quality fallback.
	LandmarkQualityScaleMin = 60
	// LandmarkQualityScaleMax is the maximum face size eligible for the landmark-assisted quality fallback.
	LandmarkQualityScaleMax = 90
	// LandmarkQualitySlack is the maximum allowed difference between the quality threshold and the detected score.
	LandmarkQualitySlack = float32(4.0)
)

// PigoQualityThreshold returns the scale-adjusted minimum Pigo quality score threshold for the provided detection scale.
func PigoQualityThreshold(scale int) (score float32) {
	score = float32(ScoreThreshold)

	// Smaller faces require higher quality.
	switch {
	case scale < 26:
		score += 12.0
	case scale < 32:
		score += 8.0
	case scale < 40:
		score += 6.0
	case scale < 50:
		score += 4.0
	case scale < 80:
		score += 2.0
	case scale < 110:
		score += 1.0
	}

	return score
}
