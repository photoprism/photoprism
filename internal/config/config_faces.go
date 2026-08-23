package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
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
// detection there is one session in total rather than one per indexing worker, and the
// count is not divided among them.
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

// FaceModelSetting returns the face embedding model as configured, without resolving it.
// It reports `face.ModelDetect` when the model is still to be detected, which is also what
// an unsupported value is treated as.
func (c *Config) FaceModelSetting() face.ModelName {
	if c == nil {
		return face.ModelNone
	}

	if !face.KnownModelName(c.options.FaceModel) {
		c.warnFaceModel("face-model", "config: unsupported face model %s, expected %s",
			clean.Log(c.options.FaceModel), face.ModelUsageString())

		return face.ModelDetect
	}

	return face.ParseModelName(c.options.FaceModel)
}

// FaceModel returns the name of the face embedding model in force, or `face.ModelNone` when
// embeddings are disabled, the model cannot be used, or none has been detected yet.
//
// Detection runs once in Config.Init and pins its result, so this reads what was settled there
// rather than asking the library, whose answer is a majority vote that flips mid-migration.
func (c *Config) FaceModel() face.ModelName {
	if c == nil {
		return face.ModelNone
	}

	if c.faceModel != "" {
		return c.faceModel
	}

	switch name := c.FaceModelSetting(); name {
	case face.ModelDetect, face.ModelNone:
		return face.ModelNone
	default:
		return c.usableFaceModel(name)
	}
}

// usableFaceModel returns the specified model when its weights are installed and may be used
// here, and `face.ModelNone` with one warning otherwise.
func (c *Config) usableFaceModel(name face.ModelName) face.ModelName {
	if err := face.LicenseRefused(name, c.Edition()); err != nil {
		c.warnFaceModel("face-model-license", "config: %s, so face embeddings are disabled", err)

		return face.ModelNone
	}

	if !face.FindEmbeddingModel(name).Installed(c.ModelsPath()) {
		// Falling forward to another model would start a second vector space the library
		// cannot compare with, and an image upgrade removes opt-in models from assets, so
		// a model that is missing disables embeddings instead.
		c.warnFaceModel("face-model-installed", "config: face model %s is not installed, disabling face embeddings", clean.Log(name))

		return face.ModelNone
	}

	return name
}

// ResolveFaceModel detects the model from the library without persisting the result, so that a
// report can name what is in force, and whether it can read the library, without changing what
// is configured.
func (c *Config) ResolveFaceModel() face.ModelName {
	if c == nil {
		return face.ModelNone
	} else if c.db == nil {
		return c.FaceModel()
	}

	counts := c.libraryFaceModels()

	if c.faceModel == "" && c.FaceModelSetting() == face.ModelDetect {
		c.faceModel = c.detectFaceModel(counts)
	}

	c.checkFaceModelMismatch(counts)

	return c.FaceModel()
}

// SetFaceModel records the model the library's vectors are now in and persists it.
//
// A migration is the one operation that changes the answer, and the setting has to follow it,
// or a later start hides the whole library from matching. A file that cannot be written costs
// the next start a detection, nothing more.
func (c *Config) SetFaceModel(name face.ModelName) {
	if c == nil || name == "" {
		return
	}

	// Set in memory as well as in the file, because the patch helper only reloads the options
	// it changed and a file that already named the model would leave this process behind.
	c.faceModel = name
	c.options.FaceModel = name

	if _, err := c.SaveOptionsPatch(Values{"FaceModel": name}); err != nil {
		log.Warnf("config: failed saving face model %s (%s)", clean.Log(name), err)
	}
}

// initFaceModel settles which embedding model this instance uses, and is called once by Init
// on a start that has a database.
func (c *Config) initFaceModel() {
	if c == nil || c.db == nil {
		return
	}

	c.reportIgnoredFaceModel()

	// An unsupported value is reported and left alone rather than detected from: writing a
	// detected name would turn a typo into a decision that outlives it.
	if !face.KnownModelName(c.options.FaceModel) {
		log.Warnf("config: unsupported face model %s, expected %s",
			clean.Log(c.options.FaceModel), face.ModelUsageString())

		return
	}

	counts := c.libraryFaceModels()

	if c.FaceModelSetting() == face.ModelDetect {
		if name := c.detectFaceModel(counts); name == face.ModelNone {
			c.faceModel = name
		} else {
			log.Infof("config: detected face model %s", clean.Log(name))
			c.SetFaceModel(name)
		}
	}

	c.checkFaceModelMismatch(counts)
}

// reportIgnoredFaceModel reports a model that an environment variable or the command line set
// while "options.yml" names a different one.
//
// The file wins for every option, and here it also keeps the instance intact: changing the
// model is a data migration, so a variable could only leave the library in two vector spaces.
// An instruction with no effect must still not be silent.
func (c *Config) reportIgnoredFaceModel() {
	if c.faceModelFlag == "" {
		return
	}

	configured := face.ParseModelName(c.faceModelFlag)
	inForce := c.FaceModelSetting()

	if configured == face.ModelDetect || configured == inForce {
		return
	}

	log.Infof("config: face model %s is configured but ignored, this library uses %s "+
		`(run "photoprism faces migrate --to %s" to change it)`, clean.Log(configured), clean.Log(inForce), clean.Log(configured))
}

// detectFaceModel returns the model that produced the library's vectors, or the first installed
// model that may be used when it holds none.
func (c *Config) detectFaceModel(counts []query.MarkerEmbeddingModelCount) face.ModelName {
	// An existing library keeps the space its vectors were generated in, or every stored
	// cluster becomes incomparable with what is indexed from now on. Its answer is not
	// filtered by the preference list either: a model that may not be used is refused with a
	// reason rather than replaced by one that cannot read the vectors.
	if name := dominantFaceModel(counts); name != "" {
		return c.usableFaceModel(name)
	}

	return c.installedFaceModel()
}

// installedFaceModel returns the first model in `face.AutoModelPreference` whose weights are
// installed and may be used here, or `face.ModelNone` when there is none.
//
// License-gated weights are skipped rather than refused: nothing has been chosen yet, so
// selecting a model the operator never asked for is exactly what the gate exists to prevent.
func (c *Config) installedFaceModel() face.ModelName {
	modelsPath := c.ModelsPath()
	edition := c.Edition()

	for _, candidate := range face.AutoModelPreference {
		if face.LicenseRefused(candidate, edition) != nil {
			continue
		} else if face.FindEmbeddingModel(candidate).Installed(modelsPath) {
			return candidate
		}
	}

	return face.ModelNone
}

// checkFaceModelMismatch blocks embedding work when the library holds vectors that cannot be
// compared with the model in force.
//
// Without it a mismatch is a filter rather than a fault: matching runs, finds nothing, reports
// nothing wrong, and indexing keeps adding vectors in the configured model's space beside the
// ones already there.
func (c *Config) checkFaceModelMismatch(counts []query.MarkerEmbeddingModelCount) {
	name := c.FaceModel()

	if name == face.ModelNone {
		face.UnblockEmbeddings()
		return
	}

	stale := 0
	models := make([]string, 0, len(counts))

	for _, count := range counts {
		if count.Markers < 1 || face.ModelsComparable(count.EmbedModel, name) {
			continue
		}

		stale += count.Markers
		models = append(models, clean.Log(recordedFaceModel(count.EmbedModel)))
	}

	if stale == 0 {
		face.UnblockEmbeddings()
		return
	}

	reason := fmt.Sprintf("%d marker(s) use %s, but this instance is configured for %s",
		stale, txt.JoinAnd(models), clean.Log(name))

	face.BlockEmbeddings(reason)

	log.Warnf(`faces: %s, so face embeddings are not processed (run "photoprism faces migrate --to %s" to migrate them)`,
		reason, clean.Log(name))
}

// warnFaceModel reports a face model problem once, because the getters are called from
// Propagate and from the config report rather than a single time per start.
func (c *Config) warnFaceModel(key, format string, args ...any) {
	if _, warned := c.faceWarned.LoadOrStore(key, true); !warned {
		log.Warnf(format, args...)
	}
}

// libraryFaceModel returns the embedding model the library's face vectors were generated
// with, or an empty name when it holds none and when the database is not connected yet.
func (c *Config) libraryFaceModel() face.ModelName {
	return dominantFaceModel(c.libraryFaceModels())
}

// libraryFaceModels returns how many face markers the library holds per embedding model, or
// nothing when it holds none and when the database is not connected yet.
//
// This runs once per process, including for every CLI invocation, so it asks the indexed
// provenance column first and only falls back to counting vectors when nothing answers.
func (c *Config) libraryFaceModels() []query.MarkerEmbeddingModelCount {
	if c == nil || c.db == nil {
		return nil
	}

	counts, err := query.RecordedMarkerEmbeddingModels()

	if err == nil && len(counts) > 0 {
		return counts
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
		return []query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: int(markers)}}
	}

	return nil
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
	}

	return recordedFaceModel(name)
}

// recordedFaceModel returns the model that produced a vector from the name recorded with it,
// so a blank name reads as the model it can only have come from rather than as unknown.
func recordedFaceModel(recorded string) face.ModelName {
	if name := face.NormalizeModelName(recorded); name != "" {
		return name
	}

	return face.ModelFaceNet
}

// FaceEmbeddingModel returns the resolved face embedding model, or nil when
// embeddings are disabled or no model is installed.
func (c *Config) FaceEmbeddingModel() *face.EmbeddingModel {
	return face.FindEmbeddingModel(c.FaceModel())
}

// ConfigureFaceEmbedder loads the embedding model with the specified name.
//
// It takes the name rather than reading the configuration, so a migration can load the model
// it writes before the instance is configured for it. Loading the active model is a no-op.
func (c *Config) ConfigureFaceEmbedder(name face.ModelName) error {
	if c == nil {
		return nil
	}

	model := face.FindEmbeddingModel(name)

	return face.ConfigureEmbedder(face.EmbedderSettings{
		Name:      name,
		Model:     model,
		ModelPath: model.FilePath(c.ModelsPath()),
		Threads:   c.FaceModelThreads(),
	})
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
	radius, _ := c.faceAcceptThresholds()
	return radius
}

// faceAcceptThresholds returns the cluster radius and match distance, falling back to the configured
// model's calibrated pair when the two together reach past face.ConfigDistMax.
//
// Resolved as a pair because a cluster accepts at their sum, so one wide option reaches the limit on
// its own - where a cluster accepts strangers as readily as the person it was built from.
func (c *Config) faceAcceptThresholds() (radius, matchDist float64) {
	pickRadius := func(m *face.EmbeddingModel) float64 { return m.ClusterRadius }
	pickMatchDist := func(m *face.EmbeddingModel) float64 { return m.MatchDist }

	radius = c.faceThreshold("face-cluster-radius", c.options.FaceClusterRadius, face.ClusterRadiusDefault, pickRadius)
	matchDist = c.faceThreshold("face-match-dist", c.options.FaceMatchDist, face.MatchDistDefault, pickMatchDist)

	if radius+matchDist <= face.ConfigDistMax {
		return radius, matchDist
	}

	model := c.FaceEmbeddingModel()
	calibratedRadius := faceModelThreshold(model, pickRadius, face.ClusterRadiusDefault)
	calibratedMatchDist := faceModelThreshold(model, pickMatchDist, face.MatchDistDefault)

	// A model whose own calibration reaches past the limit has nothing to fall back to, and
	// warning about a value nobody set would be noise. TestEmbeddingModelThresholds is where
	// that is caught, when the model is registered rather than when it is used.
	if radius == calibratedRadius && matchDist == calibratedMatchDist {
		return radius, matchDist
	}

	if _, warned := c.faceWarned.LoadOrStore("face-accept-dist", true); !warned {
		log.Warnf("config: face-cluster-radius %g and face-match-dist %g accept faces up to %g, more than the maximum of %g, using %g and %g instead",
			radius, matchDist, radius+matchDist, face.ConfigDistMax, calibratedRadius, calibratedMatchDist)
	}

	return calibratedRadius, calibratedMatchDist
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
	_, matchDist := c.faceAcceptThresholds()
	return matchDist
}

// faceThreshold returns the operator-configured clustering threshold, or the value
// calibrated for the configured embedding model when the option was left untouched.
func (c *Config) faceThreshold(flagName string, value, flagDefault float64, pick func(*face.EmbeddingModel) float64) float64 {
	minDist := c.FaceCollisionDist()
	configured := c.faceThresholdIsSet(flagName, value, flagDefault)

	if configured && value >= minDist && value <= face.ConfigDistMax {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(), pick, flagDefault)

	c.warnFaceThreshold(configured, flagName, value, minDist, face.ConfigDistMax, resolved)

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

	// These options are yaml:"-", so "options.yml" never sets them and the flag default is
	// the only value an operator did not choose. Anything else was configured deliberately.
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
