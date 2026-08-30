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
	// CollisionDistDefault is the default floor below which a recorded collision radius is discarded.
	CollisionDistDefault = 0.05
	// MatchMarginDefault is the default distance by which the nearest cluster has to beat the
	// runner-up for the marker to be given to it. Defense in depth rather than a recall lever:
	// measured against a labeled library it decides two markers, the ambiguity rule deciding the
	// rest, so the value is kept small enough not to suppress an assignment that rule earns.
	MatchMarginDefault = 0.01
	// NoMatchMargin assigns a marker to its nearest cluster however narrowly that one wins,
	// following the convention the score bars use for "switched off".
	NoMatchMargin = -1.0
	// EpsilonDefault is the numeric tolerance used during cluster comparisons, and the same for
	// every model unlike the calibrated distances: it is the gap a resolved collision leaves, which
	// is a void where nothing matches, so a wider one strands embeddings rather than separating anyone.
	EpsilonDefault = 0.001
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
	// ClusterScoreThresholdDefault is the clustering bar for a detector that registers none, and
	// for a marker no detector produced.
	ClusterScoreThresholdDefault = 20
	// ClusterScoreAuto asks a query to apply each marker's own detector bar, as distinct from zero,
	// which asks for no score filter at all.
	ClusterScoreAuto = -1
	// OverlapThresholdDefault is the default face area overlap percentage above which two
	// detections are treated as identical.
	OverlapThresholdDefault = 42
	// ClusterSizeThresholdDefault is the default minimum face size, in pixels, for clustering, and
	// bounds the source pixels an embedding rests on. Measured rather than derived: quality turns
	// at the aligned crop size and is flat above it.
	ClusterSizeThresholdDefault = ArcFaceTemplateSize
	// ClusterCoreDefault is the default number of faces required to seed a cluster core. DBSCAN
	// counts the point itself, so a person with fewer clusterable faces forms no cluster at all.
	ClusterCoreDefault = 5
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

// RadiusPercentile is the share of a cluster's member distances its radius has to cover. Taking
// the maximum instead lets one loose member decide how far a whole cluster reaches, with only the
// clamp to stop it; a small group's percentile and maximum barely differ, so this narrows the wide
// clusters it is aimed at rather than every cluster.
const RadiusPercentile = 95

// ConfigDistMax is the largest value a configurable distance threshold may take.
//
// It sits just above the widest calibrated model, so what can be configured is the range the models
// were calibrated in; past it a cluster accepts about as readily as it refuses. A value above is
// refused rather than clipped on read, which would leave the report echoing a number that never applies.
const ConfigDistMax = 1.25

// EpsilonDistMax is the widest configurable collision tolerance. Twice it bounds AmbiguityDist, so
// the cutoff under which a colliding cluster is retired can be narrowed but never widened past it.
const EpsilonDistMax = 0.01

var (
	// CropSize is the rectangular face crop, used by FaceNet, by the fallback for a face whose
	// landmarks are incomplete, and for the crop the UI displays. It is not what the ONNX models
	// consume: those are warped onto ArcFaceTemplateSize directly from the source rendition.
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
	// bar to the detector that scored each marker, and negative removes it. It gates weakly either
	// way, since confidence saturates above a cutoff; ClusterSizeThreshold keeps blurry crops out.
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
	// CollisionDist is the floor below which a cluster's recorded CollisionRadius is discarded and
	// the cluster keeps its full accept distance: narrowing that far would exclude its own members,
	// so the code stops separating the two and flags the face ambiguous instead.
	CollisionDist = CollisionDistDefault
	// MatchMargin is how much closer the nearest cluster has to be than the runner-up before a
	// marker is assigned to it. A face between two people is large, sharp and confidently scored,
	// so no detection threshold selects against it and only the margin does.
	MatchMargin = MatchMarginDefault
	// ClusterCore is the minimum number of faces required to seed a cluster core.
	ClusterCore = ClusterCoreDefault
	// SampleThreshold is the number of faces required before automatic clustering begins.
	SampleThreshold = 2 * ClusterCore
	// Epsilon is the numeric tolerance used during cluster comparisons.
	Epsilon = EpsilonDefault
)

// ClusterSize returns the minimum face size, in pixels of the image an embedding was sampled from,
// that the named model consumes without interpolating. It is the model's own input geometry, so a
// bar meaning "not invented by interpolation" keeps that meaning when the model changes.
func ClusterSize(model ModelName) int {
	if m := FindEmbeddingModel(model); m != nil {
		if w, h := m.InputSize(); w > 0 && h > 0 {
			return max(w, h)
		} else if !m.Aligned() {
			// Box-aligned models read the rectangular crop instead of the landmark template.
			return CropSize.Width
		}
	}

	return ClusterSizeThresholdDefault
}

// ClusterScore returns the score a marker the named detector produced has to reach to contribute
// to automatic clustering. FACE_CLUSTER_SCORE outranks the per-detector bars and a negative value
// removes them; unset, the bar is per marker, since two detectors' scores are not comparable and
// a library holds markers from both.
func ClusterScore(detector DetectorName) int {
	if ClusterScoreThreshold != 0 {
		return max(ClusterScoreThreshold, 0)
	}

	if d := FindDetector(detector); d != nil && d.ClusterScore > 0 {
		return d.ClusterScore
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

// AmbiguityDist returns the distance below which two embeddings of different subjects are treated
// as the same face rather than as a collision to resolve.
//
// Twice Epsilon, and derived rather than stated: resolving a collision backs a cluster off by
// Epsilon, so below this the backoff would exceed the separation it preserves.
func AmbiguityDist() float64 {
	return 2 * Epsilon
}

// AmbiguousMatch reports whether picking the nearest of two clusters is a coin toss.
//
// The quantity is the difference between two distances rather than a distance: a face between two
// people is admitted by both, and the one that takes it is then widened toward the other.
func AmbiguousMatch(bestDist, runnerUpDist float64) bool {
	if MatchMargin <= 0 || bestDist < 0 || runnerUpDist < 0 {
		return false
	}

	return runnerUpDist-bestDist < MatchMargin
}

// DetectorMigrateScore returns the detection floor a migration re-detects at for the named
// detector, on the 0-100 scale. An unregistered name falls back to the default detector, so a
// caller always gets a floor some detector enforces rather than zero.
func DetectorMigrateScore(detector DetectorName) float64 {
	d := FindDetector(detector)

	if d == nil {
		d = DefaultDetector()
	}

	// A detector registering no migration floor re-detects at its own cutoff: that recovers
	// nothing an index would not have found, which is the safe direction for a value nobody
	// calibrated, and it never returns zero.
	if d == nil || d.MigrateScore <= 0 {
		return DetectorScore(detector)
	}

	return float64(d.MigrateScore)
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
