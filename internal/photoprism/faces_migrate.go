package photoprism

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const facesMigrateBatchSize = 100

// facesMigrateMaxFailureRatio bounds the share of attempted markers that may lose a vector the
// library could have used, before the destructive finalize is refused.
//
// Only markers clearing both clustering bars count: one below them seeds no cluster and joins
// none. Counting the rest made a detector that re-finds fewer weak faces look like a storage
// outage, which refused two of three real libraries.
const facesMigrateMaxFailureRatio = 0.1

// FacesMigrateOptions controls a face embedding migration.
type FacesMigrateOptions struct {
	Target    string
	DryRun    bool
	Force     bool
	BatchSize int

	// Plan carries a summary the caller already computed, so a command that reports the
	// scope before prompting does not run the counting queries a second time. It is
	// revalidated rather than trusted.
	Plan *FacesMigratePlan
}

// FacesMigratePlan summarizes the work required for a target embedding model.
type FacesMigratePlan struct {
	Target          string
	Markers         query.FaceMigrationMarkerCounts
	MarkerModels    []query.MarkerEmbeddingModelCount
	FaceModels      []query.EmbeddingModelCount
	Subjects        int
	AssignedMarkers int

	// LowQualitySamples counts assignments too small or too poorly scored to seed a
	// replacement centroid. They keep their person; they just cannot define one.
	LowQualitySamples int

	// RecropMarkers counts markers already on the target model whose crop another detector
	// placed, or that record none. They are re-embedded so the library ends up in one crop
	// space, and are reported apart from stale markers because they lose nothing if that fails.
	// On the first run after the provenance column was added, this is every marker.
	RecropMarkers int

	// OriginalsUnavailable reports that the originals root cannot be read, which is what an
	// unmounted volume looks like. The counting queries cannot see it, so a plan would
	// otherwise be reported as clean and then fail on every file it tried to re-embed.
	OriginalsUnavailable bool
}

// FacesMigrateResult summarizes a completed face embedding migration.
type FacesMigrateResult struct {
	Target   string
	Migrated int
	Skipped  int
	Retained int
	Failed   int

	// Unreadable counts the failed markers whose file could not be read, as distinct from the
	// ones the detector simply did not find again. The two need different responses, and a
	// message that names neither sends a diagnosis toward the filesystem.
	Unreadable int

	// FailedClusterable counts the failed markers that cleared both clustering bars, which is
	// what the finalize guard measures: the rest lose a vector nothing would have used.
	FailedClusterable int

	Unlinked          int
	Invalid           int
	DetectedFiles     int
	PreservedSubjects int
	PreservedMarkers  int
	HiddenClusters    int
	RebuiltSubjects   int
	ExcludedMarkers   int
	LowQualityMarkers int
	AttentionSubjects int
}

// FacesMigrateIncompleteError reports a migration that completed with marker failures.
type FacesMigrateIncompleteError struct {
	Failed int
}

// Error returns the incomplete migration error message.
func (e *FacesMigrateIncompleteError) Error() string {
	return fmt.Sprintf("faces: migration completed with %d failed marker(s)", e.Failed)
}

// FacesMigrateAbortedError reports a migration that was refused before clusters were replaced.
type FacesMigrateAbortedError struct {
	Migrated int
	Failed   int
	Reason   string
}

// Error returns the aborted migration error message.
//
// It names what the run already committed, because embeddings are checkpointed per file:
// an operator told the index was untouched would not know that those people are missing
// faces until the migration is completed.
func (e *FacesMigrateAbortedError) Error() string {
	return fmt.Sprintf("faces: refused to finalize the migration because %s; clusters were not replaced, "+
		"but %d regenerated marker(s) stay unmatched until a run completes, so re-run once the cause is "+
		"addressed (use --force to finalize anyway, which keeps every person assignment and loses only "+
		"the vectors that could not be regenerated)", e.Reason, e.Migrated)
}

// FacesMigrateSettingError reports a completed migration whose model could not be recorded.
type FacesMigrateSettingError struct {
	Target string
	Cause  error
}

// Error returns the unrecorded migration error message.
//
// The vectors are the target's from here on, so an instance that starts again with the previous
// model hides the whole library from matching. Naming the file is what makes that recoverable
// without running the migration a second time.
func (e *FacesMigrateSettingError) Error() string {
	return fmt.Sprintf("faces: migrated to %s, but %s; set FaceModel: %s in options.yml before "+
		"starting again, or the library stays hidden from matching", e.Target, e.Cause, e.Target)
}

// Unwrap returns the error that prevented the setting from being saved.
func (e *FacesMigrateSettingError) Unwrap() error {
	return e.Cause
}

// FacesMigrateRerunError reports a migration whose finalize was rolled back after markers
// had already been regenerated, which leaves the library between two models.
type FacesMigrateRerunError struct {
	Migrated int
	Cause    error
}

// Error returns the rolled-back migration error message.
//
// The rollback itself is the safe outcome, so the message is about the state it leaves:
// the clusters are the old model's and some markers are the new one's, and only another
// run reconciles them.
func (e *FacesMigrateRerunError) Error() string {
	return fmt.Sprintf("faces: %s, so replacing the clusters was rolled back and nothing was lost; "+
		"%d regenerated marker(s) stay unmatched until the migration is run again with the server stopped",
		e.Cause, e.Migrated)
}

// Unwrap returns the error that caused the rollback.
func (e *FacesMigrateRerunError) Unwrap() error {
	return e.Cause
}

// finalizeRefused reports why a migration must not be finalized, or "" when it may proceed.
// Attempting nothing is not a failure: a library that is already migrated skips every
// marker, and refusing there would make the command permanently unusable.
func finalizeRefused(result FacesMigrateResult) string {
	attempted := result.Migrated + result.Failed

	switch {
	case attempted == 0:
		return ""
	case result.Migrated == 0:
		return fmt.Sprintf("none of %d marker(s) could be re-embedded%s", attempted, migrationFailureCause(result))
	case float64(result.FailedClusterable)/float64(attempted) > facesMigrateMaxFailureRatio:
		return fmt.Sprintf("%d of %d marker(s) that clear the clustering thresholds could not be re-embedded%s",
			result.FailedClusterable, attempted, migrationFailureCause(result))
	default:
		return ""
	}
}

// migrationFailureCause names what the failures have in common, or "" when there were none.
//
// A file that cannot be read is a storage problem; a face the detector declines to find again
// will fail identically on every re-run. They ask for opposite responses, so naming neither
// sent this diagnosis toward the filesystem for several steps.
func migrationFailureCause(result FacesMigrateResult) string {
	undetected := result.Failed - result.Unreadable

	switch {
	case result.Failed == 0:
		return ""
	case result.Unreadable == 0:
		return ", because the detector did not find their face again"
	case undetected <= 0:
		return ", because their file is missing or unreadable"
	default:
		return fmt.Sprintf(", %d because the detector did not find their face again and %d because "+
			"their file is missing or unreadable", undetected, result.Unreadable)
	}
}

// migrationTarget resolves the target model and reports why it cannot be migrated to.
// Every check reads configuration rather than the index, so it is cheap enough to repeat.
func (w *Faces) migrationTarget(target string) (string, error) {
	if w == nil || w.conf == nil {
		return "", fmt.Errorf("faces: configuration not available")
	} else if w.Disabled() {
		return "", fmt.Errorf("face recognition is disabled")
	}

	target = face.NormalizeModelName(target)

	// Exactly one embedding model is officially supported, so the ordinary migration needs no
	// target and defaults to it. Any other model has to be named. An instance where embeddings
	// were switched off keeps that decision: defaulting there would turn "off" into a data
	// migration nobody asked for.
	if target == "" || target == face.ModelAuto || target == face.ModelDetect {
		if w.conf.FaceModelSetting() == face.ModelNone {
			target = face.ModelNone
		} else {
			target = face.DefaultModelName()
		}
	}

	switch {
	case !face.KnownModelName(target) || target == face.ModelDetect:
		return "", fmt.Errorf("faces: unsupported migration model %q", target)
	case target == face.ModelNone:
		return "", fmt.Errorf("faces: cannot migrate to disabled embeddings")
	}

	// A migration is how the model is changed, so the target is not required to match what
	// is configured: it regenerates every vector in the target's space and then records it.
	if err := face.LicenseRefused(target, w.conf.Edition()); err != nil {
		return "", fmt.Errorf("faces: %s", err)
	}

	model := face.FindEmbeddingModel(target)
	if model == nil || !model.Installed(w.conf.ModelsPath()) {
		return "", fmt.Errorf("faces: embedding model %s is not installed", clean.Log(target))
	}

	if model.Aligned() && face.ActiveEngineName() != face.EngineONNX {
		return "", fmt.Errorf("faces: migration to %s requires the ONNX face detector", clean.Log(target))
	}

	return target, nil
}

// PlanMigration validates the target and returns a read-only migration summary.
func (w *Faces) PlanMigration(target string) (result FacesMigratePlan, err error) {
	target, err = w.migrationTarget(target)

	if err != nil {
		return result, err
	}

	result.Target = target

	if result.Markers, err = query.FaceMigrationCounts(target); err != nil {
		return result, err
	} else if result.MarkerModels, err = query.MarkerEmbeddingModels(); err != nil {
		return result, err
	} else if result.FaceModels, err = query.FaceEmbeddingModels(); err != nil {
		return result, err
	}

	identities, err := query.FaceMigrationIdentities()
	if err != nil {
		return result, err
	}

	result.Subjects, result.AssignedMarkers = faceMigrationSubjectCounts(identities)

	if result.LowQualitySamples, err = query.FaceMigrationLowQualityMarkers(target); err != nil {
		return result, err
	}

	// Only an aligned model consumes landmarks, so only there does the detector that placed them
	// decide the crop and with it the vector. A crop-based model reads stored geometry instead.
	if model := face.FindEmbeddingModel(target); model != nil && model.Aligned() {
		if result.RecropMarkers, err = query.FaceMigrationRecropMarkers(target, face.ActiveDetector()); err != nil {
			return result, err
		}
	}

	result.OriginalsUnavailable = originalsUnavailable(w.conf.OriginalsPath())

	return result, nil
}

// originalsUnavailable reports whether the originals root is missing or holds nothing, at the cost of
// a stat and one directory entry. It answers the case a per-file check cannot improve on: when the
// volume is not mounted every marker fails, and the plan is the last point at which that is cheap.
func originalsUnavailable(originalsPath string) bool {
	return !fs.PathExists(originalsPath) || fs.DirIsEmpty(originalsPath)
}

// faceMigrationSubjectCounts returns how many subjects the identities name and how many markers
// are assigned to one.
//
// Keyed by subject alone: a named marker with no subject row yet is not a person a cluster can be
// rebuilt for, and counting it would report it as needing attention moments before it exists.
func faceMigrationSubjectCounts(identities []query.FaceMigrationIdentity) (subjects, assigned int) {
	seen := make(map[string]struct{}, len(identities))

	for _, identity := range identities {
		if identity.SubjUID == "" {
			continue
		}

		seen[identity.SubjUID] = struct{}{}
		assigned++
	}

	return len(seen), assigned
}

// migratePlan reuses the caller's summary when it has one, and computes it otherwise.
//
// Only the counting queries are skipped: the target is validated either way, so a plan
// handed in from elsewhere cannot decide what gets migrated.
func (w *Faces) migratePlan(opt FacesMigrateOptions) (FacesMigratePlan, error) {
	if opt.Plan == nil {
		return w.PlanMigration(opt.Target)
	}

	target, err := w.migrationTarget(opt.Target)

	if err != nil {
		return FacesMigratePlan{}, err
	}

	if opt.Plan.Target != target {
		return FacesMigratePlan{}, fmt.Errorf("faces: migration plan targets %s, expected %s",
			clean.Log(opt.Plan.Target), clean.Log(target))
	}

	return *opt.Plan, nil
}

// Migrate regenerates face marker embeddings and rebuilds all clusters for the target model.
func (w *Faces) Migrate(ctx context.Context, opt FacesMigrateOptions) (result FacesMigrateResult, err error) {
	plan, err := w.migratePlan(opt)
	if err != nil {
		return result, err
	}

	result.Target = plan.Target
	result.Invalid = plan.Markers.Invalid
	result.PreservedSubjects = plan.Subjects
	result.PreservedMarkers = plan.AssignedMarkers
	result.LowQualityMarkers = plan.LowQualitySamples

	if opt.DryRun {
		return result, nil
	}

	// These guards are process-local, so they only catch a worker started in this same
	// process. The CLI runs in its own and cannot see a running server's workers.
	if mutex.IndexWorker.Running() || mutex.MetaWorker.Running() || mutex.VisionWorker.Running() {
		return result, fmt.Errorf("faces: indexing or vision worker is already running")
	} else if err = mutex.FacesWorker.Start(); err != nil {
		return result, err
	}
	defer mutex.FacesWorker.Stop()

	// The lock file is what a server in another process can see. It carries its own expiry, so
	// a run that is killed releases it rather than leaving the instance unable to index faces.
	lock, err := mutex.AcquireFileLock(w.conf.FacesLockFile(), "faces migration")
	if err != nil {
		return result, fmt.Errorf("faces: %s", err)
	}
	defer lock.Release()

	embedder, err := w.migrationEmbedder(plan.Target)
	if err != nil {
		return result, err
	}

	// A run that does not reach the point where the setting follows the data has to leave the
	// previous model loaded. After one that committed this is a no-op, since the setting
	// names the target by then.
	defer w.restoreEmbedder()

	restoreDetector, err := w.useMigrationDetector()
	if err != nil {
		return result, err
	}
	defer restoreDetector()

	return w.migrate(ctx, plan, embedder, opt, result)
}

// migrate performs the migration itself, once the plan and the embedder are resolved.
func (w *Faces) migrate(ctx context.Context, plan FacesMigratePlan, embedder face.Embedder, opt FacesMigrateOptions, result FacesMigrateResult) (FacesMigrateResult, error) {
	identities, err := query.FaceMigrationIdentities()
	if err != nil {
		return result, err
	}

	result.PreservedSubjects, result.PreservedMarkers = faceMigrationSubjectCounts(identities)

	// Captured before the clusters are replaced, because that is the last moment the
	// operator's hide decisions still exist anywhere.
	hiddenMarkers, err := query.HiddenFaceMarkers()
	if err != nil {
		return result, err
	}

	// Markers without a file cannot be re-embedded by any run, so they are reported apart
	// from failures: counting them would make every retry look partially failed.
	result.Unlinked = plan.Markers.Unlinked

	var failedMarkerUIDs []string
	batchSize := opt.BatchSize
	if batchSize < 1 {
		batchSize = facesMigrateBatchSize
	}

	after := ""
	for {
		if err = migrationCanceled(ctx, w); err != nil {
			return result, err
		}

		fileUIDs, queryErr := query.FaceMigrationFileUIDs(after, batchSize)
		if queryErr != nil {
			return result, queryErr
		} else if len(fileUIDs) == 0 {
			break
		}

		for _, fileUID := range fileUIDs {
			if err = migrationCanceled(ctx, w); err != nil {
				return result, err
			}

			fileResult, migrateErr := w.migrateFaceFile(embedder, plan.Target, fileUID)
			result.Migrated += fileResult.Migrated
			result.Skipped += fileResult.Skipped
			result.Retained += fileResult.Retained
			result.Failed += len(fileResult.Failed)
			result.Unreadable += fileResult.Unreadable
			result.FailedClusterable += fileResult.Clusterable
			failedMarkerUIDs = append(failedMarkerUIDs, fileResult.Failed...)
			if fileResult.Detected {
				result.DetectedFiles++
			}
			if migrateErr != nil {
				if len(fileResult.Failed) == 0 && fileResult.Retained == 0 {
					return result, migrateErr
				}

				// The markers this file holds are counted as failed rather than lost, and the
				// plan reported them as unreadable before the prompt, so this is the predicted
				// outcome rather than a fault: the run continues and the exit status carries it.
				// Reported through the system log, since a run over a library with missing files
				// emits one of these per file and the ordinary log persists warnings.
				event.SystemWarn([]string{"faces", "migrate", "%s, so %s could not be re-embedded"},
					migrateErr, english.Plural(len(fileResult.Failed)+fileResult.Retained, "marker", "markers"))
			}
		}

		after = fileUIDs[len(fileUIDs)-1]

		// A batch boundary is the natural cadence: the loop is cursor-paged, so it needs no timer,
		// and a run over a large library is otherwise silent for the best part of an hour.
		logMigrationProgress(plan, result)
	}

	// The finalize below is destructive and irreversible, so a run that regenerated nothing
	// must not reach it: the old vectors are the only copy that exists.
	if reason := finalizeRefused(result); reason != "" {
		if !opt.Force {
			return result, &FacesMigrateAbortedError{Migrated: result.Migrated, Failed: result.Failed, Reason: reason}
		}

		event.SystemWarn([]string{"faces", "migrate", "finalizing the migration anyway because %s"}, reason)
	}

	clusters, rebuilt, excluded, clusterErr := buildFaceMigrationClusters(plan.Target)
	if clusterErr != nil {
		return result, clusterErr
	}
	result.RebuiltSubjects = rebuilt
	result.ExcludedMarkers = excluded
	result.AttentionSubjects = max(result.PreservedSubjects-result.RebuiltSubjects, 0)

	// Any failure here rolls the whole finalize back, so the clusters are still the old
	// model's while the markers this run regenerated are the target's. That is recoverable
	// only by running again, which the operator has to be told rather than left to infer.
	if err = query.FinalizeFaceMigration(plan.Target, identities, clusters, failedMarkerUIDs); err != nil {
		return result, &FacesMigrateRerunError{Migrated: result.Migrated, Cause: err}
	}

	// The vectors are in the target's space from here on, so the setting follows the data even
	// when markers failed: a start that read the previous model would hide the whole library
	// from matching. Both have to happen before the clustering below, which is gated on them,
	// and a write that failed is reported once the run has finished rather than losing it.
	settingErr := w.conf.SetFaceModel(plan.Target)
	face.UnblockEmbeddings()

	entity.UpdateFaces.Store(true)
	if err = w.start(FacesOptions{Force: true}); err != nil {
		return result, unrecordedFaceModel(settingErr, plan.Target, err)
	} else if err = w.Audit(false, ""); err != nil {
		return result, unrecordedFaceModel(settingErr, plan.Target, err)
	}

	// Re-clustering has run by now, so the replacement clusters exist to be hidden again.
	if result.HiddenClusters, err = query.RestoreHiddenFaces(hiddenMarkers); err != nil {
		return result, unrecordedFaceModel(settingErr, plan.Target, err)
	}

	// Subject covers and counts are denormalized, so without this the People view keeps
	// advertising the totals the library had before the clusters were replaced.
	if updateErr := query.UpdateSubjectCovers(true); updateErr != nil {
		event.SystemError([]string{"faces", "migrate", "%s (update covers)"}, updateErr)
	}

	if updateErr := entity.UpdateSubjectCounts(true); updateErr != nil {
		event.SystemError([]string{"faces", "migrate", "%s (update counts)"}, updateErr)
	}

	if err = unrecordedFaceModel(settingErr, plan.Target, nil); err != nil {
		return result, err
	}

	if result.Failed > 0 {
		return result, &FacesMigrateIncompleteError{Failed: result.Failed}
	}

	return result, nil
}

// migrationEmbedder loads the embedder that writes the target's vectors and validates its
// contract. It configures the target itself rather than taking the active model, which is what
// lets a migration run on an instance still configured for the model it replaces.
func (w *Faces) migrationEmbedder(target string) (embedder face.Embedder, err error) {
	if vision.Config == nil {
		return nil, fmt.Errorf("faces: vision configuration not available")
	}

	// A target that cannot be loaded, or that turns out not to honor its contract, must not
	// leave the instance without the model it was running on.
	defer func() {
		if err != nil {
			w.restoreEmbedder()
		}
	}()

	if err = w.conf.ConfigureFaceEmbedder(target); err != nil {
		return nil, fmt.Errorf("faces: embedding model %s could not be loaded (%s)", clean.Log(target), err)
	}

	model := vision.Config.Model(vision.ModelTypeFace)
	if model == nil {
		return nil, fmt.Errorf("faces: face vision model not configured")
	}

	embedder = model.MigrationFaceModel()
	if embedder == nil {
		return nil, fmt.Errorf("faces: embedding model %s could not be loaded", clean.Log(target))
	} else if embedder.ModelName() != target {
		return nil, fmt.Errorf("faces: loaded embedding model %s, expected %s", clean.Log(embedder.ModelName()), clean.Log(target))
	}

	registered := face.FindEmbeddingModel(target)
	if registered == nil || embedder.Dims() != registered.Dims {
		return nil, fmt.Errorf("faces: embedding model %s has unexpected dimensions", clean.Log(target))
	}

	return embedder, nil
}

// unrecordedFaceModel keeps a setting that could not be saved attached to whatever failed after
// it, or returns that error unchanged when the setting was saved.
//
// It outranks what follows, because the setting decides whether the library is matched at all and
// the two failures are correlated: a full disk or a read-only volume causes both.
func unrecordedFaceModel(settingErr error, target string, err error) error {
	if settingErr == nil {
		return err
	}

	return errors.Join(&FacesMigrateSettingError{Target: target, Cause: settingErr}, err)
}

// restoreEmbedder loads the embedding model the configuration names, which is the target after
// a migration committed and the previous model otherwise.
func (w *Faces) restoreEmbedder() {
	if err := w.conf.ConfigureFaceEmbedder(w.conf.FaceModel()); err != nil {
		event.SystemWarn([]string{"faces", "migrate", "%s (restore embedding model)"}, err)
	}
}

// useMigrationDetector lowers the detection floor for the duration of a migration and returns the
// function that restores the configured one.
//
// A false positive costs an index a thumbnail to reject; a miss costs a migration a curated
// marker's vector, so the run detects at the lowest floor the detectors have shipped with. A
// configured FACE_SCORE stands: that is a decision rather than a calibration.
func (w *Faces) useMigrationDetector() (restore func(), err error) {
	restore = func() {}

	if w == nil || w.conf == nil {
		return restore, nil
	}

	score := w.conf.FaceScore()

	if score <= 0 {
		score = face.MigrationScoreThreshold
	}

	// The cutoff lives in the inference session and in the filter Detect applies afterwards, so
	// both have to move or the second undoes the first.
	previous := face.ScoreThreshold

	if err = w.conf.ConfigureFaceDetector(score); err != nil {
		return restore, fmt.Errorf("faces: %s (configure detector)", err)
	}

	face.ScoreThreshold = score

	return func() {
		face.ScoreThreshold = previous

		if restoreErr := w.conf.ConfigureFaceDetector(0); restoreErr != nil {
			event.SystemWarn([]string{"faces", "migrate", "%s (restore detector)"}, restoreErr)
		}
	}, nil
}

// faceMigrationFile reports what one file's markers contributed to a migration.
//
// Retained markers keep the vector they already hold, so they are neither migrated nor failed:
// counting them as either would make a detector change look like data loss or like work done.
type faceMigrationFile struct {
	Migrated int
	Skipped  int
	Retained int
	Failed   []string
	// Unreadable counts the failed markers this file lost because it could not be read at all,
	// as distinct from the ones a successful detection did not find again.
	Unreadable int
	// Clusterable counts the failed markers that cleared both clustering bars.
	Clusterable int
	Detected    bool
}

// migrateFaceFile checkpoints all stale marker embeddings associated with a file.
func (w *Faces) migrateFaceFile(embedder face.Embedder, target, fileUID string) (result faceMigrationFile, err error) {
	markers, err := query.FaceMigrationMarkers(fileUID)
	if err != nil {
		return result, err
	}

	stale, recrop := staleMigrationMarkers(markers, embedder, target)
	result.Skipped = len(markers) - len(stale)

	if len(stale) == 0 {
		return result, nil
	}

	file, err := query.FileByUID(fileUID)
	if err != nil {
		// The lookup is scoped, so a file the library removed is simply absent, and "record
		// not found" would read as a database fault rather than the state the plan counted.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("file was deleted")
		}

		result.Failed, result.Retained, result.Clusterable = unresolvedMigrationMarkers(stale, recrop)
		result.Unreadable = len(result.Failed)

		return result, err
	}

	var generated map[string]face.Embeddings
	var landmarks map[string]json.RawMessage

	// A blank detector means no detection ran, so the one already recorded still describes
	// the crop. It is taken from the detection rather than from the active engine, which by
	// the time this returns may name whatever a concurrent reconfiguration left loaded.
	detectModel := ""

	if embedder.Aligned() {
		result.Detected = true
		generated, landmarks, detectModel, err = w.detectMigrationEmbeddings(embedder, file, markers, stale)
	} else {
		generated, err = w.cropMigrationEmbeddings(embedder, file, stale)
	}
	if err != nil {
		// Detection reads the file, so a failure here is the file rather than the detector
		// declining to find a face in it.
		result.Failed, result.Retained, result.Clusterable = unresolvedMigrationMarkers(stale, recrop)
		result.Unreadable = len(result.Failed)

		return result, err
	}

	unresolved := make(entity.Markers, 0, len(stale))

	for _, marker := range stale {
		if values, ok := generated[marker.MarkerUID]; ok && face.ValidEmbeddings(values, embedder.Dims()) {
			result.Migrated++
			continue
		}

		unresolved = append(unresolved, marker)
		delete(generated, marker.MarkerUID)
		delete(landmarks, marker.MarkerUID)
	}

	result.Failed, result.Retained, result.Clusterable = unresolvedMigrationMarkers(unresolved, recrop)

	if len(generated) > 0 {
		if err = query.SaveFaceMigrationEmbeddings(target, detectModel, generated, landmarks); err != nil {
			return faceMigrationFile{Skipped: result.Skipped, Detected: result.Detected}, err
		}
	}

	return result, nil
}

// staleMigrationMarkers returns the markers a migration to target has to re-embed, and the subset
// of them that only needs a new crop.
//
// The crop is an axis of the embedding space, so an unrecorded or foreign detector makes a marker
// stale even when its vector is the target's: leaving it puts one library in two crop spaces. A
// first run therefore re-crops everything, which costs time and risks nothing.
func staleMigrationMarkers(markers entity.Markers, embedder face.Embedder, target string) (stale entity.Markers, recrop map[string]bool) {
	stale = make(entity.Markers, 0, len(markers))
	recrop = make(map[string]bool)
	detector := face.ActiveDetector()

	for i := range markers {
		if !face.ValidEmbeddings(markers[i].Embeddings(), embedder.Dims()) || !face.ModelsComparable(markers[i].EmbedModel, target) {
			stale = append(stale, markers[i])
			continue
		}

		// Only an aligned embedder consumes landmarks, so only there does the detector decide
		// the crop. A crop-based one reads the stored geometry whichever detector wrote it.
		if !embedder.Aligned() || face.DetectorsComparable(markers[i].DetectModel, detector) {
			continue
		}

		recrop[markers[i].MarkerUID] = true
		stale = append(stale, markers[i])
	}

	return stale, recrop
}

// unresolvedMigrationMarkers splits the markers whose embeddings could not be regenerated into
// those that must be discarded, the number that keep a usable vector, and how many of the
// discarded ones the library could have clustered.
//
// Detection finds no marker a person drew by hand, so re-cropping one is expected to fail. Its
// vector is still in the target's space, and discarding it would make a detector change cost
// exactly the assignments the migration exists to preserve.
func unresolvedMigrationMarkers(markers entity.Markers, recrop map[string]bool) (failed []string, retained, clusterable int) {
	for i := range markers {
		if recrop[markers[i].MarkerUID] {
			retained++
			continue
		}

		failed = append(failed, markers[i].MarkerUID)

		if markers[i].Clusterable() {
			clusterable++
		}
	}

	return failed, retained, clusterable
}

// cropMigrationEmbeddings generates embeddings directly from stored marker geometry.
func (w *Faces) cropMigrationEmbeddings(embedder face.Embedder, file *entity.File, markers entity.Markers) (result map[string]face.Embeddings, err error) {
	result = make(map[string]face.Embeddings, len(markers))
	if w == nil || w.conf == nil || embedder == nil || file == nil {
		return result, fmt.Errorf("faces: migration crop input is invalid")
	}

	width, height := embedder.CropSize()
	size := crop.Size{Width: width, Height: height, Options: crop.DefaultOptions}
	source := ConfigFileName(w.conf, file.FileRoot, file.FileName)

	for _, marker := range markers {
		area := markerCropArea(marker)
		thumbName, thumbErr := crop.ThumbFileName(file.FileHash, area, size, w.conf.ThumbCachePath())
		if thumbErr != nil {
			thumbName = source
		}

		img, _, cropErr := crop.ImageFromThumb(thumbName, area, size, false)
		if cropErr != nil {
			event.SystemWarn([]string{"faces", "migrate", "failed cropping marker %s (%s)"}, clean.Log(marker.MarkerUID), cropErr)
			continue
		}

		if embeddings := embedder.Run(img); face.ValidEmbeddings(embeddings, embedder.Dims()) {
			result[marker.MarkerUID] = embeddings
		}
	}

	return result, nil
}

// detectMigrationEmbeddings redetects a file and maps aligned embeddings to stored markers, also
// reporting the landmarks each detection placed and the detector that placed them.
//
// Provenance travels with the vectors rather than being read back from configuration, and the
// landmarks are what make it usable: without them the column attests the crop alone.
func (w *Faces) detectMigrationEmbeddings(embedder face.Embedder, file *entity.File, markers, stale entity.Markers) (result map[string]face.Embeddings, landmarks map[string]json.RawMessage, detectModel string, err error) {
	result = make(map[string]face.Embeddings, len(stale))
	landmarks = make(map[string]json.RawMessage, len(stale))

	thumbName, err := migrationDetectionThumb(w.conf, w.conf.ThumbCachePath(), file)
	if err != nil {
		return result, landmarks, "", err
	}

	// At the smallest size the detectors are trained for rather than at FACE_SIZE: a marker's size
	// is in pixels of the thumbnail it was detected in, and the detector before this one fell back
	// to a larger one, so a legacy marker can sit well under the ordinary floor - which no score
	// recovers. The retry pass only fires when a picture yields nothing, which is not this case.
	detected, err := face.Detect(thumbName, face.MinSizeThreshold)
	if err != nil {
		return result, landmarks, "", err
	}

	// Only the detections a stored marker claims are embedded. Everything else this floor admits
	// would be inferred and thrown away, and the smallest face decides the crop rendition for the
	// whole file, so embedding them all would make the real crops worse as well as slower.
	assigned, order := assignedMigrationDetections(markers, stale, detected)
	vision.GenerateEmbeddings(embedder, thumbName, assigned, true)

	for markerUID, i := range order {
		if detectedFace := assigned[i]; face.ValidEmbeddings(detectedFace.Embeddings, embedder.Dims()) {
			result[markerUID] = detectedFace.Embeddings
			detectModel = detectedFace.DetectModel

			if points := detectedFace.RelativeLandmarksJSON(); len(points) > 0 {
				landmarks[markerUID] = points
			}
		}
	}

	return result, landmarks, detectModel, nil
}

// assignedMigrationDetections returns the detections the stale markers claim, and where each
// marker's own detection sits in that slice. Matching considers every marker the file holds, so
// a detection a marker needing no work accounts for is not handed to a second marker as well.
func assignedMigrationDetections(markers, stale entity.Markers, detected face.Faces) (assigned face.Faces, order map[string]int) {
	assignments := matchMigrationDetections(markers, detected)
	assigned = make(face.Faces, 0, len(stale))
	order = make(map[string]int, len(stale))

	for _, marker := range stale {
		if i, ok := assignments[marker.MarkerUID]; ok {
			order[marker.MarkerUID] = len(assigned)
			assigned = append(assigned, detected[i])
		}
	}

	return assigned, order
}

// migrationDetectionThumb returns a cached or newly generated detection thumbnail.
func migrationDetectionThumb(conf *config.Config, thumbPath string, file *entity.File) (string, error) {
	if file == nil || file.FileHash == "" {
		return "", fmt.Errorf("faces: migration file is invalid")
	}

	size := thumb.Sizes[thumb.Fit720]
	if cached, err := size.FileName(file.FileHash, thumbPath); err == nil && fs.FileExists(cached) {
		return cached, nil
	}

	mediaFile, err := NewMediaFile(ConfigFileName(conf, file.FileRoot, file.FileName))
	if err != nil {
		return "", err
	}

	return mediaFile.Thumbnail(thumbPath, thumb.Fit720)
}

type migrationDetectionPair struct {
	markerUID string
	detected  int
	overlap   int
}

// matchMigrationDetections assigns each detected face to at most one stored marker, and returns
// the index of the detection each marker was given.
func matchMigrationDetections(markers entity.Markers, detected face.Faces) map[string]int {
	pairs := make([]migrationDetectionPair, 0)
	for _, marker := range markers {
		area := markerCropArea(marker)
		for i := range detected {
			if overlap := detected[i].CropArea().OverlapPercent(area); overlap > face.OverlapThresholdFloor {
				pairs = append(pairs, migrationDetectionPair{markerUID: marker.MarkerUID, detected: i, overlap: overlap})
			}
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].overlap != pairs[j].overlap {
			return pairs[i].overlap > pairs[j].overlap
		} else if pairs[i].markerUID != pairs[j].markerUID {
			return pairs[i].markerUID < pairs[j].markerUID
		}
		return pairs[i].detected < pairs[j].detected
	})

	result := make(map[string]int)
	used := make(map[int]bool)
	for _, pair := range pairs {
		if _, ok := result[pair.markerUID]; ok || used[pair.detected] {
			continue
		}
		result[pair.markerUID] = pair.detected
		used[pair.detected] = true
	}

	return result
}

// markerCropArea returns the normalized crop geometry stored on a marker.
func markerCropArea(marker entity.Marker) crop.Area {
	return crop.Area{Name: "face", X: marker.X, Y: marker.Y, W: marker.W, H: marker.H}
}

// buildFaceMigrationClusters creates one replacement cluster per identified subject, seeded from
// the assignments that agree with their own midpoint, and reports how many were left out.
//
// One subject at a time, so that a whole library's embedding blobs never have to be resident.
func buildFaceMigrationClusters(model string) (result []query.FaceMigrationCluster, rebuilt, excluded int, err error) {
	registered := face.FindEmbeddingModel(model)
	if registered == nil {
		return result, 0, 0, fmt.Errorf("faces: unsupported migration model %s", clean.Log(model))
	}

	subjectUIDs, err := query.FaceMigrationSubjectUIDs(model)
	if err != nil {
		return result, 0, 0, err
	}

	for _, subjectUID := range subjectUIDs {
		markers, markerErr := query.FaceMigrationSubjectMarkers(model, subjectUID)
		if markerErr != nil {
			return result, rebuilt, excluded, markerErr
		}

		group := make(entity.Markers, 0, len(markers))
		for _, marker := range markers {
			if face.ValidEmbeddings(marker.Embeddings(), registered.Dims) {
				group = append(group, marker)
			}
		}

		if len(group) == 0 {
			continue
		}

		samples, dropped := centroidSamples(group, registered, subjectUID)
		excluded += dropped

		embeddings := make(face.Embeddings, 0, len(samples))
		for _, marker := range samples {
			if values := marker.Embeddings(); values.One() {
				embeddings = append(embeddings, values[0])
			}
		}

		if len(embeddings) == 0 {
			continue
		}

		cluster := entity.NewFace(subjectUID, entity.SrcManual, embeddings, model)
		if cluster == nil || cluster.ID == "" || cluster.EmbedModel != model || len(cluster.Embedding()) == 0 {
			event.SystemWarn([]string{"faces", "migrate", "failed rebuilding cluster for subject %s"}, clean.Log(subjectUID))
			continue
		}

		distances := make(map[string]float64, len(samples))
		for _, marker := range samples {
			distances[marker.MarkerUID] = marker.Embeddings().Dist(cluster.Embedding())
		}

		result = append(result, query.FaceMigrationCluster{Face: *cluster, MarkerDistances: distances})
		rebuilt++
	}

	return result, rebuilt, excluded, nil
}

// manualSubjectAssignment reports whether a person set this marker's subject rather than
// the matcher. Auto- and XMP-sourced assignments are the ones a previous model may have got
// wrong, which is what the outlier rule is looking for.
func manualSubjectAssignment(subjSrc string) bool {
	return subjSrc != "" && subjSrc != entity.SrcAuto && subjSrc != entity.SrcXmp
}

// centroidSamples returns the markers that may define a subject's replacement cluster, and how
// many assignments were left out. An outlier keeps its assignment and loses only its vote.
//
// A hand-set assignment is exempt from the ordinary outlier distance, which there measures the
// model rather than the assignment, but not from the widest distance the cluster can accept.
func centroidSamples(group entity.Markers, registered *face.EmbeddingModel, subjectUID string) (kept entity.Markers, excluded int) {
	if len(group) < 3 {
		return group, 0
	}

	embeddings := make(face.Embeddings, 0, len(group))
	for _, marker := range group {
		if values := marker.Embeddings(); values.One() {
			embeddings = append(embeddings, values[0])
		}
	}

	midpoint, _, _ := face.EmbeddingsMidpoint(embeddings)
	if len(midpoint) == 0 {
		return group, 0
	}

	// The widest distance a cluster of this model can ever accept. SampleRadius is clamped
	// to ClusterRadius where it is written and again where it is read, so no cluster reaches
	// further than this whatever its samples. Seeding beyond it would relink a marker to a
	// cluster that then refuses it, leaving an assignment matching would never have made.
	maxAccept := min(registered.ClusterRadius+registered.MatchDist, face.AcceptDistMax)

	kept = make(entity.Markers, 0, len(group))

	for _, marker := range group {
		// Embeddings.Dist reports -1 when nothing is comparable, which would otherwise read
		// as the closest possible sample.
		switch d := marker.Embeddings().Dist(midpoint); {
		case d < 0 || d > maxAccept:
			continue
		case d > registered.ClusterDist && !manualSubjectAssignment(marker.SubjSrc):
			continue
		}

		kept = append(kept, marker)
	}

	// Every sample disagreeing with their own midpoint means the assignments are too
	// scattered to tell signal from outlier, so none of them is dropped.
	if len(kept) < 2 {
		return group, 0
	}

	if excluded = len(group) - len(kept); excluded > 0 {
		log.Infof("faces: excluded %d outlier(s) from the cluster of subject %s", excluded, clean.Log(subjectUID))
	}

	return kept, excluded
}

// logMigrationProgress reports how far a run has got, against the total the plan named.
//
// Every processed state is counted rather than the migrated ones alone, so a re-run that skips
// most of the library still reports honestly instead of appearing to stall.
func logMigrationProgress(plan FacesMigratePlan, result FacesMigrateResult) {
	total := plan.Markers.Valid

	if total < 1 {
		return
	}

	done := min(result.Migrated+result.Skipped+result.Failed+result.Retained, total)

	log.Infof("faces: processed %d of %d markers (%d%%)", done, total, done*100/total)
}

// migrationCanceled reports context and worker cancellation consistently between batches.
func migrationCanceled(ctx context.Context, w *Faces) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	if w != nil && w.Canceled() {
		return fmt.Errorf("faces: migration canceled")
	}

	return nil
}
