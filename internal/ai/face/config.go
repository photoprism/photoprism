package face

import (
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

// Thresholds tuned for the FaceNet embedding space PhotoPrism has shipped since 2021.
// They are the fallback whenever no embedding model is configured, and stay separate
// from the variables below because Config.Propagate overwrites those at runtime.
const (
	// ClusterDistDefault is the default distance threshold that defines the cluster core.
	ClusterDistDefault = 0.64
	// ClusterRadiusDefault is the default maximum normalized distance for cluster samples.
	ClusterRadiusDefault = 0.42
	// MatchDistDefault is the default distance offset used to match faces with clusters.
	MatchDistDefault = 0.4
	// CollisionDistDefault is the default distance below which embeddings cannot be distinguished.
	CollisionDistDefault = 0.05
	// EpsilonDefault is the default numeric tolerance used during cluster comparisons.
	EpsilonDefault = 0.01
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
	ClusterScoreThreshold = 20
	// SizeThreshold is the minimum detected face size, in pixels.
	SizeThreshold = 25
	// ClusterSizeThreshold is the minimum face size, in pixels, for faces considered when forming clusters.
	ClusterSizeThreshold = 60
	// ClusterDist is the similarity distance threshold that defines the cluster core.
	ClusterDist = ClusterDistDefault
	// ClusterRadius is the maximum normalized distance for cluster samples.
	ClusterRadius = ClusterRadiusDefault
	// MatchDist is the distance offset threshold used to match new faces with existing clusters.
	MatchDist = MatchDistDefault
	// CollisionDist is the minimum distance under which embeddings cannot be distinguished.
	CollisionDist = CollisionDistDefault
	// ClusterCore is the minimum number of faces required to seed a cluster core.
	ClusterCore = 4
	// SampleThreshold is the number of faces required before automatic clustering begins.
	SampleThreshold = 2 * ClusterCore
	// Epsilon is the numeric tolerance used during cluster comparisons.
	Epsilon = EpsilonDefault
)
