package face

import (
	"github.com/photoprism/photoprism/internal/crop"
)

var CropSize = crop.Sizes[crop.Tile160]
var MatchDist = 0.46
var ClusterDist = 0.64
var ClusterCore = 4
var ClusterMinScore = 15
var ClusterMinSize = 80
var SampleThreshold = 2 * ClusterCore
var OverlapThreshold = 42
var OverlapThresholdFloor = OverlapThreshold - 1
var ScoreThreshold = 9.0

// QualityThreshold returns the scale adjusted quality score threshold.
func QualityThreshold(scale int) (score float32) {
	score = float32(ScoreThreshold)

	// Smaller faces require higher quality.
	switch {
	case scale < 26:
		score += 26.0
	case scale < 32:
		score += 16.0
	case scale < 40:
		score += 11.0
	case scale < 50:
		score += 9.0
	case scale < 80:
		score += 6.0
	case scale < 110:
		score += 2.0
	}

	return score
}
