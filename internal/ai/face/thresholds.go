package face

import (
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

var CropSize = crop.Sizes[crop.Tile160]          // Face image crop size for FaceNet.
var OverlapThreshold = 42                        // Face area overlap threshold in percent.
var OverlapThresholdFloor = OverlapThreshold - 1 // Reduced overlap area to avoid rounding inconsistencies.
var ScoreThreshold = 9.0                         // Min face score.
var LandmarkQualityFloor = float32(5.0)          // Min score when both eyes are located.
var LandmarkQualityScaleMin = 60                 // Min face size eligible for landmark-based quality fallback.
var LandmarkQualityScaleMax = 90                 // Max face size eligible for landmark-based quality fallback.
var LandmarkQualitySlack = float32(4.0)          // Max allowed gap between quality threshold and score.
var ClusterScoreThreshold = 15                   // Min score for faces forming a cluster.
var SizeThreshold = 25                           // Min face size in pixels.
var ClusterSizeThreshold = 50                    // Min size for faces forming a cluster in pixels.
var ClusterDist = 0.64                           // Similarity distance threshold of faces forming a cluster core.
var MatchDist = 0.46                             // Dist offset threshold for matching new faces with clusters.
var ClusterCore = 4                              // Min number of faces forming a cluster core.
var SampleThreshold = 2 * ClusterCore            // Threshold for automatic clustering to start.

// QualityThreshold returns the scale adjusted quality score threshold.
func QualityThreshold(scale int) (score float32) {
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
