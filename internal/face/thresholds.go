package face

import (
	"github.com/photoprism/photoprism/internal/crop"
)

var CropSize = crop.Sizes[crop.Tile160]
var ClusterCore = 4
var ClusterMinScore = 15
var ClusterMinSize = 95
var ClusterRadius = 0.69
var MatchRadius = 0.46
var SampleThreshold = 2 * ClusterCore
var OverlapThreshold = 42
var OverlapThresholdFloor = OverlapThreshold - 1
var ScoreThreshold = float32(9.0)

// ScaleScoreThreshold returns the scale adjusted face score threshold.
func ScaleScoreThreshold(scale int) float32 {
	// Smaller faces require higher quality.
	switch {
	case scale < 26:
		return ScoreThreshold + 26.0
	case scale < 32:
		return ScoreThreshold + 16.0
	case scale < 40:
		return ScoreThreshold + 11.0
	case scale < 50:
		return ScoreThreshold + 9.0
	case scale < 80:
		return ScoreThreshold + 6.0
	case scale < 110:
		return ScoreThreshold + 2.0
	}

	return ScoreThreshold
}
