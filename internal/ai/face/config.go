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
	// SizeThresholdDefault is the default minimum detected face size, in pixels.
	SizeThresholdDefault = 25
	// ScoreThresholdDefault leaves the cutoff to the detector, on the 0-100 confidence scale the
	// ONNX detectors report. A shared default either sits below every detector's own cutoff and
	// gates nothing, or above one of them and silently overrules its calibration.
	ScoreThresholdDefault = 0.0
	// NoScoreThreshold applies no score cutoff at all, following the convention the other numeric
	// options use for "switched off". It is distinct from ScoreThresholdDefault, which cannot
	// express this: zero already means "let the detector decide".
	NoScoreThreshold = -1.0
	// MigrationScoreThreshold is the detection floor a face embedding migration runs at, on the
	// 0-100 scale, unless an operator configured one. Re-embedding keeps a marker only when the
	// detector finds its face again, so a miss here discards a curated marker instead of adding
	// a false positive to an index - the opposite trade to the one indexing makes.
	MigrationScoreThreshold = 9.0
	// ClusterScoreThresholdDefault is the clustering bar for a detector that registers none, and
	// for a marker no detector produced.
	ClusterScoreThresholdDefault = 20
	// ClusterScoreAuto asks a query to apply each marker's own detector bar, as distinct from zero,
	// which asks for no score filter at all.
	ClusterScoreAuto = -1
	// OverlapThresholdDefault is the default face area overlap percentage above which two
	// detections are treated as identical.
	OverlapThresholdDefault = 42
	// ClusterSizeThresholdDefault is the default minimum face size, in pixels, for clustering.
	ClusterSizeThresholdDefault = 60
	// ClusterCoreDefault is the default number of faces required to seed a cluster core.
	ClusterCoreDefault = 4
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
	OverlapThreshold = OverlapThresholdDefault
	// OverlapThresholdFloor is the relaxed overlap threshold used to avoid rounding inconsistencies.
	OverlapThresholdFloor = OverlapThreshold - 1
	// ScoreThreshold is the base minimum face score accepted by the detector.
	ScoreThreshold = ScoreThresholdDefault
	// ClusterScoreThreshold is the operator's own minimum score for faces that contribute to
	// automatic clustering, assigned by Config.Propagate from FACE_CLUSTER_SCORE. Zero leaves the
	// bar to the detector that scored each marker, and negative removes it. It is a weak gate
	// either way, because detector confidence saturates above the detector's own cutoff;
	// ClusterSizeThreshold is what keeps an interpolated crop out.
	ClusterScoreThreshold = 0
	// SizeThreshold is the minimum detected face size, in pixels. Config.Propagate assigns it from
	// FACE_SIZE, so a reader that bounds what detection produced compares against the same value.
	SizeThreshold = SizeThresholdDefault
	// RetrySizeThreshold is the minimum face size, in pixels, used by the second pass that runs
	// only when the first finds nothing. Crowd photographs reduce every face below SizeThreshold,
	// so without it a frame full of people is indexed as containing none.
	RetrySizeThreshold = 10
	// MinSizeThreshold is the smallest configurable face size. It is where the detectors stop
	// being trained rather than a policy choice: YuNet states a lower bound of about ten pixels,
	// so a smaller setting would ask for faces no model in the registry can find.
	MinSizeThreshold = 10
	// ClusterSizeThreshold is the minimum face size, in pixels, for faces considered when forming clusters.
	ClusterSizeThreshold = ClusterSizeThresholdDefault
	// ClusterDist is the similarity distance threshold that defines the cluster core.
	ClusterDist = ClusterDistDefault
	// ClusterRadius is the maximum normalized distance for cluster samples.
	ClusterRadius = ClusterRadiusDefault
	// MatchDist is the distance offset threshold used to match new faces with existing clusters.
	MatchDist = MatchDistDefault
	// CollisionDist is the minimum distance under which embeddings cannot be distinguished.
	CollisionDist = CollisionDistDefault
	// ClusterCore is the minimum number of faces required to seed a cluster core.
	ClusterCore = ClusterCoreDefault
	// SampleThreshold is the number of faces required before automatic clustering begins.
	SampleThreshold = 2 * ClusterCore
	// Epsilon is the numeric tolerance used during cluster comparisons.
	Epsilon = EpsilonDefault
)

// ClusterScore returns the score a marker the named detector produced has to reach to contribute
// to automatic clustering.
//
// FACE_CLUSTER_SCORE outranks the per-detector bars, and a negative value removes them, because
// a value an operator chose is not a calibration a marker was never scored against. Unset, the
// bar is per marker: two detectors' scores are not comparable, and a library holds markers from
// both. An unregistered detector and a row predating the column keep the shared default.
func ClusterScore(detector DetectorName) int {
	if ClusterScoreThreshold != 0 {
		return max(ClusterScoreThreshold, 0)
	}

	if d := FindDetector(detector); d != nil && d.ClusterMinScore > 0 {
		return d.ClusterMinScore
	}

	return ClusterScoreThresholdDefault
}

// DetectorScore returns the detection cutoff calibrated for the named detector, on the 0-100
// scale. An unregistered name falls back to the default detector, so a report always states a
// cutoff that some detector actually enforces rather than zero.
func DetectorScore(detector DetectorName) float64 {
	d := FindDetector(detector)

	if d == nil {
		d = DefaultDetector()
	}

	if d == nil || d.MinScore <= 0 {
		return ScoreThresholdDefault
	}

	return float64(d.MinScore)
}

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
