package photoprism

import (
	"context"
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const facesMigrateBatchSize = 100

// facesMigrateMaxFailureRatio bounds the share of attempted markers that may fail before
// the destructive finalize is refused. A storage outage fails every marker it touches, so
// the bound separates a few unreadable files from a library that is not readable at all.
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
}

// FacesMigrateResult summarizes a completed face embedding migration.
type FacesMigrateResult struct {
	Target            string
	Migrated          int
	Skipped           int
	Failed            int
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
		"but %d regenerated marker(s) stay unmatched until a run completes, so re-run after fixing "+
		"storage (use --force to finalize anyway)", e.Reason, e.Migrated)
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
func finalizeRefused(migrated, failed int) string {
	attempted := migrated + failed

	switch {
	case attempted == 0:
		return ""
	case migrated == 0:
		return fmt.Sprintf("none of %d marker(s) could be re-embedded", attempted)
	case float64(failed)/float64(attempted) > facesMigrateMaxFailureRatio:
		return fmt.Sprintf("%d of %d marker(s) could not be re-embedded", failed, attempted)
	default:
		return ""
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

	configured := face.NormalizeModelName(w.conf.FaceModel())
	target = face.NormalizeModelName(target)

	if target == "" || target == face.ModelAuto {
		target = configured
	}

	switch {
	case !face.KnownModelName(target) || target == face.ModelAuto:
		return "", fmt.Errorf("faces: unsupported migration model %q", target)
	case target == face.ModelNone:
		return "", fmt.Errorf("faces: cannot migrate to disabled embeddings")
	case target != configured:
		// The configured model may have been resolved from "auto", which follows whichever
		// model owns most of the vectors and therefore points back at the old one while a
		// migration is only half done. Naming the way out keeps that from reading as a
		// refusal the operator cannot act on.
		return "", fmt.Errorf("faces: migration target %s does not match configured model %s, "+
			"set PHOTOPRISM_FACE_MODEL=%s and re-run to continue", clean.Log(target), clean.Log(configured), clean.Log(target))
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

	return result, nil
}

// faceMigrationSubjectCounts returns how many subjects the identities name and how many
// markers are assigned to one.
//
// Keyed by subject alone: a named marker that has no subject row yet is not a person the
// migration can rebuild a cluster for, and counting it here would report it as needing
// attention even though the rebuild creates its subject moments later.
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

	embedder, err := w.migrationEmbedder(plan.Target)
	if err != nil {
		return result, err
	}

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

			migrated, skipped, failed, detected, migrateErr := w.migrateFaceFile(embedder, plan.Target, fileUID)
			result.Migrated += migrated
			result.Skipped += skipped
			result.Failed += len(failed)
			failedMarkerUIDs = append(failedMarkerUIDs, failed...)
			if detected {
				result.DetectedFiles++
			}
			if migrateErr != nil {
				if len(failed) == 0 {
					return result, migrateErr
				}
				log.Errorf("faces: %s while migrating file %s", migrateErr, clean.Log(fileUID))
			}
		}

		after = fileUIDs[len(fileUIDs)-1]
	}

	// The finalize below is destructive and irreversible, so a run that regenerated nothing
	// must not reach it: the old vectors are the only copy that exists.
	if reason := finalizeRefused(result.Migrated, result.Failed); reason != "" {
		if !opt.Force {
			return result, &FacesMigrateAbortedError{Migrated: result.Migrated, Failed: result.Failed, Reason: reason}
		}

		log.Warnf("faces: finalizing the migration anyway because %s", reason)
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

	entity.UpdateFaces.Store(true)
	if err = w.start(FacesOptions{Force: true}); err != nil {
		return result, err
	} else if err = w.Audit(false, ""); err != nil {
		return result, err
	}

	// Re-clustering has run by now, so the replacement clusters exist to be hidden again.
	if result.HiddenClusters, err = query.RestoreHiddenFaces(hiddenMarkers); err != nil {
		return result, err
	}

	// Subject covers and counts are denormalized, so without this the People view keeps
	// advertising the totals the library had before the clusters were replaced.
	if updateErr := query.UpdateSubjectCovers(true); updateErr != nil {
		log.Errorf("faces: %s (update covers)", updateErr)
	}

	if updateErr := entity.UpdateSubjectCounts(true); updateErr != nil {
		log.Errorf("faces: %s (update counts)", updateErr)
	}

	if result.Failed > 0 {
		return result, &FacesMigrateIncompleteError{Failed: result.Failed}
	}

	return result, nil
}

// migrationEmbedder returns the configured local embedder after validating its contract.
func (w *Faces) migrationEmbedder(target string) (face.Embedder, error) {
	if vision.Config == nil {
		return nil, fmt.Errorf("faces: vision configuration not available")
	}

	model := vision.Config.Model(vision.ModelTypeFace)
	if model == nil {
		return nil, fmt.Errorf("faces: face vision model not configured")
	}

	embedder := model.FaceModel()
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

// migrateFaceFile checkpoints all stale marker embeddings associated with a file.
func (w *Faces) migrateFaceFile(embedder face.Embedder, target, fileUID string) (migrated, skipped int, failed []string, detected bool, err error) {
	markers, err := query.FaceMigrationMarkers(fileUID)
	if err != nil {
		return 0, 0, nil, false, err
	}

	stale := make(entity.Markers, 0, len(markers))
	for i := range markers {
		// Rows written before the provenance column hold FaceNet vectors, so migrating a
		// legacy library to FaceNet has nothing to regenerate and must be a no-op.
		if face.ValidEmbeddings(markers[i].Embeddings(), embedder.Dims()) && face.ModelsComparable(markers[i].EmbedModel, target) {
			skipped++
		} else {
			stale = append(stale, markers[i])
		}
	}
	if len(stale) == 0 {
		return 0, skipped, nil, false, nil
	}

	file, err := query.FileByUID(fileUID)
	if err != nil {
		return 0, skipped, faceMigrationMarkerUIDs(stale), false, err
	}

	var generated map[string]face.Embeddings
	if embedder.Aligned() {
		detected = true
		generated, err = w.detectMigrationEmbeddings(embedder, file, markers, stale)
	} else {
		generated, err = w.cropMigrationEmbeddings(embedder, file, stale)
	}
	if err != nil {
		return 0, skipped, faceMigrationMarkerUIDs(stale), detected, err
	}

	for _, marker := range stale {
		if values, ok := generated[marker.MarkerUID]; ok && face.ValidEmbeddings(values, embedder.Dims()) {
			migrated++
		} else {
			failed = append(failed, marker.MarkerUID)
			delete(generated, marker.MarkerUID)
		}
	}

	if len(generated) > 0 {
		if err = query.SaveFaceMigrationEmbeddings(target, generated); err != nil {
			return 0, skipped, nil, detected, err
		}
	}

	return migrated, skipped, failed, detected, err
}

// faceMigrationMarkerUIDs returns the marker UIDs in their current order.
func faceMigrationMarkerUIDs(markers entity.Markers) []string {
	result := make([]string, 0, len(markers))
	for _, marker := range markers {
		result = append(result, marker.MarkerUID)
	}

	return result
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
			log.Warnf("faces: failed cropping marker %s (%s)", clean.Log(marker.MarkerUID), cropErr)
			continue
		}

		if embeddings := embedder.Run(img); face.ValidEmbeddings(embeddings, embedder.Dims()) {
			result[marker.MarkerUID] = embeddings
		}
	}

	return result, nil
}

// detectMigrationEmbeddings redetects a file and maps aligned embeddings to stored markers.
func (w *Faces) detectMigrationEmbeddings(embedder face.Embedder, file *entity.File, markers, stale entity.Markers) (result map[string]face.Embeddings, err error) {
	result = make(map[string]face.Embeddings, len(stale))
	thumbName, err := migrationDetectionThumb(w.conf, w.conf.ThumbCachePath(), file)
	if err != nil {
		return result, err
	}

	detected, err := face.Detect(thumbName, w.conf.FaceSize())
	if err != nil {
		return result, err
	}
	vision.GenerateEmbeddings(embedder, thumbName, detected, true)

	assignments := matchMigrationDetections(markers, detected)
	for _, marker := range stale {
		if detectedFace, ok := assignments[marker.MarkerUID]; ok && face.ValidEmbeddings(detectedFace.Embeddings, embedder.Dims()) {
			result[marker.MarkerUID] = detectedFace.Embeddings
		}
	}

	return result, nil
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

// matchMigrationDetections assigns each detected face to at most one stored marker.
func matchMigrationDetections(markers entity.Markers, detected face.Faces) map[string]face.Face {
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

	result := make(map[string]face.Face)
	used := make(map[int]bool)
	for _, pair := range pairs {
		if _, ok := result[pair.markerUID]; ok || used[pair.detected] {
			continue
		}
		result[pair.markerUID] = detected[pair.detected]
		used[pair.detected] = true
	}

	return result
}

// markerCropArea returns the normalized crop geometry stored on a marker.
func markerCropArea(marker entity.Marker) crop.Area {
	return crop.Area{Name: "face", X: marker.X, Y: marker.Y, W: marker.W, H: marker.H}
}

// buildFaceMigrationClusters creates one replacement cluster per identified subject,
// seeded from the assignments that agree with their own midpoint, and reports how many
// were left out.
//
// Subjects are rebuilt one at a time so that the embedding blobs of a whole library never
// have to be resident at once, matching the batching the re-embedding loop already uses.
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
			log.Warnf("faces: failed rebuilding cluster for subject %s", clean.Log(subjectUID))
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

// centroidSamples returns the markers that may define a subject's replacement cluster, and
// how many assignments were left out of it.
//
// An assignment the previous model got wrong would otherwise help shape the new centroid
// and widen it enough to attract further wrong faces. Outliers keep their assignment; they
// are only excluded from the samples, so nothing a person set is discarded here.
//
// A hand-set assignment is exempt from the ordinary outlier distance. Those are
// concentrated on small children, unusual angles and lens distortion - the cases embedding
// matching handles worst, which is why a person had to name them - so distance from the
// centroid measures the model rather than the assignment, and dropping the sample removes
// the width the cluster needs to reach that face. Every source is still held to the
// absolute bound, past which a vector says nothing about anyone.
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

	kept = make(entity.Markers, 0, len(group))

	for _, marker := range group {
		// Embeddings.Dist reports -1 when nothing is comparable, which would otherwise read
		// as the closest possible sample.
		switch d := marker.Embeddings().Dist(midpoint); {
		case d < 0 || d > face.AcceptDistMax:
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
