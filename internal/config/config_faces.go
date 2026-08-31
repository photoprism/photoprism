package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Global options, in the order "photoprism show config" and "photoprism faces status" report them.

// FaceEngine returns the runtime that face detection runs on, or `face.EngineNone` when it is
// off, the config is nil, or the vision subsystem is not initialized.
//
// Every registered detector runs on ONNX, so the runtime follows the detector in force rather
// than being configured on its own.
func (c *Config) FaceEngine() string {
	if c == nil || vision.Config == nil {
		return face.EngineNone
	}

	if c.FaceDetector() == face.DetectorNone {
		return face.EngineNone
	}

	return face.EngineONNX
}

// FaceRun returns when face detection and recognition are scheduled to run. An unsupported value
// is reported and applies as `vision.RunAuto`.
func (c *Config) FaceRun() vision.RunType {
	if c == nil {
		return vision.RunNever
	}

	if configured := c.options.FaceRun; configured != "" && !vision.KnownRunType(configured) {
		c.warnFaceConfig("face-run", "config: unsupported face run type %s, expected %s",
			clean.Log(configured), vision.RunTypeUsageString())
	}

	return vision.ParseRunType(c.options.FaceRun)
}

// FaceEngineRunType returns the effective run type for face detection and recognition, which is
// `FACE_RUN`. Detection and embedding always run together, so one schedule covers both, and it
// falls back to RunNever when faces are disabled or no detector may be used.
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

	return c.FaceRun()
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

	when = vision.ParseRunType(when)
	run := c.FaceEngineRunType()

	// Faces stay out of the scheduled sweep unless it was asked for by name: re-detecting pictures
	// an earlier pass examined finds nothing while the detector is unchanged, and costs a full
	// decode per file. A detector change is what makes another pass worthwhile.
	if when == vision.RunOnSchedule && run != vision.RunOnSchedule && run != vision.RunAlways {
		return false
	}

	// Every other schedule is decided by the shared table, so face detection and a vision model
	// cannot disagree about what a run type means.
	if should, decided := vision.ShouldRunAt(run, when); decided {
		return should
	}

	// "auto" is where faces differ deliberately: they detect inline only on a host fast enough,
	// and skip the scheduled pass so a background job does not start detecting unannounced.
	switch when {
	case vision.RunAuto, vision.RunAlways, vision.RunManual, vision.RunOnDemand:
		return true
	case vision.RunOnIndex:
		return c.faceEngineRunsOnIndex()
	case vision.RunNewlyIndexed:
		return !c.faceEngineRunsOnIndex()
	}

	return false
}

// faceEngineRunsOnIndex reports whether this host is fast enough to detect faces while indexing
// rather than deferring them to the pass over newly indexed files.
//
// It reads the count that is not divided among the indexing workers, whose number follows the
// database driver - dividing would tie the schedule to the storage backend rather than the machine.
func (c *Config) faceEngineRunsOnIndex() bool {
	return c.FaceModelThreads() > 2
}

// Face detection options.

// FaceDetectorSetting returns the detector as configured, without resolving it. It reports
// `face.DetectorAuto` when it is to be derived, which an unsupported value asks for too.
//
// The deprecated `FACE_ENGINE` is consulted only when this option is unset, and only `none` carries
// over: its other values name a runtime every detector shares, so they add nothing to the default.
func (c *Config) FaceDetectorSetting() face.DetectorName {
	if c == nil {
		return face.DetectorNone
	}

	configured := face.DetectorAuto

	if !face.KnownDetectorName(c.options.FaceDetector) {
		c.warnFaceConfig("face-detector", "config: unsupported face detector %s, expected %s",
			clean.Log(c.options.FaceDetector), face.DetectorUsageString())
	} else {
		configured = face.ParseDetectorName(c.options.FaceDetector)
	}

	if face.ParseEngine(c.options.FaceEngine) != face.EngineNone {
		return configured
	}

	// Only an unset detector lets the deprecated value decide. "auto" is a value an operator
	// chose, and it is the first thing they reach for to turn detection back on, so it has to
	// win as plainly as a detector name does.
	if c.options.FaceDetector == "" {
		return face.DetectorNone
	}

	c.infoFaceConfig("face-engine-ignored", "config: face-engine %s is ignored, because face-detector %s is configured",
		clean.Log(c.options.FaceEngine), clean.Log(configured))

	return configured
}

// FaceDetector returns the name of the detector in force, or `face.DetectorNone` when detection
// is disabled or no detector may be used.
func (c *Config) FaceDetector() face.DetectorName {
	if c == nil {
		return face.DetectorNone
	}

	switch name := c.FaceDetectorSetting(); name {
	case face.DetectorNone:
		return face.DetectorNone
	case face.DetectorAuto:
		return c.derivedFaceDetector()
	default:
		return c.usableFaceDetector(name)
	}
}

// usableFaceDetector returns the specified detector when its weights are installed and may be
// used here, and `face.DetectorNone` with one warning otherwise.
//
// Falling forward to another detector would place different landmarks and therefore a different
// crop, so a detector that was asked for and cannot run disables detection instead.
func (c *Config) usableFaceDetector(name face.DetectorName) face.DetectorName {
	if err := face.DetectorLicenseRefused(name, c.Edition()); err != nil {
		c.warnFaceConfig("face-detector-license", "config: %s, so face detection is disabled", err)

		return face.DetectorNone
	}

	if !face.FindDetector(name).Installed(c.ModelsPath()) {
		c.warnFaceConfig("face-detector-installed", "config: face detector %s is not installed, disabling face detection", clean.Log(name))

		return face.DetectorNone
	}

	return name
}

// derivedFaceDetector returns the detector the configured embedding model pairs with, or the
// first installed detector whose weights may be redistributed.
//
// Gated weights are reached through the pairing only, never through the scan: the pairing follows
// a model the operator selected explicitly, while the scan runs when nothing has been chosen.
func (c *Config) derivedFaceDetector() face.DetectorName {
	modelsPath := c.ModelsPath()

	// The model in force rather than the effective one: this pairing is reached only because an
	// operator selected the model explicitly, and an auto-detected one has not been.
	if model := face.FindEmbeddingModel(c.FaceModel()); model != nil && model.Detector != "" {
		if face.DetectorLicenseRefused(model.Detector, c.Edition()) == nil &&
			face.FindDetector(model.Detector).Installed(modelsPath) {
			return model.Detector
		}
	}

	for _, detector := range face.Detectors {
		if !detector.LicenseGated() && detector.Installed(modelsPath) {
			return detector.Name
		}
	}

	return face.DetectorNone
}

// FaceEngineModelPath returns the absolute path to the detector weights to load.
//
// A detector that was named and refused still names its own artifact, so a report says what was
// asked for rather than pointing at the default, and the caller sees it as missing.
func (c *Config) FaceEngineModelPath() string {
	if c == nil {
		return ""
	}

	setting := c.FaceDetectorSetting()

	// Detection that was switched off names no artifact. Falling through to the default would
	// report a path for weights this instance will not load, which reads as though a detector
	// were running - and that row is what an operator checks to find out whether one is.
	if setting == face.DetectorNone {
		return ""
	}

	models := c.ModelsPath()
	detector := face.FindDetector(c.FaceDetector())

	if detector == nil {
		detector = face.FindDetector(setting)
	}

	if detector == nil {
		detector = face.DefaultDetector()
	}

	if path := detector.InstalledPath(models); path != "" {
		return path
	}

	return detector.Path(models)
}

// FaceDetectorThreads returns the thread count for ONNX face detection.
//
// The automatic value divides the cores by the number of indexing workers, because detection
// takes no lock: that many run at once, each with its own pool, so a count derived from the
// cores alone oversubscribes the machine by exactly that factor.
func (c *Config) FaceDetectorThreads() int {
	if c == nil {
		return 1
	} else if threads := c.faceThreadsSetting(c.options.FaceDetectorThreads); threads > 0 {
		return threads
	}

	return max(runtime.NumCPU()/max(c.IndexWorkers(), 1), 1)
}

// ConfigureFaceDetector loads the detection engine for the detector in force, applying the
// specified minimum score on the 0-100 scale instead of the configured one. Zero asks for the
// configured value, and a negative score switches the cutoff off. The cutoff is baked into the
// inference session, so lowering it means loading the detector again rather than passing it.
func (c *Config) ConfigureFaceDetector(minScore float64) error {
	if c == nil {
		return nil
	}

	if minScore == 0 {
		minScore = c.FaceScore()
	}

	return face.ConfigureEngine(face.EngineSettings{
		Name: c.FaceEngine(),
		ONNX: face.ONNXOptions{
			ModelPath:      c.FaceEngineModelPath(),
			Threads:        c.FaceDetectorThreads(),
			ScoreThreshold: detectorScoreThreshold(minScore),
		},
	})
}

// detectorScoreThreshold converts a minimum score from the 0-100 scale operators read scores in
// to the 0-1 scale a detector reports them on. Without it no reachable value could lower a
// calibrated cutoff. The two sentinels carry through unscaled, because neither is a score:
// zero leaves the cutoff to the detector, and negative switches it off.
func detectorScoreThreshold(minScore float64) float32 {
	if minScore <= 0 {
		return float32(minScore)
	}

	return float32(minScore / 100)
}

// FaceSize returns the face size threshold in pixels. It falls back to the shipped default
// rather than to `face.SizeThreshold`, which Propagate assigns from this.
func (c *Config) FaceSize() int {
	if c.options.FaceSize < face.MinSizeThreshold || c.options.FaceSize > 10000 {
		return face.SizeThresholdDefault
	}

	return c.options.FaceSize
}

// FaceSizeRetry returns the face size threshold for the second detection pass, which runs only
// when the first found no face at all. A negative value disables it, and zero selects the default.
//
// Zero has to mean the default, or a configuration that never set the option would turn the
// fallback off. The result never exceeds the ordinary threshold, which could only find fewer.
func (c *Config) FaceSizeRetry() int {
	size := c.options.FaceSizeRetry

	switch {
	case size < 0:
		return 0
	case size == 0 || size > face.SizeThresholdDefault*10:
		size = face.RetrySizeThreshold
	}

	return min(size, c.FaceSize())
}

// FaceScore returns the configured minimum detection score on the 0-100 scale, zero when each
// detector's own calibrated cutoff is to decide, and -1 when no cutoff is to be applied at all.
//
// Unset means zero rather than a number, because a shared default either sits below every
// detector's cutoff and gates nothing, or above one and silently overrules a calibration. Negative
// follows the convention the other numeric options use - see DefaultStorageFree.
func (c *Config) FaceScore() float64 {
	switch {
	case c.options.FaceScore < 0:
		return face.NoScoreThreshold
	case c.options.FaceScore < 1 || c.options.FaceScore > 100:
		return face.ScoreThresholdDefault
	}

	return c.options.FaceScore
}

// FaceScoreEffective returns the detection cutoff actually in force on the 0-100 scale, which is
// the detector's own whenever FACE_SCORE is unset, and -1 when it is switched off. A report has
// to state this rather than the raw option, which is zero in the ordinary case and reads as
// "nothing is filtered" - the one thing zero does not mean.
func (c *Config) FaceScoreEffective() float64 {
	if score := c.FaceScore(); score != face.ScoreThresholdDefault {
		return score
	}

	return face.DetectorScore(c.FaceDetector())
}

// FaceMigrateSize returns the minimum face size a migration re-detects at, in pixels of the
// detection thumbnail.
//
// It does not inherit FACE_SIZE. A marker's size is in the pixels of the thumbnail it was detected
// in, and an earlier detector fell back to a larger one, so a legacy marker can sit well under the
// ordinary floor - which no score recovers, because the detector never emits the candidate.
func (c *Config) FaceMigrateSize() int {
	if c == nil {
		return face.MinSizeThreshold
	}

	if size := c.options.FaceMigrateSize; size >= face.MinSizeThreshold && size <= 10000 {
		return size
	}

	return face.MinSizeThreshold
}

// FaceMigrateScore returns the detection cutoff a migration re-detects at, on the 0-100 scale.
//
// A migration makes the opposite trade to an index: a false positive there costs a thumbnail to
// reject, a miss costs a curated marker its vector. It therefore has its own floor, registered per
// detector like the cutoff itself. A configured FACE_SCORE still stands when this is unset.
func (c *Config) FaceMigrateScore() float64 {
	if c == nil {
		return face.DefaultDetectorMigrateScore()
	}

	switch {
	case c.options.FaceMigrateScore < 0:
		return face.NoScoreThreshold
	case c.options.FaceMigrateScore >= 1 && c.options.FaceMigrateScore <= 100:
		return c.options.FaceMigrateScore
	}

	if score := c.FaceScore(); score != face.ScoreThresholdDefault {
		return score
	}

	return face.DetectorMigrateScore(c.FaceDetector())
}

// FaceOverlap returns the face area overlap threshold in percent.
func (c *Config) FaceOverlap() int {
	if c.options.FaceOverlap < 1 || c.options.FaceOverlap > 100 {
		return face.OverlapThresholdDefault
	}

	return c.options.FaceOverlap
}

// Face recognition options.

// FaceModelSetting returns the face embedding model as configured, without resolving it.
// It reports `face.ModelAuto` when the model is still to be detected, which is also what
// an unsupported value is treated as.
func (c *Config) FaceModelSetting() face.ModelName {
	if c == nil {
		return face.ModelNone
	}

	if !face.KnownModelName(c.options.FaceModel) {
		c.warnFaceConfig("face-model", "config: unsupported face model %s, expected %s",
			clean.Log(c.options.FaceModel), face.ModelUsageString())

		return face.ModelAuto
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
	case face.ModelAuto, face.ModelNone:
		return face.ModelNone
	default:
		return c.usableFaceModel(name)
	}
}

// usableFaceModel returns the specified model when its weights are installed and may be used
// here, and `face.ModelNone` with one warning otherwise.
func (c *Config) usableFaceModel(name face.ModelName) face.ModelName {
	if err := face.LicenseRefused(name, c.Edition()); err != nil {
		c.warnFaceConfig("face-model-license", "config: %s, so face embeddings are disabled", err)

		return face.ModelNone
	}

	if !face.FindEmbeddingModel(name).Installed(c.ModelsPath()) {
		// Falling forward to another model would start a second vector space the library
		// cannot compare with, and an image upgrade removes opt-in models from assets, so
		// a model that is missing disables embeddings instead.
		c.warnFaceConfig("face-model-installed", "config: face model %s is not installed, disabling face embeddings", clean.Log(name))

		return face.ModelNone
	}

	return name
}

// EffectiveFaceModel returns the model whose calibration applies: the one in force, or the one
// detection would settle on while the library has not been asked yet.
//
// A report runs before the database is connected, where FaceModel reports none for a model that
// is merely undetected - which left every calibrated distance in it blank.
func (c *Config) EffectiveFaceModel() face.ModelName {
	if c == nil {
		return face.ModelNone
	}

	if name := c.FaceModel(); name != face.ModelNone {
		return name
	} else if c.FaceModelSetting() == face.ModelAuto {
		return c.installedFaceModel()
	}

	return face.ModelNone
}

// ResolveFaceModel pins the model in memory and computes the block verdict, without writing
// anything to "options.yml".
//
// It is what a report calls to name the model in force and whether it can read the library. The
// configured value stays as it was, so a report cannot turn a request to detect into a decision.
func (c *Config) ResolveFaceModel() face.ModelName {
	if c == nil {
		return face.ModelNone
	} else if c.db == nil {
		return c.FaceModel()
	}

	counts, ok := c.libraryFaceModels()

	if c.faceModel == "" && c.faceModelDetects() {
		c.faceModel = c.detectFaceModel(counts)
	}

	if ok {
		c.checkFaceModelMismatch(counts)
	}

	return c.FaceModel()
}

// faceModelDetects reports whether the model has to be worked out from the library, which an
// unsupported value asks for as much as an empty one does.
func (c *Config) faceModelDetects() bool {
	return c.FaceModelSetting() == face.ModelAuto
}

// initFaceModel settles which embedding model this instance uses, and is called once by Init
// on a start that has a database.
func (c *Config) initFaceModel() {
	if c == nil || c.db == nil {
		return
	}

	c.reportIgnoredFaceModel()

	counts, ok := c.libraryFaceModels()

	if c.faceModelDetects() {
		c.faceModel = c.detectFaceModel(counts)

		// Persist only an answer the library actually gave. A value derived from an unreadable
		// database or from an empty one is a default rather than a finding, and writing it down
		// would outlive the moment it was true - a restored library would then be refused by a
		// model nothing in it was produced with.
		switch {
		case !ok || len(counts) == 0 || c.faceModel == face.ModelNone:
			log.Debugf("config: face model %s is not recorded yet", clean.Log(c.faceModel))
		case !face.KnownModelName(c.options.FaceModel):
			// An unsupported value applies to the run as if nothing were set, but writing a
			// detected name would turn a typo into a decision that outlives it.
			log.Warnf("config: face model %s is detected but not recorded, because %s is not a supported value",
				clean.Log(c.faceModel), clean.Log(c.options.FaceModel))
		default:
			log.Infof("config: detected face model %s", clean.Log(c.faceModel))

			if err := c.SetFaceModel(c.faceModel); err != nil {
				// The value applies to this process either way, and the next start detects again.
				log.Warnf("config: %s", err)
			}
		}
	}

	if ok {
		c.checkFaceModelMismatch(counts)
	}
}

// SetFaceModel records the model the library's vectors are now in and persists it, reporting
// whether the file could be written.
//
// A caller that changed the data has to treat a failed write as a failed run: the setting has to
// follow the migration, or a later start hides the whole library from matching.
func (c *Config) SetFaceModel(name face.ModelName) error {
	if c == nil || name == "" {
		return nil
	}

	// Set in memory as well as in the file: the patch helper applies only the keys it wrote, and
	// a file that already named the model would leave this process behind.
	c.faceModel = name
	c.options.FaceModel = name

	// Before the write, so a setting that could not be persisted still applies to this process:
	// what follows the model is a distance space, and clustering at the previous one's calibration
	// rewrites the library at thresholds its vectors were never measured in.
	c.PropagateFaceModel()

	if _, err := c.SaveOptionsPatch(Values{"FaceModel": name}); err != nil {
		return fmt.Errorf("failed saving face model %s (%s)", clean.Log(name), err)
	}

	return nil
}

// PropagateFaceModel assigns the values in the face package that the embedding model in force
// calibrates, and loads the embedder for it.
//
// Called by Propagate as well, but the model is also settled outside a start: a migration commits
// its vectors and records the new model while the process runs, and every threshold below is in
// the distance space of a model, not a number that carries from one to the next.
func (c *Config) PropagateFaceModel() {
	if c == nil {
		return
	}

	face.ClusterSizeThreshold = c.FaceClusterSize()
	face.CollisionDist = c.FaceCollisionDist()
	face.Epsilon = c.FaceEpsilonDist()
	face.ClusterRadius = c.FaceClusterRadius()
	face.ClusterDist = c.FaceClusterDist()
	face.MatchDist = c.FaceMatchDist()

	if err := c.ConfigureFaceEmbedder(c.FaceModel()); err != nil {
		log.Warnf("faces: %s (configure embedding model)", err)
	}
}

// ClearFaceModel removes a pinned embedding model from "options.yml" and from this process, so
// the next start detects one again.
//
// Keeping vectors comparable is the only reason the model is pinned, and a reset leaves none.
// The pin would otherwise outlive its reason, undoable only by hand-editing the file.
func (c *Config) ClearFaceModel() error {
	if c == nil {
		return nil
	}

	// The file is written first, so a failure leaves this process agreeing with it rather than
	// resolving a model that a restart would pin again.
	if _, err := c.DeleteOptionsPatch("FaceModel"); err != nil {
		return fmt.Errorf("failed clearing the face model (%s)", err)
	}

	c.faceModel = ""
	c.options.FaceModel = ""

	return nil
}

// reportIgnoredFaceModel reports a model that an environment variable or the command line set
// while "options.yml" names a different one.
//
// The file wins for every option, and here it keeps the instance intact: changing the model is a
// data migration, so a variable could only leave the library in two vector spaces.
func (c *Config) reportIgnoredFaceModel() {
	if c.faceModelFlag == "" {
		return
	}

	configured := face.ParseModelName(c.faceModelFlag)
	inForce := c.FaceModelSetting()

	if configured == face.ModelAuto || configured == inForce {
		return
	}

	// Disabling embeddings is not something a migration can carry out, so that value is the one
	// case where the way to apply it is the file rather than the command.
	if configured == face.ModelNone {
		log.Infof("config: face model %s is configured but ignored, this library uses %s "+
			"(set FaceModel in %s to change it)", clean.Log(configured), clean.Log(inForce), clean.Log(c.OptionsYaml()))

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
// nothing wrong, and indexing keeps adding vectors in a second space beside the first.
func (c *Config) checkFaceModelMismatch(counts []query.MarkerEmbeddingModelCount) {
	name := c.FaceModel()

	// Embeddings that were turned off are not a mismatch: nothing is generated, and the vectors
	// a library already holds stay comparable with each other.
	if name == face.ModelNone && c.FaceModelSetting() == face.ModelNone {
		face.UnblockEmbeddings()
		return
	}

	// A model that could not be used reads none of the stored vectors, so every one of them is
	// stale: clustering them under another model's distances would rewrite the library at
	// thresholds it was never calibrated for.
	stale, models := staleFaceModels(counts, name)

	if stale == 0 {
		face.UnblockEmbeddings()
		return
	}

	if name == face.ModelNone {
		reason := fmt.Sprintf("%d marker(s) use %s, which this instance cannot load",
			stale, txt.JoinAnd(models))

		face.BlockEmbeddings(reason)

		// Why it cannot be loaded was reported by the getter, under a different prefix and
		// possibly many lines earlier, so this names where to read it back.
		log.Warnf(`faces: %s, so face embeddings are not processed `+
			`(run "photoprism faces status" to see why)`, reason)

		return
	}

	// A migration commits in its own process and records its target, which nothing reloads here.
	// Naming the model this one still holds would ask for the library to be migrated back into
	// the space it was just migrated out of, at the moment the advice is most likely to be taken.
	if c.CheckFaceModelSuperseded() {
		return
	}

	reason := fmt.Sprintf("%d marker(s) use %s, but this instance is configured for %s",
		stale, txt.JoinAnd(models), clean.Log(name))

	face.BlockEmbeddings(reason)

	log.Warnf(`faces: %s, so face embeddings are not processed (run "photoprism faces migrate --to %s" to migrate them)`,
		reason, clean.Log(name))
}

// CheckFaceModelSuperseded pauses face embedding work and asks for a restart when "options.yml"
// records an embedding model this process has not loaded, reporting whether it did.
//
// What a completed migration leaves behind is a setting, and a running instance keeps clustering
// and embedding in the model it started with - which compares vectors of two different lengths.
func (c *Config) CheckFaceModelSuperseded() bool {
	superseded := c.SupersededFaceModel()

	if superseded == "" {
		return false
	}

	face.BlockEmbeddings(fmt.Sprintf("face model %s is recorded in %s but not loaded here",
		clean.Log(superseded), clean.Log(filepath.Base(c.OptionsYaml()))))

	// Keyed by the model, so a second migration in the same process is reported again, while a
	// worker that wakes every few minutes does not repeat the first.
	c.warnFaceConfig("face-model-superseded-"+superseded,
		"faces: face model %s is recorded in %s but not loaded here, so face embeddings are paused until this instance is restarted",
		clean.Log(superseded), clean.Log(filepath.Base(c.OptionsYaml())))

	return true
}

// SupersededFaceModel returns the embedding model recorded in "options.yml" when this process
// holds a different one, and an empty name when the two agree or the file names none.
//
// The file is read rather than remembered: a migration persists its target from another process,
// so this is what tells a setting that moved on from an instance pointed at the wrong library.
// The two look alike from the library and have opposite remedies.
func (c *Config) SupersededFaceModel() face.ModelName {
	if c == nil {
		return ""
	}

	fileName := c.OptionsYaml()

	if fileName == "" || !fs.FileExists(fileName) {
		return ""
	}

	b, err := os.ReadFile(fileName) //nolint:gosec // path derived from the config directory

	if err != nil {
		return ""
	}

	values := Values{}

	if err = yaml.Unmarshal(b, &values); err != nil {
		return ""
	}

	recorded, _ := values["FaceModel"].(string)
	persisted := face.ParseModelName(recorded)

	// Only a model that names a vector space is a verdict: "auto" asks for detection, an
	// unsupported value applies as if nothing were set, and "none" generates no vector at all.
	if persisted == face.ModelAuto || persisted == face.ModelNone {
		return ""
	}

	current := c.FaceModel()

	// A process with no model in force has none to be superseded, and a restart installs no
	// weights: why it could not be loaded is reported where that is decided.
	if current == face.ModelNone || face.ModelsComparable(persisted, current) {
		return ""
	}

	return persisted
}

// staleFaceModels returns how many of the counted markers hold a vector the specified model
// cannot read, and which models produced them. Nothing is readable when no model is in force.
func staleFaceModels(counts []query.MarkerEmbeddingModelCount, name face.ModelName) (stale int, models []string) {
	models = make([]string, 0, len(counts))

	for _, count := range counts {
		if count.Markers < 1 || name != face.ModelNone && face.ModelsComparable(count.EmbedModel, name) {
			continue
		}

		stale += count.Markers

		if recorded := clean.Log(recordedFaceModel(count.EmbedModel)); !slices.Contains(models, recorded) {
			models = append(models, recorded)
		}
	}

	return stale, models
}

// libraryFaceModel returns the embedding model the library's face vectors were generated
// with, or an empty name when it holds none and when the database is not connected yet.
func (c *Config) libraryFaceModel() face.ModelName {
	counts, _ := c.libraryFaceModels()
	return dominantFaceModel(counts)
}

// libraryFaceModels returns how many face markers the library holds per embedding model, and
// whether the library could be asked at all.
//
// The two are reported apart, or a database that is not readable yet would be pinned as an empty
// library. It runs once per CLI invocation, so it asks the indexed column rather than every blob.
func (c *Config) libraryFaceModels() (counts []query.MarkerEmbeddingModelCount, ok bool) {
	if c == nil || c.db == nil {
		return nil, false
	}

	counts, err := query.RecordedMarkerEmbeddingModels()

	if err != nil {
		// The schema is migrated after the configuration is propagated, so on the first
		// start after an upgrade the provenance column does not exist yet. Its face markers
		// prove what they hold: a library that records no model can only hold FaceNet.
		log.Debugf("config: %s (find face embedding models)", err)

		markers, countErr := query.FaceMarkersWithVectors()

		if countErr != nil {
			log.Debugf("config: %s (count face markers)", countErr)
			return nil, false
		} else if markers > 0 {
			return []query.MarkerEmbeddingModelCount{{EmbedModel: "", Markers: int(markers)}}, true
		}

		return nil, true
	}

	// Vectors written before the column are not in the counts above, and leaving them out
	// would let a handful of migrated markers outvote a whole legacy library and hide it from
	// the mismatch check - which is the silent filtering the check exists to replace.
	legacy, legacyErr := query.LegacyFaceMarkersWithVectors()

	if legacyErr != nil {
		log.Debugf("config: %s (count legacy face markers)", legacyErr)
		return counts, false
	} else if legacy > 0 {
		counts = append(counts, query.MarkerEmbeddingModelCount{EmbedModel: "", Markers: int(legacy)})
	}

	return counts, true
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
	return face.FindEmbeddingModel(c.EffectiveFaceModel())
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

// FaceModelThreads returns the thread count for face embedding inference.
//
// Embeddings are generated one at a time behind the model session lock, so unlike
// detection there is one session in total rather than one per indexing worker, and the
// count is not divided among them.
func (c *Config) FaceModelThreads() int {
	if c == nil {
		return 1
	} else if threads := c.faceThreadsSetting(c.options.FaceModelThreads); threads > 0 {
		return threads
	}

	return max(runtime.NumCPU()/2, 1)
}

// FaceMatchDist returns the offset distance when matching faces with clusters.
func (c *Config) FaceMatchDist() float64 {
	_, matchDist := c.faceAcceptThresholds()
	return matchDist
}

// FaceMatchMargin returns how much closer the nearest cluster has to be than the runner-up before
// a marker is assigned to it, or 0 when a marker always goes to its nearest cluster.
//
// It does not go through faceThreshold, which floors a value at the collision distance: this is the
// difference between two distances, and a useful margin is legitimately smaller than either.
func (c *Config) FaceMatchMargin() float64 {
	value := c.options.FaceMatchMargin
	configured := c.faceThresholdIsSet("face-match-margin", value)

	if !configured || value == 0 {
		return face.MatchMarginDefault
	}

	// Any negative value switches the check off, face.NoMatchMargin being the documented spelling.
	// A decision rather than an out-of-range value, so it applies without a warning.
	if value < 0 {
		return 0
	}

	if value <= face.ConfigDistMax {
		return value
	}

	c.warnFaceThreshold(true, "face-match-margin", value, 0, face.ConfigDistMax, face.MatchMarginDefault)

	return face.MatchMarginDefault
}

// FaceCollisionDist returns the minimum distance used to differentiate embeddings.
//
// It does not go through faceThreshold, which takes this value as its lower bound and
// would recurse.
func (c *Config) FaceCollisionDist() float64 {
	value := c.options.FaceCollisionDist
	configured := c.faceThresholdIsSet("face-collision-dist", value)

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
//
// Bounded by face.EpsilonDistMax rather than by the default: the ceiling states how far the
// ambiguity cutoff may be widened, so lowering the default cannot invalidate a setting in range.
func (c *Config) FaceEpsilonDist() float64 {
	value := c.options.FaceEpsilonDist
	configured := c.faceThresholdIsSet("face-epsilon-dist", value)

	if value > 0 && value <= face.EpsilonDistMax && configured {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(),
		func(m *face.EmbeddingModel) float64 { return m.Epsilon }, face.EpsilonDefault)

	c.warnFaceThreshold(configured && value != 0, "face-epsilon-dist", value, 0, face.EpsilonDistMax, resolved)

	return resolved
}

// FaceClusterSize returns the size threshold for faces forming a cluster in pixels.
//
// Resolved from the configured model when unset, because the bar means "the model consumes this
// without interpolating" and every model states that as its own input geometry.
func (c *Config) FaceClusterSize() int {
	if c.options.FaceClusterSize < 20 || c.options.FaceClusterSize > 10000 {
		return face.ClusterSize(c.EffectiveFaceModel())
	}

	return c.options.FaceClusterSize
}

// FaceClusterScore returns the operator's own quality threshold for faces forming a cluster, zero
// when the bar is left to the detector that scored each marker, and -1 when it is switched off.
//
// Unset has to stay distinguishable from a value, because the bar is per marker: collapsing it to
// the detector in force would judge markers a different one scored by the wrong calibration.
func (c *Config) FaceClusterScore() int {
	switch {
	case c.options.FaceClusterScore < 0:
		return -1
	case c.options.FaceClusterScore < 1 || c.options.FaceClusterScore > 100:
		return 0
	}

	return c.options.FaceClusterScore
}

// FaceClusterScoreEffective returns the clustering bar in force for a marker the detector in force
// produced, which is what a report has to state: the raw option is zero in the ordinary case, and
// zero reads as "every face may cluster".
func (c *Config) FaceClusterScoreEffective() int {
	if score := c.FaceClusterScore(); score != 0 {
		return score
	}

	return c.detectorClusterScore()
}

// detectorClusterScore returns the clustering bar registered for the detector in force. A bar that
// is a meaningful step above one detector's cutoff sits below another's and gates nothing, so it
// follows the detector the same way the cutoff itself does.
func (c *Config) detectorClusterScore() int {
	if d := face.FindDetector(c.FaceDetector()); d != nil && d.ClusterScore > 0 {
		return d.ClusterScore
	}

	return face.ClusterScoreThresholdDefault
}

// FaceSampleThreshold returns how many new markers an automatic clustering pass needs before it
// runs. Derived from FACE_CLUSTER_CORE rather than configured on its own, and read through the
// getter rather than through face.SampleThreshold: the commands that report it never propagate,
// so the global still holds the shipped default there.
func (c *Config) FaceSampleThreshold() int {
	return 2 * c.FaceClusterCore()
}

// FaceClusterCore returns the number of faces forming a cluster core.
//
// Bounded below at 2, since a cluster seeded from one embedding has no centroid and is not offered
// for matching at all: accepting 1 here would leave every automatic cluster invisible to the
// matcher, with no error and no result.
func (c *Config) FaceClusterCore() int {
	if c.options.FaceClusterCore < 2 || c.options.FaceClusterCore > 100 {
		return face.ClusterCoreDefault
	}

	return c.options.FaceClusterCore
}

// FaceRecomputeStats reports whether a matching pass should derive a cluster's radius from the
// markers it holds, rather than from the widest distance one pass happened to accept.
//
// Transitional, and off while the width guard is calibrated so a percentile sweep moves one variable.
// It goes away once the recompute is the only behavior, which needs the backfill that reaches the
// clusters a pass never visits - until then a run would fix the active rows and leave the stale ones.
func (c *Config) FaceRecomputeStats() bool {
	return c != nil && c.options.FaceRecomputeStats
}

// FaceClusterPercentile returns the percentile of a cluster's member distances its stored radius is
// derived from, so a calibration run can compare the shipped value against the maximum.
//
// 100 is the maximum, which lets one loose member decide how far a whole cluster reaches. Out of
// range reads as unset, since a percentile below 1 selects no member at all.
func (c *Config) FaceClusterPercentile() int {
	if c.options.FaceClusterPercentile < 1 || c.options.FaceClusterPercentile > 100 {
		return face.ClusterPercentileDefault
	}

	return c.options.FaceClusterPercentile
}

// FaceClusterSplitRounds returns how often a group wider than its own accept distance may be
// re-clustered before it is given up on. The flag is hidden because it is a probe.
//
// Anything below face.ClusterSplitOff reads as unset, so a mistyped -2 cannot silently remove the
// only width limit an anonymous cluster has.
func (c *Config) FaceClusterSplitRounds() int {
	value := c.options.FaceClusterSplitRounds

	// Zero discards a wide group rather than meaning "unset", so it counts only when it was asked
	// for: an Options built without the flag defaults holds it and would otherwise switch splitting
	// off for every caller that does not go through the CLI.
	if value < face.ClusterSplitOff || value == 0 && !c.flagIsSet("face-cluster-split-rounds") {
		return face.ClusterSplitRoundsDefault
	}

	return value
}

// flagIsSet reports whether a flag was given on the command line or through its environment
// variable, which is what tells an explicit zero from the zero value of a struct.
func (c *Config) flagIsSet(name string) bool {
	return c != nil && c.cliCtx != nil && c.cliCtx.IsSet(name)
}

// FaceClusterSplitShrink returns how much each split round shortens the link distance by. One or
// more is refused rather than clamped, since it would spend the budget repeating the previous pass.
func (c *Config) FaceClusterSplitShrink() float64 {
	if v := c.options.FaceClusterSplitShrink; v > 0 && v < 1 {
		return v
	}

	return face.ClusterSplitShrinkDefault
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

// faceThreshold returns the operator-configured clustering threshold, or the value
// calibrated for the configured embedding model when the option was left untouched.
func (c *Config) faceThreshold(flagName string, value, flagDefault float64, pick func(*face.EmbeddingModel) float64) float64 {
	minDist := c.FaceCollisionDist()
	configured := c.faceThresholdIsSet(flagName, value)

	if configured && value >= minDist && value <= face.ConfigDistMax {
		return value
	}

	resolved := faceModelThreshold(c.FaceEmbeddingModel(), pick, flagDefault)

	c.warnFaceThreshold(configured, flagName, value, minDist, face.ConfigDistMax, resolved)

	return resolved
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

// Shared helpers and the migration lock.

// faceThreadsSetting returns the configured thread count, falling back to the deprecated
// option that set detection and embedding together. The two derive different defaults, so
// the shared value is only consulted where nothing more specific was configured.
func (c *Config) faceThreadsSetting(threads int) int {
	if threads > 0 {
		return threads
	}

	return c.options.FaceEngineThreads
}

// warnFaceConfig reports a face configuration problem once, because the getters are called from
// Propagate and from the config report rather than a single time per start.
func (c *Config) warnFaceConfig(key, format string, args ...any) {
	if _, warned := c.faceWarned.LoadOrStore(key, true); !warned {
		log.Warnf(format, args...)
	}
}

// infoFaceConfig reports a face setting that has no effect once. It is not a fault, so it is
// reported at info level, but an instruction that is ignored must still not be silent.
func (c *Config) infoFaceConfig(key, format string, args ...any) {
	if _, warned := c.faceWarned.LoadOrStore(key, true); !warned {
		log.Infof(format, args...)
	}
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
//
// These options carry no flag default, because the value that applies is calibrated per embedding
// model and "photoprism --help" cannot name one. They are also yaml:"-", so "options.yml" never
// sets them and anything but zero was chosen deliberately.
func (c *Config) faceThresholdIsSet(flagName string, value float64) bool {
	if c.cliCtx != nil && c.cliCtx.IsSet(flagName) {
		return true
	}

	return value != 0
}

// FacesLockFile returns the path of the lock a face embedding migration holds while it runs.
//
// Beside the storage serial rather than in the config or cache path: it describes what this
// instance is doing now, is worthless to a restored backup, and must survive a cleared cache.
func (c *Config) FacesLockFile() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.StoragePath(), "faces.lock")
}

// FacesLocked returns a description of the face migration currently holding the lock, or "" when
// none is. Workers that write markers or clusters consult it, because the worker activities are
// process-local and a migration ordinarily runs from the command line rather than the server.
func (c *Config) FacesLocked() string {
	return mutex.FileLockHeld(c.FacesLockFile())
}
