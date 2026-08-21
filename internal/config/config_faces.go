package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// FaceEngine returns the configured face detection engine. When the config is
// nil or the vision subsystem is not initialized it reports `face.EngineNone`
// so callers can short-circuit gracefully.
func (c *Config) FaceEngine() string {
	if c == nil {
		return face.EngineNone
	} else if c.options.FaceEngine == face.EngineONNX || c.options.FaceEngine == face.EngineNone {
		return c.options.FaceEngine
	}

	if vision.Config == nil {
		return face.EngineNone
	}

	desired := face.ParseEngine(c.options.FaceEngine)
	modelPath := c.FaceEngineModelPath()

	if desired == face.EngineAuto {
		if _, err := os.Stat(modelPath); err == nil {
			desired = face.EngineONNX
		} else {
			desired = face.EngineNone
		}

		c.options.FaceEngine = desired
	}

	return desired
}

// FaceEngineRunType returns the effective run type for the face detection engine.
// Detection and embedding always run together, so we defer to the face model
// configuration in the vision subsystem. If no detection model is configured,
// or faces are disabled entirely, the run type falls back to RunNever.
func (c *Config) FaceEngineRunType() vision.RunType {
	if c == nil {
		return vision.RunNever
	}

	if vision.Config == nil {
		return vision.RunNever
	}

	if c.DisableFaces() || c.FaceEngine() == face.EngineNone {
		return vision.RunNever
	}

	return vision.Config.RunType(vision.ModelTypeFace)
}

// FaceEngineShouldRun reports whether the face detection engine should execute in the
// specified scheduling context. The decision mirrors the face model run schedule in
// the vision subsystem, so detection stays aligned with embedding generation.
func (c *Config) FaceEngineShouldRun(when vision.RunType) bool {
	if c == nil {
		return false
	}

	if c.DisableFaces() || c.FaceEngine() == face.EngineNone {
		return false
	}

	run := c.FaceEngineRunType()
	when = vision.ParseRunType(when)

	switch run {
	case vision.RunNever:
		return false
	case vision.RunManual:
		return when == vision.RunManual
	case vision.RunAlways:
		return when != vision.RunNever
	case vision.RunNewlyIndexed:
		return when == vision.RunManual || when == vision.RunNewlyIndexed || when == vision.RunOnDemand
	case vision.RunOnDemand:
		return when == vision.RunAuto || when == vision.RunManual || when == vision.RunNewlyIndexed || when == vision.RunOnDemand
	case vision.RunOnSchedule:
		return when == vision.RunAuto || when == vision.RunManual || when == vision.RunOnSchedule || when == vision.RunOnDemand
	case vision.RunOnIndex:
		return when == vision.RunManual || when == vision.RunOnIndex
	case vision.RunAuto:
		fallthrough
	default:
		switch when {
		case vision.RunAuto, vision.RunAlways, vision.RunManual, vision.RunOnDemand:
			return true
		case vision.RunOnIndex:
			return c.faceEngineRunsOnIndex()
		case vision.RunNewlyIndexed:
			return !c.faceEngineRunsOnIndex()
		case vision.RunOnSchedule, vision.RunNever:
			return false
		}
	}

	return false
}

// FaceEngineThreads returns the thread count for ONNX face detection.
//
// The automatic value divides the cores by the number of indexing workers, because face
// detection takes no lock: that many detections run at once, each with its own thread
// pool, so a per-session count derived from the cores alone oversubscribes the machine
// by exactly that factor. The derived value is not written back to the options, so that
// FaceModelThreads keeps deriving its own count.
func (c *Config) FaceEngineThreads() int {
	if c == nil {
		return 1
	} else if c.options.FaceEngineThreads > 0 {
		return c.options.FaceEngineThreads
	}

	return max(runtime.NumCPU()/max(c.IndexWorkers(), 1), 1)
}

// FaceModelThreads returns the thread count for face embedding inference.
//
// Embeddings are generated one at a time behind the model session lock, so unlike
// detection they never run once per indexing worker and keep the undivided count.
func (c *Config) FaceModelThreads() int {
	if c == nil {
		return 1
	} else if c.options.FaceEngineThreads > 0 {
		return c.options.FaceEngineThreads
	}

	return max(runtime.NumCPU()/2, 1)
}

// faceEngineRunsOnIndex reports whether this host is fast enough to detect faces while
// indexing rather than deferring them to the pass over newly indexed files.
//
// It reads the count that is not divided among the indexing workers, because that
// divisor follows the database driver and the available memory, which would tie the
// schedule to the storage backend instead of to the capability of the machine.
func (c *Config) faceEngineRunsOnIndex() bool {
	return c.FaceModelThreads() > 2
}

// FaceEngineModelPath returns the absolute path to the bundled SCRFD ONNX detector.
func (c *Config) FaceEngineModelPath() string {
	if c == nil {
		return ""
	}

	dir := filepath.Join(c.ModelsPath(), "scrfd")
	primary := filepath.Join(dir, face.DefaultONNXModelFilename)

	if _, err := os.Stat(primary); err == nil {
		return primary
	}

	alt := filepath.Join(dir, "scrfd_500m_bnkps_shape640x640.onnx")

	if _, err := os.Stat(alt); err == nil {
		return alt
	}

	return primary
}

// FaceModel returns the name of the configured face embedding model. Unsupported
// values are reported and treated as `face.ModelAuto`, which keeps the embedding space
// a library already uses and otherwise resolves to the first installed model in
// `face.AutoModelPreference`.
func (c *Config) FaceModel() string {
	if c == nil {
		return face.ModelNone
	}

	if c.options.FaceModel == "" {
		c.options.FaceModel = face.ModelAuto
	} else if !face.KnownModelName(c.options.FaceModel) {
		log.Warnf("config: unsupported face model %s, expected %s", clean.Log(c.options.FaceModel), face.ModelUsageString())
		c.options.FaceModel = face.ModelAuto
	}

	modelsPath := c.ModelsPath()

	if name := face.ParseModelName(c.options.FaceModel); name == face.ModelNone {
		return name
	} else if name != face.ModelAuto {
		if face.FindEmbeddingModel(name).Installed(modelsPath) {
			return name
		}

		// Falling forward to another model would start a second vector space the library
		// cannot compare with, and an image upgrade removes opt-in models from assets, so
		// an explicitly requested model that is missing disables embeddings instead.
		log.Warnf("config: face model %s is not installed, disabling face embeddings", clean.Log(name))

		return face.ModelNone
	}

	resolved := face.ModelNone

	// An existing library keeps the space its vectors were generated in. Resolving to a
	// different model would leave every stored cluster incomparable with everything
	// indexed from now on, which is why "auto" asks the library before the preference list.
	if name := c.libraryFaceModel(); name == "" {
		resolved = installedFaceModel(modelsPath)
	} else if face.FindEmbeddingModel(name).Installed(modelsPath) {
		resolved = name
	} else {
		// Anything indexed from now on would land in a second vector space that cannot be
		// compared with what the library already holds, so embeddings stop until the model
		// is reinstalled or the library is migrated.
		log.Warnf("config: library was indexed with face model %s, which is not installed, disabling face embeddings", clean.Log(name))
		resolved = face.ModelNone
	}

	// Cache the resolved name so the lookup, the query, and any warning only happen once.
	// An answer found before the database was connected could not consult the library,
	// so it stays provisional rather than freezing the preference list into place.
	if c.db != nil {
		c.options.FaceModel = resolved
	}

	return resolved
}

// installedFaceModel returns the first model in AutoModelPreference whose weights exist
// in the specified path, or ModelNone when none of them are installed.
func installedFaceModel(modelsPath string) face.ModelName {
	for _, candidate := range face.AutoModelPreference {
		if face.FindEmbeddingModel(candidate).Installed(modelsPath) {
			return candidate
		}
	}

	return face.ModelNone
}

// libraryFaceModel returns the embedding model the library's face vectors were generated
// with, or an empty name when it holds none and when the database is not connected yet.
//
// This runs once per process, including for every CLI invocation, so it asks the indexed
// provenance column first and only falls back to counting vectors when nothing answers.
func (c *Config) libraryFaceModel() face.ModelName {
	if c == nil || c.db == nil {
		return ""
	}

	counts, err := query.RecordedMarkerEmbeddingModels()

	if err == nil && len(counts) > 0 {
		return dominantFaceModel(counts)
	} else if err != nil {
		// The schema is migrated after the configuration is propagated, so on the first
		// start after an upgrade the provenance column does not exist yet.
		log.Debugf("config: %s (find face embedding models)", err)
	}

	// A library whose markers record no model cannot hold anything but FaceNet vectors,
	// which is what its face markers prove.
	if markers, countErr := query.FaceMarkersWithVectors(); countErr != nil {
		log.Debugf("config: %s (count face markers)", countErr)
	} else if markers > 0 {
		return face.ModelFaceNet
	}

	return ""
}

// dominantFaceModel returns the model that produced most of the counted vectors, or an
// empty name when none were counted. What is counted depends on the caller: the recorded
// provenance alone, or every vector, in which case a blank name means the vector predates
// the provenance column and is therefore FaceNet.
func dominantFaceModel(counts []query.MarkerEmbeddingModelCount) face.ModelName {
	var name face.ModelName
	var markers int

	for _, count := range counts {
		if count.Markers > markers {
			name, markers = face.NormalizeModelName(count.EmbedModel), count.Markers
		}
	}

	if markers == 0 {
		return ""
	} else if name == "" {
		return face.ModelFaceNet
	}

	return name
}

// FaceEmbeddingModel returns the resolved face embedding model, or nil when
// embeddings are disabled or no model is installed.
func (c *Config) FaceEmbeddingModel() *face.EmbeddingModel {
	return face.FindEmbeddingModel(c.FaceModel())
}

// FaceModelPath returns the absolute path of the configured face embedding model.
func (c *Config) FaceModelPath() string {
	if c == nil {
		return ""
	}

	return c.FaceEmbeddingModel().FilePath(c.ModelsPath())
}

// FaceModelLicense returns the license of the configured face embedding model weights,
// or an empty string when no model is configured.
func (c *Config) FaceModelLicense() string {
	return c.FaceEmbeddingModel().WeightLicense()
}

// FaceModelDims returns the embedding length of the configured face model, or 0 when none is configured.
func (c *Config) FaceModelDims() int {
	if m := c.FaceEmbeddingModel(); m != nil {
		return m.Dims
	}

	return 0
}

// FaceSize returns the face size threshold in pixels.
func (c *Config) FaceSize() int {
	if c.options.FaceSize < 20 || c.options.FaceSize > 10000 {
		return face.SizeThreshold
	}

	return c.options.FaceSize
}

// FaceScore returns the face quality score threshold.
func (c *Config) FaceScore() float64 {
	if c.options.FaceScore < 1 || c.options.FaceScore > 100 {
		return face.ScoreThreshold
	}

	return c.options.FaceScore
}

// FaceOverlap returns the face area overlap threshold in percent.
func (c *Config) FaceOverlap() int {
	if c.options.FaceOverlap < 1 || c.options.FaceOverlap > 100 {
		return face.OverlapThreshold
	}

	return c.options.FaceOverlap
}

// FaceClusterSize returns the size threshold for faces forming a cluster in pixels.
func (c *Config) FaceClusterSize() int {
	if c.options.FaceClusterSize < 20 || c.options.FaceClusterSize > 10000 {
		return face.ClusterSizeThreshold
	}

	return c.options.FaceClusterSize
}

// FaceClusterScore returns the quality threshold for faces forming a cluster.
func (c *Config) FaceClusterScore() int {
	if c.options.FaceClusterScore < 1 || c.options.FaceClusterScore > 100 {
		return face.ClusterScoreThreshold
	}

	return c.options.FaceClusterScore
}

// FaceClusterCore returns the number of faces forming a cluster core.
func (c *Config) FaceClusterCore() int {
	if c.options.FaceClusterCore < 1 || c.options.FaceClusterCore > 100 {
		return face.ClusterCore
	}

	return c.options.FaceClusterCore
}

// FaceClusterDist returns the radius of faces forming a cluster core.
func (c *Config) FaceClusterDist() float64 {
	return c.faceThreshold("face-cluster-dist", c.options.FaceClusterDist, face.ClusterDistDefault,
		func(m *face.EmbeddingModel) float64 { return m.ClusterDist })
}

// FaceClusterRadius returns the maximum radius used when matching face clusters.
func (c *Config) FaceClusterRadius() float64 {
	return c.faceThreshold("face-cluster-radius", c.options.FaceClusterRadius, face.ClusterRadiusDefault,
		func(m *face.EmbeddingModel) float64 { return m.ClusterRadius })
}

// FaceCollisionDist returns the minimum distance used to differentiate embeddings.
//
// It does not go through faceThreshold, which takes this value as its lower bound and
// would recurse.
func (c *Config) FaceCollisionDist() float64 {
	value := c.options.FaceCollisionDist
	configured := c.faceThresholdIsSet("face-collision-dist", value, face.CollisionDistDefault)

	if value > 0 && value <= 1 && configured {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(),
		func(m *face.EmbeddingModel) float64 { return m.CollisionDist }, face.CollisionDistDefault)

	// Zero means "use the model default", so only a real value can be out of range.
	c.warnFaceThreshold(configured && value != 0, "face-collision-dist", value, 0, 1, resolved)

	return resolved
}

// FaceEpsilonDist returns the distance slack applied to collision checks.
func (c *Config) FaceEpsilonDist() float64 {
	value := c.options.FaceEpsilonDist
	configured := c.faceThresholdIsSet("face-epsilon-dist", value, face.EpsilonDefault)

	if value > 0 && value <= 0.1 && configured {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(),
		func(m *face.EmbeddingModel) float64 { return m.Epsilon }, face.EpsilonDefault)

	c.warnFaceThreshold(configured && value != 0, "face-epsilon-dist", value, 0, 0.1, resolved)

	return resolved
}

// FaceMatchDist returns the offset distance when matching faces with clusters.
func (c *Config) FaceMatchDist() float64 {
	return c.faceThreshold("face-match-dist", c.options.FaceMatchDist, face.MatchDistDefault,
		func(m *face.EmbeddingModel) float64 { return m.MatchDist })
}

// faceThreshold returns the operator-configured clustering threshold, or the value
// calibrated for the configured embedding model when the option was left untouched.
func (c *Config) faceThreshold(flagName string, value, flagDefault float64, pick func(*face.EmbeddingModel) float64) float64 {
	minDist := c.FaceCollisionDist()
	configured := c.faceThresholdIsSet(flagName, value, flagDefault)

	if configured && value >= minDist && value <= face.AcceptDistMax {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(), pick, flagDefault)

	c.warnFaceThreshold(configured, flagName, value, minDist, face.AcceptDistMax, resolved)

	return resolved
}

// warnFaceThreshold reports an out-of-range option once. A value that is silently replaced
// looks like a setting that had no effect, and the getters are called from Propagate and
// the config report, so the warning is emitted once per option rather than per call.
func (c *Config) warnFaceThreshold(configured bool, flagName string, value, minValue, maxValue, resolved float64) {
	if !configured {
		return
	}

	if _, warned := c.faceWarned.LoadOrStore(flagName, true); !warned {
		log.Warnf("config: %s %g is out of range (%g-%g), using %g instead", flagName, value, minValue, maxValue, resolved)
	}
}

// faceThresholdIsSet reports whether an operator configured a clustering threshold explicitly.
// The value alone cannot answer this, because the CLI flags carry the FaceNet defaults so that
// "photoprism --help" documents them, and those defaults are copied into the options.
func (c *Config) faceThresholdIsSet(flagName string, value, flagDefault float64) bool {
	if c.cliCtx != nil && c.cliCtx.IsSet(flagName) {
		return true
	}

	// Values loaded from "options.yml" are applied after the CLI context, so anything
	// that differs from the flag default was configured deliberately.
	return value != flagDefault
}

// faceModelThreshold returns a clustering threshold in the configured model's distance
// scale, falling back to the FaceNet-tuned default when no model is configured or the
// model carries no calibrated value.
func faceModelThreshold(m *face.EmbeddingModel, pick func(*face.EmbeddingModel) float64, fallback float64) float64 {
	if m == nil {
		return fallback
	}

	if v := pick(m); v > 0 {
		return v
	}

	return fallback
}
