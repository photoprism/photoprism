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

// InterOpThreads is how many threads an ONNX session may use to run graph nodes in
// parallel. Inter-op parallelism only pays off for graphs with independent branches, and
// the detection and embedding graphs are sequential CNNs, so a second thread would add
// scheduling overhead and compete with the intra-op pool for nothing.
const InterOpThreads = 1

// AcceptDistMax is the highest distance at which an embedding may still join a cluster, whatever a
// stored sample radius says. Two independent unit vectors average sqrt(2) ~ 1.41 apart and a sizable
// share land nearer, so this caps what a misconfiguration can do rather than naming a safe value.
const AcceptDistMax = 1.4

// ConfigDistMax is the largest value a configurable distance threshold may take.
//
// It sits just above the widest calibrated model, so what can be configured is the range the models
// were calibrated in; past it a cluster accepts about as readily as it refuses. A value above is
// refused rather than clipped on read, which would leave the report echoing a number that never applies.
const ConfigDistMax = 1.25

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
	// RetrySizeThreshold is the minimum face size, in pixels, used by the second pass that runs
	// only when the first finds nothing. Crowd photographs reduce every face below SizeThreshold,
	// so without it a frame full of people is indexed as containing none.
	RetrySizeThreshold = 10
	// MinSizeThreshold is the smallest configurable face size. It is where the detectors stop
	// being trained rather than a policy choice: YuNet states a lower bound of about ten pixels,
	// so a smaller setting would ask for faces no model in the registry can find.
	MinSizeThreshold = 10
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

// ClampSampleRadius limits a cluster sample radius to the configured range.
func ClampSampleRadius(radius float64) float64 {
	if radius > ClusterRadius {
		return ClusterRadius
	} else if radius < 0 {
		return 0
	}

	return radius
}

// AcceptDist returns the distance below which an embedding joins a cluster with the
// specified sample radius. Clamping here rather than only where the radius is stored
// keeps a recalibrated ClusterRadius authoritative for rows written by an earlier one.
func AcceptDist(sampleRadius float64) float64 {
	if dist := ClampSampleRadius(sampleRadius) + MatchDist; dist < AcceptDistMax {
		return dist
	}

	return AcceptDistMax
}
