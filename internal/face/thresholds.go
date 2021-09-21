package face

import (
	"github.com/photoprism/photoprism/internal/crop"
)

var CropSize = crop.Sizes[crop.Tile160]
var ClusterCore = 4
var ClusterRadius = 0.58
var ClusterMinScore = 15
var ClusterMinSize = 100
var SampleThreshold = 2 * ClusterCore
var OverlapThreshold = 42
var OverlapThresholdFloor = OverlapThreshold - 1
var ScoreThreshold = float32(9.0)

// ScaleScoreThreshold returns the scale adjusted face score threshold.
func ScaleScoreThreshold(scale int) float32 {
	// Smaller faces require higher quality.
	switch {
	case scale <= 25:
		return ScoreThreshold + 21.0
	case scale < 30:
		return ScoreThreshold + 12.5
	case scale < 50:
		return ScoreThreshold + 9.5
	case scale < 80:
		return ScoreThreshold + 6.0
	case scale < 110:
		return ScoreThreshold + 2.0
	}

	return ScoreThreshold
}
