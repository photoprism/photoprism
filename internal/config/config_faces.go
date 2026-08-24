package config

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

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

// FaceDetectorSetting returns the detector as configured, without resolving it. It reports
// `face.DetectorDetect` when the detector is to be derived from the embedding model, which is
// also what an unsupported value is treated as.
//
// The deprecated `FACE_ENGINE` is consulted only when nothing configured this option, and only
// `none` carries over: its other values name a runtime rather than a model, and all of them
// mean "detection is enabled", which this option's own default already expresses.
func (c *Config) FaceDetectorSetting() face.DetectorName {
	if c == nil {
		return face.DetectorNone
	}

	configured := face.DetectorDetect

	if !face.KnownDetectorName(c.options.FaceDetector) {
		c.warnFaceConfig("face-detector", "config: unsupported face detector %s, expected %s",
			clean.Log(c.options.FaceDetector), face.DetectorUsageString())
	} else {
		configured = face.ParseDetectorName(c.options.FaceDetector)
	}

	if face.ParseEngine(c.options.FaceEngine) != face.EngineNone {
		return configured
	}

	if configured == face.DetectorDetect {
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
	case face.DetectorDetect:
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
// Gated weights are reached only through the pairing, never through the scan: the pairing is
// downstream of a model the operator selected explicitly and that the same gate already let
// through, while the scan is what runs when nothing has been chosen.
func (c *Config) derivedFaceDetector() face.DetectorName {
	modelsPath := c.ModelsPath()

	if model := c.FaceEmbeddingModel(); model != nil && model.Detector != "" {
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

// initFaceDetector retires a persisted face engine value in favor of the detector option that
// supersedes it, and is called once by Init.
//
// Only `none` is carried over, because it is the one value that means the same in both. The
// rest name a runtime every detector shares, so they say "detection is enabled", which the
// detector option's own default already expresses.
func (c *Config) initFaceDetector() {
	if c == nil {
		return
	}

	// Only what the file holds is migrated. An environment variable is read afresh on every
	// start, so persisting one would outlive the moment it was set and keep disabling detection
	// after the operator removed it.
	_, values, err := c.loadOptionsYAML()

	if err != nil {
		log.Debugf("config: %s (replace the deprecated face engine option)", err)
		return
	}

	engine, found := values["FaceEngine"]

	if !found {
		return
	}

	patch := Values{}

	// A file that already names a detector keeps it, because the deprecated value was consulted
	// only while nothing else was configured.
	if _, named := values["FaceDetector"]; !named && face.ParseEngine(fmt.Sprintf("%v", engine)) == face.EngineNone {
		c.options.FaceDetector = face.DetectorNone
		patch["FaceDetector"] = face.DetectorNone
	}

	c.options.FaceEngine = ""

	if _, saveErr := c.SaveOptionsUpdate(patch, []string{"FaceEngine"}); saveErr != nil {
		log.Warnf("config: failed replacing the deprecated face engine option (%s)", saveErr)
		return
	}

	log.Infof("config: replaced the deprecated face engine option in %s", clean.Log(c.OptionsYaml()))
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

// FaceDetectorThreads returns the thread count for ONNX face detection.
//
// The automatic value divides the cores by the number of indexing workers, because face
// detection takes no lock: that many detections run at once, each with its own thread
// pool, so a per-session count derived from the cores alone oversubscribes the machine
// by exactly that factor.
func (c *Config) FaceDetectorThreads() int {
	if c == nil {
		return 1
	} else if threads := c.faceThreadsSetting(c.options.FaceDetectorThreads); threads > 0 {
		return threads
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
	} else if threads := c.faceThreadsSetting(c.options.FaceModelThreads); threads > 0 {
		return threads
	}

	return max(runtime.NumCPU()/2, 1)
}

// faceThreadsSetting returns the configured thread count, falling back to the deprecated
// option that set detection and embedding together. The two derive different defaults, so
// the shared value is only consulted where nothing more specific was configured.
func (c *Config) faceThreadsSetting(threads int) int {
	if threads > 0 {
		return threads
	}

	return c.options.FaceEngineThreads
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

// FaceEngineModelPath returns the absolute path to the detector weights to load.
//
// When nothing is installed it names the artifact that would have been loaded, so the caller
// reports a missing detector rather than an empty path.
func (c *Config) FaceEngineModelPath() string {
	if c == nil {
		return ""
	}

	models := c.ModelsPath()

	if path := face.FindDetector(c.FaceDetector()).InstalledPath(models); path != "" {
		return path
	}

	return face.DefaultDetector().Path(models)
}

// FaceModelSetting returns the face embedding model as configured, without resolving it.
// It reports `face.ModelDetect` when the model is still to be detected, which is also what
// an unsupported value is treated as.
func (c *Config) FaceModelSetting() face.ModelName {
	if c == nil {
		return face.ModelNone
	}

	if !face.KnownModelName(c.options.FaceModel) {
		c.warnFaceConfig("face-model", "config: unsupported face model %s, expected %s",
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

// ResolveFaceModel pins the model in memory and computes the block verdict, without writing
// anything to "options.yml".
//
// It is what a report calls to name the model in force and whether it can read the library:
// what is configured stays as it was, so `photoprism faces config` still shows "detect".
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
	return c.FaceModelSetting() == face.ModelDetect
}

// SetFaceModel records the model the library's vectors are now in and persists it, reporting
// whether the file could be written.
//
// A migration is the one operation that changes the answer, and the setting has to follow it,
// or a later start hides the whole library from matching - so a caller that changed the data
// has to treat a failed write as a failed run rather than as a warning.
func (c *Config) SetFaceModel(name face.ModelName) error {
	if c == nil || name == "" {
		return nil
	}

	// Set in memory as well as in the file, because the patch helper only reloads the options
	// it changed and a file that already named the model would leave this process behind.
	c.faceModel = name
	c.options.FaceModel = name

	if _, err := c.SaveOptionsPatch(Values{"FaceModel": name}); err != nil {
		return fmt.Errorf("failed saving face model %s (%s)", clean.Log(name), err)
	}

	return nil
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
// nothing wrong, and indexing keeps adding vectors in the configured model's space beside the
// ones already there.
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
			`(run "photoprism faces config" to see why)`, reason)

		return
	}

	reason := fmt.Sprintf("%d marker(s) use %s, but this instance is configured for %s",
		stale, txt.JoinAnd(models), clean.Log(name))

	face.BlockEmbeddings(reason)

	log.Warnf(`faces: %s, so face embeddings are not processed (run "photoprism faces migrate --to %s" to migrate them)`,
		reason, clean.Log(name))
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

// libraryFaceModel returns the embedding model the library's face vectors were generated
// with, or an empty name when it holds none and when the database is not connected yet.
func (c *Config) libraryFaceModel() face.ModelName {
	counts, _ := c.libraryFaceModels()
	return dominantFaceModel(counts)
}

// libraryFaceModels returns how many face markers the library holds per embedding model, and
// whether the library could be asked at all.
//
// Answering "no vectors" and "no answer" with the same value would let a database that is not
// readable yet be pinned as an empty library, so the two are reported apart. This runs once per
// process, including for every CLI invocation, so it asks the indexed provenance column and
// completes it with the rows that predate the column rather than counting every vector.
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
	if c.options.FaceSize < face.MinSizeThreshold || c.options.FaceSize > 10000 {
		return face.SizeThreshold
	}

	return c.options.FaceSize
}

// FaceSizeRetry returns the face size threshold for the second detection pass, which runs only
// when the first found no face at all.
//
// A negative value disables the second pass, and zero selects the default, so that a configuration
// which never set the option behaves like one that asked for the default rather than silently
// turning the fallback off. The result never exceeds the ordinary threshold, because a retry asking
// for larger faces could only find fewer than the pass that already found none.
func (c *Config) FaceSizeRetry() int {
	size := c.options.FaceSizeRetry

	switch {
	case size < 0:
		return 0
	case size == 0 || size > face.SizeThreshold*10:
		size = face.RetrySizeThreshold
	}

	return min(size, c.FaceSize())
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
