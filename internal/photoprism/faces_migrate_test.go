package photoprism

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type migrationTestEmbedder struct {
	name    string
	dims    int
	aligned bool
}

// ModelName returns the configured test model name.
func (e *migrationTestEmbedder) ModelName() face.ModelName { return e.name }

// Dims returns the configured test embedding length.
func (e *migrationTestEmbedder) Dims() int { return e.dims }

// CropSize returns the configured test crop size.
func (e *migrationTestEmbedder) CropSize() (int, int) { return 16, 16 }

// Aligned reports whether the test embedder requires landmark alignment.
func (e *migrationTestEmbedder) Aligned() bool { return e.aligned }

// Run returns a deterministic test embedding. It has a magnitude, because a vector without one is
// not a face and is refused where embeddings are written.
func (e *migrationTestEmbedder) Run(image.Image) face.Embeddings {
	result := make(face.Embedding, e.dims)

	if e.dims > 0 {
		result[0] = 1
	}

	return face.Embeddings{result}
}

// Close releases the test embedder.
func (e *migrationTestEmbedder) Close() error { return nil }

// TestFacesMigrateIncompleteError_Error pins that the message names what was actually lost. A
// marker an earlier detector placed and this one cannot find again is the ordinary outcome, and
// reporting the total as a fault tells an operator to fix something that is not broken.
func TestFacesMigrateIncompleteError_Error(t *testing.T) {
	t.Run("Clusterable", func(t *testing.T) {
		err := (&FacesMigrateIncompleteError{Failed: 900, Clusterable: 3}).Error()
		assert.Contains(t, err, "3 marker(s) that clear the clustering thresholds")
		assert.NotContains(t, err, "900")
	})
	t.Run("Unreadable", func(t *testing.T) {
		err := (&FacesMigrateIncompleteError{Failed: 5, Unreadable: 5}).Error()
		assert.Contains(t, err, "missing or unreadable")
	})
	t.Run("Both", func(t *testing.T) {
		err := (&FacesMigrateIncompleteError{Failed: 9, Clusterable: 4, Unreadable: 5}).Error()
		assert.Contains(t, err, "4 marker(s)")
		assert.Contains(t, err, "5 file(s)")
	})
}

func TestFacesMigrateAbortedError_Error(t *testing.T) {
	t.Run("NamesWhatWasCommitted", func(t *testing.T) {
		// Embeddings are checkpointed per file, so the message has to account for the
		// markers a refused run already regenerated and left unmatched.
		err := (&FacesMigrateAbortedError{Migrated: 7, Failed: 90, Reason: "storage is unreadable"}).Error()

		assert.Contains(t, err, "storage is unreadable")
		assert.Contains(t, err, "clusters were not replaced")
		assert.Contains(t, err, "7 regenerated marker(s) stay unmatched")
		assert.Contains(t, err, "--force")
	})
	t.Run("NothingRegenerated", func(t *testing.T) {
		err := (&FacesMigrateAbortedError{Migrated: 0, Failed: 12, Reason: "nothing could be re-embedded"}).Error()
		assert.Contains(t, err, "0 regenerated marker(s)")
	})
}

// installedOtherFaceModel returns an installed model that is not the configured one and can be
// migrated to in this environment, skipping the test when there is none.
func installedOtherFaceModel(t *testing.T, conf *config.Config) face.ModelName {
	t.Helper()

	configured := face.NormalizeModelName(conf.FaceModel())

	for _, name := range []face.ModelName{face.ModelFaceNet, face.ModelSFace} {
		m := face.FindEmbeddingModel(name)

		if name == configured || !m.Installed(conf.ModelsPath()) {
			continue
		} else if m.Aligned() && face.ActiveEngineName() != face.EngineONNX {
			continue
		}

		return name
	}

	t.Skip("faces: no other embedding model is installed")

	return ""
}

// otherFaceModel returns a registered embedding model that is not the specified one, so
// tests can exercise the cross-model guards without assuming which model a test library
// resolves to.
func otherFaceModel(t *testing.T, configured face.ModelName) face.ModelName {
	t.Helper()

	configured = face.NormalizeModelName(configured)

	for _, name := range face.EmbeddingModelNames() {
		if name != configured {
			return name
		}
	}

	t.Fatalf("no embedding model other than %s is registered", configured)

	return ""
}

func TestFaces_PlanMigration(t *testing.T) {
	w := NewFaces(config.TestConfig())

	t.Run("ConfiguredModel", func(t *testing.T) {
		// An empty target means "the model this instance is configured for", which is
		// what makes the migration converge instead of creating a third vector space.
		result, err := w.PlanMigration("")
		require.NoError(t, err)
		assert.Equal(t, face.NormalizeModelName(w.conf.FaceModel()), result.Target)
		assert.Positive(t, result.Markers.Total)
	})
	t.Run("MismatchedTarget", func(t *testing.T) {
		_, err := w.PlanMigration(otherFaceModel(t, w.conf.FaceModel()))
		require.Error(t, err)
	})
	t.Run("DisabledTarget", func(t *testing.T) {
		_, err := w.PlanMigration(face.ModelNone)
		require.Error(t, err)
	})
	t.Run("MissingConfig", func(t *testing.T) {
		_, err := (&Faces{}).PlanMigration("")
		require.Error(t, err)
	})
	t.Run("OriginalsUnavailable", func(t *testing.T) {
		// The counting queries cannot see the filesystem, so the plan carries this separately.
		// The seeded test library keeps its media elsewhere, which is the unreadable case.
		result, err := w.PlanMigration("")
		require.NoError(t, err)
		assert.True(t, result.OriginalsUnavailable)

		name := filepath.Join(w.conf.OriginalsPath(), "plan-migration-test.jpg")
		require.NoError(t, os.WriteFile(name, []byte("test"), fs.ModeFile))
		t.Cleanup(func() { _ = os.Remove(name) })

		result, err = w.PlanMigration("")
		require.NoError(t, err)
		assert.False(t, result.OriginalsUnavailable)
	})
}

func TestOriginalsUnavailable(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		assert.True(t, originalsUnavailable(filepath.Join(t.TempDir(), "not-mounted")))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, originalsUnavailable(t.TempDir()))
	})
	t.Run("NoPath", func(t *testing.T) {
		assert.True(t, originalsUnavailable(""))
	})
	t.Run("Populated", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("test"), fs.ModeFile))
		assert.False(t, originalsUnavailable(dir))
	})
}

func TestFaces_MigrateDryRun(t *testing.T) {
	w := NewFaces(config.TestConfig())
	configured := face.NormalizeModelName(w.conf.FaceModel())

	result, err := w.Migrate(context.Background(), FacesMigrateOptions{DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, configured, result.Target)
	assert.Equal(t, 0, result.Migrated)

	// A dry run reports the scope and changes nothing, including what is configured.
	assert.Equal(t, configured, face.NormalizeModelName(w.conf.FaceModel()))

	_, err = w.Migrate(context.Background(), FacesMigrateOptions{Target: face.ModelNone, DryRun: true})
	require.Error(t, err)
}

func TestFaces_migrationEmbedder(t *testing.T) {
	w := NewFaces(config.TestConfig())
	configured := face.NormalizeModelName(w.conf.FaceModel())

	embedder, err := w.migrationEmbedder(configured)
	require.NoError(t, err)
	assert.Equal(t, configured, embedder.ModelName())

	_, err = w.migrationEmbedder(otherFaceModel(t, w.conf.FaceModel()))
	require.Error(t, err)
}

func TestUnrecordedFaceModel(t *testing.T) {
	t.Run("Saved", func(t *testing.T) {
		assert.NoError(t, unrecordedFaceModel(nil, face.ModelSFace, nil))
	})
	t.Run("SavedWithAnotherError", func(t *testing.T) {
		cause := errors.New("clustering failed")

		assert.Equal(t, cause, unrecordedFaceModel(nil, face.ModelSFace, cause))
	})
	t.Run("Unsaved", func(t *testing.T) {
		err := unrecordedFaceModel(errors.New("permission denied"), face.ModelSFace, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "set FaceModel: sface in options.yml")
	})
	t.Run("UnsavedKeepsWhatFailedAfterIt", func(t *testing.T) {
		// A read-only volume fails the write and then the clustering, so reporting only the
		// second would hide the one that decides whether the library is matched at all.
		cause := errors.New("clustering failed")
		err := unrecordedFaceModel(errors.New("permission denied"), face.ModelSFace, cause)

		require.Error(t, err)
		assert.ErrorIs(t, err, cause)

		var settingErr *FacesMigrateSettingError
		require.True(t, errors.As(err, &settingErr))
		assert.Equal(t, face.ModelSFace, settingErr.Target)
	})
}

func TestFaces_restoreEmbedder(t *testing.T) {
	t.Run("AFailedTargetLeavesTheModelInPlace", func(t *testing.T) {
		// The embedder is the one this process runs on, so a target that cannot be loaded
		// must not leave the instance without the model it was running on.
		w := NewFaces(config.TestConfig())
		configured := face.NormalizeModelName(w.conf.FaceModel())

		_, err := w.migrationEmbedder(otherFaceModel(t, configured))

		require.Error(t, err)
		assert.Equal(t, configured, face.ConfiguredModel())
	})
	t.Run("RestoresTheConfiguredModel", func(t *testing.T) {
		w := NewFaces(config.TestConfig())
		configured := face.NormalizeModelName(w.conf.FaceModel())

		require.NoError(t, w.conf.ConfigureFaceEmbedder(face.ModelNone))
		require.Equal(t, face.ModelNone, face.ConfiguredModel())

		w.restoreEmbedder()

		assert.Equal(t, configured, face.ConfiguredModel())
	})
}

func TestFaces_migrateFaceFile(t *testing.T) {
	embedder := &migrationTestEmbedder{name: face.ModelFaceNet, dims: 512}
	w := NewFaces(config.TestConfig())

	empty, err := w.migrateFaceFile(embedder, face.ModelFaceNet, "missing-file")
	require.NoError(t, err)
	assert.Zero(t, empty.Migrated)
	assert.Zero(t, empty.Skipped)
	assert.Zero(t, empty.Retained)
	assert.Empty(t, empty.Failed)
	assert.False(t, empty.Detected)

	_, err = w.migrateFaceFile(embedder, face.ModelFaceNet, "")
	require.Error(t, err)

	marker := &entity.Marker{
		MarkerUID:  rnd.GenerateUID('m'),
		FileUID:    rnd.GenerateUID('f'),
		MarkerType: entity.MarkerFace,
		W:          0.1,
		H:          0.1,
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	// A file the library removed is absent from the scoped lookup, and the plan already
	// counted its markers as unreadable, so the error has to name that rather than report a
	// bare "record not found" an operator would read as a database fault.
	deleted, err := w.migrateFaceFile(embedder, face.ModelFaceNet, marker.FileUID)
	require.Error(t, err)
	assert.Equal(t, "file was deleted", err.Error())
	assert.Equal(t, []string{marker.MarkerUID}, deleted.Failed)
	assert.Zero(t, deleted.Retained)

	t.Run("CropPathKeepsTheRecordedDetector", func(t *testing.T) {
		// Re-cropping reads the stored geometry, so the detector that produced it is still
		// the one that produced the crop and must survive the checkpoint.
		hash := "c0ffee00000000000000000000000000000000a5"
		photo := entity.Photo{PhotoUID: rnd.GenerateUID('p'), PhotoName: "migrate-crop", PhotoType: entity.MediaImage}
		require.NoError(t, photo.Save())
		file := &entity.File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     rnd.GenerateUID('f'),
			FileHash:    hash,
			FileName:    "migrate-crop/a.jpg",
			FileRoot:    entity.RootOriginals,
			FilePrimary: true,
			FileType:    "jpg",
		}
		require.NoError(t, file.Create())

		stale := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        file.FileUID,
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			EmbedModel:     face.ModelSFace,
			DetectModel:    "legacy-detector",
			EmbeddingsJSON: face.Embeddings{face.Embedding{1, 0, 0, 0}}.JSON(),
			W:              1,
			H:              1,
		}
		require.NoError(t, entity.Db().Create(stale).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(stale) })

		cropConf := config.NewMinimalTestConfig(t.TempDir())
		require.NoError(t, cropConf.CreateDirectories())
		cropWorker := NewFaces(cropConf)

		thumbName, thumbErr := thumb.Sizes[thumb.Fit720].FileName(hash, cropConf.ThumbCachePath())
		require.NoError(t, thumbErr)
		require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 64, 64)), thumbName))

		cropEmbedder := &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4}
		cropped, cropErr := cropWorker.migrateFaceFile(cropEmbedder, face.ModelFaceNet, file.FileUID)

		require.NoError(t, cropErr)
		assert.False(t, cropped.Detected)
		assert.Empty(t, cropped.Failed)
		assert.Zero(t, cropped.Retained)
		require.Equal(t, 1, cropped.Migrated)

		saved, savedErr := query.MarkerByUID(stale.MarkerUID)
		require.NoError(t, savedErr)
		assert.Equal(t, face.ModelFaceNet, saved.EmbedModel)
		// Deliberately not the active engine: a value equal to it could not tell a preserved
		// detector from one the crop path wrongly re-attributed.
		assert.NotEqual(t, face.ActiveEngineName(), "legacy-detector")
		assert.Equal(t, "legacy-detector", saved.DetectModel)
	})
}

func TestUnresolvedMigrationMarkers(t *testing.T) {
	markers := entity.Markers{{MarkerUID: "m1"}, {MarkerUID: "m2"}, {MarkerUID: "m3"}}
	t.Run("NothingToRecrop", func(t *testing.T) {
		failed, counts := unresolvedMigrationMarkers(markers, nil)
		assert.Equal(t, []string{"m1", "m2", "m3"}, failed)
		assert.Zero(t, counts.Retained)
		// The fixtures carry neither a size nor a score, so none of them could have clustered.
		assert.Zero(t, counts.Clusterable)
		assert.Zero(t, counts.Named)
		assert.Zero(t, counts.Manual)
	})
	t.Run("RecropIsRetained", func(t *testing.T) {
		// A marker that only needed a new crop keeps the vector it holds, so it must not reach
		// the failed list: that list is what the finalize blanks.
		failed, counts := unresolvedMigrationMarkers(markers, map[string]bool{"m2": true})
		assert.Equal(t, []string{"m1", "m3"}, failed)
		assert.Equal(t, 1, counts.Retained)
	})
	t.Run("EveryMarkerIsRetained", func(t *testing.T) {
		failed, counts := unresolvedMigrationMarkers(markers, map[string]bool{"m1": true, "m2": true, "m3": true})
		assert.Empty(t, failed)
		assert.Equal(t, 3, counts.Retained)
		assert.Zero(t, counts.Clusterable)
	})
	t.Run("NoMarkers", func(t *testing.T) {
		failed, counts := unresolvedMigrationMarkers(nil, map[string]bool{"m1": true})
		assert.Empty(t, failed)
		assert.Zero(t, counts.Retained)
		assert.Zero(t, counts.Clusterable)
	})
	t.Run("CountsOnlyWhatCouldHaveClustered", func(t *testing.T) {
		// The finalize guard measures real loss, so a marker under either clustering bar must
		// not count toward it however many of them a detector change fails to re-find.
		mixed := entity.Markers{
			{MarkerUID: "big", Size: face.ClusterSizeThreshold, Score: 100},
			{MarkerUID: "small", Size: face.ClusterSizeThreshold - 1, Score: 100},
			{MarkerUID: "weak", Size: face.ClusterSizeThreshold, Score: 1},
		}

		failed, counts := unresolvedMigrationMarkers(mixed, nil)

		assert.Len(t, failed, 3)
		assert.Equal(t, 1, counts.Clusterable)
	})
	t.Run("SplitsTheLossByWhatMakesAMarkerWorthKeeping", func(t *testing.T) {
		// A marker nobody named is attrition on a library an unreliable detector indexed; one a
		// person named is the cost, and a hand-drawn one is the least recoverable. A bare failure
		// count cannot tell them apart, which is what makes a migration floor unjudgeable.
		mixed := entity.Markers{
			{MarkerUID: "anon"},
			{MarkerUID: "named", SubjUID: "js6sg6b1qekk9jx8"},
			{MarkerUID: "drawn", SubjUID: "js6sg6b1qekk9jx9", SubjSrc: entity.SrcManual},
			// ClearSubject blanks subj_uid and records the source of the clearing, so this is what
			// a marker somebody named and later un-named looks like. It is effort spent either way.
			{MarkerUID: "unnamed", SubjSrc: entity.SrcManual},
		}

		failed, counts := unresolvedMigrationMarkers(mixed, nil)

		assert.Len(t, failed, 4)
		assert.Equal(t, 2, counts.Named, "every assigned marker counts, however it was assigned")
		assert.Equal(t, 2, counts.Manual, "including one with no subject, so the counts overlap rather than nest")
		// Three of the four markers are named or manual, yet the counts sum to four. That gap is
		// the overlap, and reporting one as a subset of the other is what it would hide.
		assert.Equal(t, 4, counts.Named+counts.Manual)
	})
}

func TestStaleMigrationMarkers(t *testing.T) {
	vector := face.Embeddings{face.Embedding{1, 0, 0, 0}}.JSON()
	detector := face.ActiveDetector()
	require.NotEmpty(t, detector)
	require.NotEqual(t, face.DetectorNone, detector)

	current := entity.Marker{MarkerUID: "mcurrent", EmbedModel: face.ModelFaceNet, DetectModel: detector, EmbeddingsJSON: vector}
	other := entity.Marker{MarkerUID: "mother", EmbedModel: face.ModelFaceNet, DetectModel: "some-other-detector", EmbeddingsJSON: vector}
	blank := entity.Marker{MarkerUID: "mblank", EmbedModel: face.ModelFaceNet, EmbeddingsJSON: vector}
	wrongModel := entity.Marker{MarkerUID: "mmodel", EmbedModel: face.ModelSFace, DetectModel: detector, EmbeddingsJSON: vector}

	markers := entity.Markers{current, other, blank, wrongModel}

	t.Run("Aligned", func(t *testing.T) {
		// An aligned embedder consumes the landmarks, so a crop another detector placed is
		// stale even though its vector is the target's, and is recorded as re-croppable.
		stale, recrop := staleMigrationMarkers(markers, &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4, aligned: true}, face.ModelFaceNet)
		assert.Equal(t, []string{"mother", "mblank", "mmodel"}, markerUIDsOf(stale))
		assert.Equal(t, map[string]bool{"mother": true, "mblank": true}, recrop)
	})
	t.Run("NotAligned", func(t *testing.T) {
		// A crop-based embedder reads the stored geometry, so the detector that placed it
		// changes nothing and only the incomparable vector is stale.
		stale, recrop := staleMigrationMarkers(markers, &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4}, face.ModelFaceNet)
		assert.Equal(t, []string{"mmodel"}, markerUIDsOf(stale))
		assert.Empty(t, recrop)
	})
	t.Run("UnreadableVector", func(t *testing.T) {
		// A vector of the wrong width cannot be compared at all, so it is stale outright and
		// must never be retained: there is nothing usable to keep.
		stale, recrop := staleMigrationMarkers(entity.Markers{current}, &migrationTestEmbedder{name: face.ModelFaceNet, dims: 512, aligned: true}, face.ModelFaceNet)
		assert.Equal(t, []string{"mcurrent"}, markerUIDsOf(stale))
		assert.Empty(t, recrop)
	})
	t.Run("NoMarkers", func(t *testing.T) {
		stale, recrop := staleMigrationMarkers(nil, &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4, aligned: true}, face.ModelFaceNet)
		assert.Empty(t, stale)
		assert.Empty(t, recrop)
	})
}

// markerUIDsOf returns the marker UIDs in their current order, so a test can assert on the
// selection a helper made rather than on the marker rows it copied.
func markerUIDsOf(markers entity.Markers) []string {
	result := make([]string, 0, len(markers))
	for _, marker := range markers {
		result = append(result, marker.MarkerUID)
	}

	return result
}

func TestFaces_cropMigrationEmbeddings(t *testing.T) {
	embedder := &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4}
	w := NewFaces(config.TestConfig())
	file := &entity.File{FileHash: "0123456789012345678901234567890123456789", FileName: "missing.jpg", FileRoot: entity.RootOriginals}
	markers := entity.Markers{{MarkerUID: "m1", MarkerType: entity.MarkerFace, W: 0.5, H: 0.5}}

	result, _, err := w.cropMigrationEmbeddings(embedder, file, markers)
	require.NoError(t, err)
	assert.Empty(t, result)

	_, _, err = w.cropMigrationEmbeddings(embedder, nil, markers)
	require.Error(t, err)

	successConf := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, successConf.CreateDirectories())
	successWorker := NewFaces(successConf)
	hash := "0123456789012345678901234567890123456789"
	thumbName, err := thumb.Sizes[thumb.Fit720].FileName(hash, successConf.ThumbCachePath())
	require.NoError(t, err)
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 64, 64)), thumbName))
	result, _, err = successWorker.cropMigrationEmbeddings(
		embedder,
		&entity.File{FileHash: hash},
		entity.Markers{{MarkerUID: "m2", W: 1, H: 1}},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestFaces_detectMigrationEmbeddings(t *testing.T) {
	embedder := &migrationTestEmbedder{name: face.ModelSFace, dims: 4, aligned: true}
	w := NewFaces(config.TestConfig())

	result, landmarks, detectModel, err := w.detectMigrationEmbeddings(embedder, nil, nil, nil)
	require.Error(t, err)
	assert.Empty(t, result)
	assert.Empty(t, landmarks)
	assert.Empty(t, detectModel, "a run that could not detect names no detector")
}

// TestFaces_detectMigrationEmbeddings_Landmarks pins the producer half of the provenance pair:
// the landmarks a re-detection placed have to travel with the vector it produced, or the recorded
// detector attests a crop while the stored landmarks are an earlier detector's.
func TestFaces_detectMigrationEmbeddings_Landmarks(t *testing.T) {
	prev := face.UseEngine(nil)
	t.Cleanup(func() {
		if current := face.UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	c := config.TestConfig()
	detectorPath := face.DefaultDetector().Path(c.ModelsPath())

	if _, err := os.Stat(detectorPath); err != nil {
		t.Skipf("faces: skipping, %s is not available", filepath.Base(detectorPath))
	}

	require.NoError(t, face.ConfigureEngine(face.EngineSettings{
		Name: face.EngineONNX,
		ONNX: face.ONNXOptions{ModelPath: detectorPath, Threads: 1},
	}))

	// The detection thumbnail is what the migration re-detects, so a real face has to be cached
	// under the file's hash for the run to have anything to map.
	source, _, err := fs.DecodeImageFile(filepath.Join("..", "ai", "face", "testdata", "1.jpg"))
	require.NoError(t, err)

	hash := "1111111111111111111111111111111111111111"
	thumbName, err := thumb.Sizes[thumb.Fit720].FileName(hash, c.ThumbCachePath())
	require.NoError(t, err)
	require.NoError(t, thumb.Save(source, thumbName))

	detected, err := face.DetectWithRetry(thumbName, c.FaceSize(), c.FaceSizeRetry())
	require.NoError(t, err)
	require.NotEmpty(t, detected, "the test image must yield a face to map")

	file := &entity.File{FileUID: rnd.GenerateUID('f'), FileHash: hash}
	marker := entity.NewFaceMarker(detected[0], *file, "")
	require.NotNil(t, marker)
	marker.MarkerUID = rnd.GenerateUID('m')

	markers := entity.Markers{*marker}
	embedder := &migrationTestEmbedder{name: face.ModelSFace, dims: 4, aligned: true}

	w := NewFaces(c)
	result, details, detectModel, err := w.detectMigrationEmbeddings(embedder, file, markers, markers)

	require.NoError(t, err)
	require.Contains(t, result, marker.MarkerUID)
	assert.Equal(t, face.DefaultDetector().Name, detectModel)
	require.Contains(t, details, marker.MarkerUID, "what the detection recorded must travel with the vector")

	detection := details[marker.MarkerUID]
	assert.True(t, json.Valid(detection.Landmarks))
	assert.Contains(t, string(detection.Landmarks), "eye_l")

	// The clustering bars are looked up by detect_model, so the score has to come from the
	// detection the provenance names rather than from whatever placed the marker originally.
	assert.Positive(t, detection.Score)
	assert.Positive(t, detection.Size)
}

func TestMigrationDetectionThumb(t *testing.T) {
	c := config.TestConfig()
	thumbPath := t.TempDir()
	hash := "0123456789012345678901234567890123456789"
	name, err := thumb.Sizes[thumb.Fit720].FileName(hash, thumbPath)
	require.NoError(t, err)
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 8, 8)), name))

	result, err := migrationDetectionThumb(c, thumbPath, &entity.File{FileHash: hash})
	require.NoError(t, err)
	assert.Equal(t, name, result)

	_, err = migrationDetectionThumb(c, thumbPath, nil)
	require.Error(t, err)
}

func TestMatchMigrationDetections(t *testing.T) {
	markers := entity.Markers{
		{MarkerUID: "m1", X: 0.1, Y: 0.1, W: 0.2, H: 0.2},
		{MarkerUID: "m2", X: 0.6, Y: 0.6, W: 0.2, H: 0.2},
	}
	detected := face.Faces{
		{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)},
		{Rows: 100, Cols: 100, Area: face.NewArea("face", 70, 70, 20)},
	}

	result := matchMigrationDetections(markers, detected)
	require.Len(t, result, 2)
	assert.Equal(t, 0, result["m1"])
	assert.Equal(t, 1, result["m2"])
	assert.Empty(t, matchMigrationDetections(nil, detected))
}

// TestAssignedMigrationDetections pins that only the detections a stale marker claims are handed
// on to be embedded. Migration detects at the smallest size the detectors are trained for, so a
// file yields detections no marker accounts for, and inferring those would cost the run their
// time and the real crops their rendition.
func TestAssignedMigrationDetections(t *testing.T) {
	markers := entity.Markers{
		{MarkerUID: "m1", X: 0.1, Y: 0.1, W: 0.2, H: 0.2},
		{MarkerUID: "m2", X: 0.6, Y: 0.6, W: 0.2, H: 0.2},
	}
	detected := face.Faces{
		{Rows: 100, Cols: 100, Area: face.NewArea("face", 20, 20, 20)},
		{Rows: 100, Cols: 100, Area: face.NewArea("face", 70, 70, 20)},
		{Rows: 100, Cols: 100, Area: face.NewArea("face", 5, 90, 4)},
	}

	t.Run("Success", func(t *testing.T) {
		// Only m2 needs work, so the detection m1 accounts for must not be embedded, and must
		// still be unavailable to m2.
		assigned, order := assignedMigrationDetections(markers, entity.Markers{markers[1]}, detected)

		require.Len(t, assigned, 1)
		assert.Equal(t, detected[1].Area, assigned[0].Area)
		assert.Equal(t, map[string]int{"m2": 0}, order)
	})
	t.Run("ContestedDetection", func(t *testing.T) {
		// One detection, two markers overlapping it, and the one that needs no work wins it.
		// Matching only the stale markers would hand it to m2 instead, which is the regression
		// passing every marker exists to prevent - and which a fixture where each marker has
		// its own detection cannot see.
		contested := entity.Markers{
			{MarkerUID: "m1", X: 0.40, Y: 0.40, W: 0.10, H: 0.10},
			{MarkerUID: "m2", X: 0.42, Y: 0.42, W: 0.10, H: 0.10},
		}
		one := face.Faces{{Rows: 100, Cols: 100, Score: 80, Area: face.NewArea("face", 45, 45, 10)}}

		require.Equal(t, map[string]int{"m1": 0}, matchMigrationDetections(contested, one),
			"the detection must belong to m1")

		assigned, order := assignedMigrationDetections(contested, entity.Markers{contested[1]}, one)

		assert.Empty(t, assigned, "m2 must not be given a detection m1 accounts for")
		assert.Empty(t, order)
	})
	t.Run("EveryMarkerIsStale", func(t *testing.T) {
		assigned, order := assignedMigrationDetections(markers, markers, detected)

		require.Len(t, assigned, 2, "the unclaimed detection must be left out")
		assert.Equal(t, detected[0].Area, assigned[order["m1"]].Area)
		assert.Equal(t, detected[1].Area, assigned[order["m2"]].Area)
	})
	t.Run("NoDetections", func(t *testing.T) {
		assigned, order := assignedMigrationDetections(markers, markers, nil)

		assert.Empty(t, assigned)
		assert.Empty(t, order)
	})
}

// TestOversizedMigrationDetection pins the bound that keeps a containing box from claiming a
// marker. OverlapPercent divides by the marker's own surface, so a detection that merely
// contains it scores a perfect 100 while the correctly fitting one scores less.
func TestOversizedMigrationDetection(t *testing.T) {
	marker := crop.Area{Name: "face", X: 0.4, Y: 0.4, W: 0.1, H: 0.1}

	t.Run("SameSize", func(t *testing.T) {
		assert.False(t, oversizedMigrationDetection(crop.Area{X: 0.4, Y: 0.4, W: 0.1, H: 0.1}, marker))
	})
	t.Run("SlightlyLarger", func(t *testing.T) {
		// Ordinary detector-to-detector drift must still match.
		assert.False(t, oversizedMigrationDetection(crop.Area{X: 0.38, Y: 0.38, W: 0.14, H: 0.14}, marker))
	})
	t.Run("HeadAndShoulders", func(t *testing.T) {
		assert.True(t, oversizedMigrationDetection(crop.Area{X: 0.3, Y: 0.3, W: 0.3, H: 0.3}, marker))
	})
	t.Run("EmptyMarker", func(t *testing.T) {
		// Nothing to compare against, so nothing is rejected on size.
		assert.False(t, oversizedMigrationDetection(crop.Area{X: 0.3, Y: 0.3, W: 0.3, H: 0.3}, crop.Area{}))
	})
}

// TestMatchMigrationDetectionsPrefersConfidence pins that a tie on overlap is broken by detection
// score rather than by the order the detector emitted them. Containment scores 100, so the
// migration's lower floors put several candidates at the top of the list.
func TestMatchMigrationDetectionsPrefersConfidence(t *testing.T) {
	markers := entity.Markers{{MarkerUID: "m1", X: 0.4, Y: 0.4, W: 0.1, H: 0.1}}
	detected := face.Faces{
		{Rows: 100, Cols: 100, Score: 12, Area: face.NewArea("face", 45, 45, 10)},
		{Rows: 100, Cols: 100, Score: 92, Area: face.NewArea("face", 45, 45, 10)},
	}

	result := matchMigrationDetections(markers, detected)

	require.Contains(t, result, "m1")
	assert.Equal(t, 1, result["m1"], "the more confident detection must claim the marker")
}

func TestClusterableMarkers(t *testing.T) {
	markers := entity.Markers{
		{MarkerUID: "big", Size: face.ClusterSizeThreshold, Score: 100},
		{MarkerUID: "small", Size: face.ClusterSizeThreshold - 1, Score: 100},
	}

	assert.Equal(t, 1, clusterableMarkers(markers))
	assert.Zero(t, clusterableMarkers(nil))
}

func TestRetainedMigrationMarkers(t *testing.T) {
	markers := entity.Markers{{MarkerUID: "m1"}, {MarkerUID: "m2"}}

	assert.Len(t, retainedMigrationMarkers(markers, map[string]bool{"m2": true}), 1)
	assert.Empty(t, retainedMigrationMarkers(markers, nil))
}

func TestMarkerCropArea(t *testing.T) {
	result := markerCropArea(entity.Marker{X: 0.1, Y: 0.2, W: 0.3, H: 0.4})
	assert.Equal(t, float32(0.1), result.X)
	assert.Equal(t, float32(0.4), result.H)
	assert.Zero(t, markerCropArea(entity.Marker{}).W)
}

func TestValidMigrationEmbeddingsUsage(t *testing.T) {
	assert.True(t, face.ValidEmbeddings(face.Embeddings{{0.1, 0.2}}, 2))
	assert.False(t, face.ValidEmbeddings(nil, 2))
	assert.False(t, face.ValidEmbeddings(face.Embeddings{{0.1}}, 2))
	assert.False(t, face.ValidEmbeddings(face.Embeddings{{0.1, math.NaN()}}, 2))
}

func TestBuildFaceMigrationClusters(t *testing.T) {
	// RandomEmbedding follows the configured model, so the marker has to claim that same
	// model or its vector is the wrong length for the space the clusters are rebuilt in.
	target := face.ConfiguredModel()
	subjectUID := rnd.GenerateUID('j')

	newMarker := func(subjSrc string) *entity.Marker {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			SubjUID:        subjectUID,
			SubjSrc:        subjSrc,
			Size:           face.ClusterSizeThreshold + 10,
			Score:          face.ClusterScore("") + 10,
			EmbedModel:     target,
			EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
			W:              0.1,
			H:              0.1,
		}
		require.NoError(t, entity.Db().Create(m).Error)
		t.Cleanup(func() { entity.UnscopedDb().Delete(m) })

		return m
	}

	manual := newMarker(entity.SrcManual)
	automatic := newMarker(entity.SrcAuto)

	result, rebuilt, excluded, err := buildFaceMigrationClusters(target)
	require.NoError(t, err)
	assert.Equal(t, rebuilt, len(result))
	assert.GreaterOrEqual(t, excluded, 0)

	var cluster *query.FaceMigrationCluster
	for i := range result {
		if result[i].Face.SubjUID == subjectUID {
			cluster = &result[i]
		}
	}
	require.NotNil(t, cluster, "the subject must get a replacement cluster")

	// Both assignments have to seed the cluster, or it comes out too narrow to match with.
	assert.Contains(t, cluster.MarkerDistances, manual.MarkerUID)
	assert.Contains(t, cluster.MarkerDistances, automatic.MarkerUID)
	assert.Equal(t, 2, cluster.Face.Samples)

	_, _, _, err = buildFaceMigrationClusters("unknown")
	require.Error(t, err)
}

func TestFaces_migrationTarget(t *testing.T) {
	w := NewFaces(config.TestConfig())
	configured := face.NormalizeModelName(w.conf.FaceModel())
	t.Run("EmptyResolvesToConfigured", func(t *testing.T) {
		target, err := w.migrationTarget("")
		require.NoError(t, err)
		assert.Equal(t, configured, target)
	})
	t.Run("OtherModelIsAllowed", func(t *testing.T) {
		// The migration is how the model is changed, so a target that differs from the
		// configured model is its normal input rather than a refusal.
		other := installedOtherFaceModel(t, w.conf)

		target, err := w.migrationTarget(other)
		require.NoError(t, err)
		assert.Equal(t, other, target)
	})
	t.Run("GatedModelWithoutAcceptance", func(t *testing.T) {
		t.Setenv(face.LicenseAcceptanceVar, "")

		_, err := w.migrationTarget(face.ModelArcFaceR50)
		require.Error(t, err)
		assert.Contains(t, err.Error(), face.LicenseAcceptanceVar)
	})
	t.Run("Disabled", func(t *testing.T) {
		_, err := w.migrationTarget(face.ModelNone)
		require.Error(t, err)
	})
	t.Run("MissingConfig", func(t *testing.T) {
		_, err := (&Faces{}).migrationTarget("")
		require.Error(t, err)
	})
}

func TestFaces_migratePlan(t *testing.T) {
	w := NewFaces(config.TestConfig())
	configured := face.NormalizeModelName(w.conf.FaceModel())
	t.Run("WithoutPlan", func(t *testing.T) {
		plan, err := w.migratePlan(FacesMigrateOptions{})
		require.NoError(t, err)
		assert.Equal(t, configured, plan.Target)
		assert.Positive(t, plan.Markers.Total)
	})
	t.Run("ReusesSuppliedPlan", func(t *testing.T) {
		// Counts the caller already gathered are trusted; the target never is.
		supplied := FacesMigratePlan{Target: configured, Subjects: 7, AssignedMarkers: 42}
		plan, err := w.migratePlan(FacesMigrateOptions{Target: configured, Plan: &supplied})
		require.NoError(t, err)
		assert.Equal(t, 7, plan.Subjects)
		assert.Equal(t, 42, plan.AssignedMarkers)
		assert.Zero(t, plan.Markers.Total, "a supplied plan must not be re-counted")
	})
	t.Run("RejectsForeignTarget", func(t *testing.T) {
		supplied := FacesMigratePlan{Target: otherFaceModel(t, w.conf.FaceModel())}
		_, err := w.migratePlan(FacesMigrateOptions{Plan: &supplied})
		require.Error(t, err)
	})
}

func TestManualSubjectAssignment(t *testing.T) {
	t.Run("Manual", func(t *testing.T) {
		assert.True(t, manualSubjectAssignment(entity.SrcManual))
	})
	t.Run("Automatic", func(t *testing.T) {
		// What the previous model decided is what the outlier rule is looking for.
		assert.False(t, manualSubjectAssignment(entity.SrcAuto))
	})
	t.Run("Xmp", func(t *testing.T) {
		// XMP names never seed a shared face either, so they follow the automatic rule.
		assert.False(t, manualSubjectAssignment(entity.SrcXmp))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, manualSubjectAssignment(""))
	})
	t.Run("Image", func(t *testing.T) {
		assert.True(t, manualSubjectAssignment(entity.SrcImage))
	})
}

func TestCentroidSamples(t *testing.T) {
	registered := face.FindEmbeddingModel(face.ModelSFace)
	require.NotNil(t, registered)

	// A tight group plus one vector pointing elsewhere, which is what an assignment the
	// previous model got wrong looks like once it has been re-embedded.
	marker := func(v []float32) entity.Marker {
		m := entity.Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: entity.MarkerFace}
		m.SetEmbeddings(face.NewEmbeddings([][]float32{v}), face.ModelSFace, face.EngineONNX)

		return m
	}
	near := func(i int) []float32 {
		v := make([]float32, registered.Dims)
		v[0] = 1
		v[1+i] = 0.05

		return v
	}
	// The two bands the exemption is defined by, taken from the model rather than written
	// down, so a recalibration cannot silently move a fixture out of the band it names.
	maxAccept := min(registered.ClusterRadius+registered.MatchDist, face.AcceptDistMax)

	// atGroupDist returns a vector whose distance to the midpoint of {near(0..2), itself}
	// is target. The outlier pulls that midpoint toward itself, so the angle that produces
	// a given distance is searched for rather than derived.
	atGroupDist := func(target float64) []float32 {
		best := make([]float32, registered.Dims)
		bestErr := math.Inf(1)

		for step := 0; step <= 400; step++ {
			theta := float64(step) / 400 * math.Pi
			v := make([]float32, registered.Dims)
			v[0] = float32(math.Cos(theta))
			v[registered.Dims-1] = float32(math.Sin(theta))

			group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), marker(v)}

			embeddings := make(face.Embeddings, 0, len(group))
			for i := range group {
				embeddings = append(embeddings, group[i].Embeddings()...)
			}

			midpoint, _, count := face.EmbeddingsMidpoint(embeddings)
			if count == 0 {
				continue
			}

			if e := math.Abs(group[3].Embeddings().Dist(midpoint) - target); e < bestErr {
				bestErr, best = e, v
			}
		}

		require.Less(t, bestErr, 0.02, "fixture must be able to reach %.2f from the midpoint", target)

		return best
	}

	// Beyond the outlier distance but inside the widest distance the resulting cluster can
	// accept, which is the band where the manual exemption applies.
	far := func() []float32 {
		return atGroupDist((registered.ClusterDist + maxAccept) / 2)
	}
	// Past that bound, where no cluster of this model would re-accept it.
	beyondAcceptance := func() []float32 {
		return atGroupDist(min(maxAccept+0.15, 1.9))
	}

	t.Run("DropsOutlier", func(t *testing.T) {
		outlier := marker(far())
		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), outlier}

		kept, dropped := centroidSamples(group, registered, "jsubject00000001")
		assert.Equal(t, 1, dropped)
		assert.Len(t, kept, 3)
		for _, m := range kept {
			assert.NotEqual(t, outlier.MarkerUID, m.MarkerUID, "the outlier must not define the centroid")
		}
	})
	t.Run("KeepsAgreeingGroup", func(t *testing.T) {
		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2))}
		kept, dropped := centroidSamples(group, registered, "jsubject00000002")
		assert.Len(t, kept, 3)
		assert.Zero(t, dropped)
	})
	t.Run("TooFewToJudge", func(t *testing.T) {
		// Two samples have no majority to appeal to, so neither is treated as the outlier.
		group := entity.Markers{marker(near(0)), marker(far())}
		kept, _ := centroidSamples(group, registered, "jsubject00000003")
		assert.Len(t, kept, 2)
	})
	t.Run("AllScattered", func(t *testing.T) {
		// Nothing agrees with the midpoint, so there is no signal to separate from noise.
		group := entity.Markers{}
		for i := range 4 {
			v := make([]float32, registered.Dims)
			v[i*8] = 1
			group = append(group, marker(v))
		}
		kept, _ := centroidSamples(group, registered, "jsubject00000004")
		assert.Len(t, kept, 4)
	})
	t.Run("KeepsManualOutlier", func(t *testing.T) {
		// Hand-named faces are concentrated on what embedding matching handles worst, so a
		// long distance measures the model rather than the assignment. Dropping the sample
		// would take away the width the cluster needs to reach that face.
		outlier := marker(far())
		outlier.SubjSrc = entity.SrcManual
		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), outlier}

		kept, dropped := centroidSamples(group, registered, "jsubject00000006")

		assert.Len(t, kept, 4, "a hand-set assignment is exempt from the outlier distance")
		assert.Zero(t, dropped)
	})
	t.Run("DropsManualBeyondAcceptance", func(t *testing.T) {
		// Seeding past the widest distance the cluster can accept would relink the marker
		// to a cluster that then refuses it, which is a link matching would never make and
		// cannot renew. The exemption reaches to that bound and no further.
		outlier := marker(beyondAcceptance())
		outlier.SubjSrc = entity.SrcManual
		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), outlier}

		kept, dropped := centroidSamples(group, registered, "jsubject00000007")

		require.Len(t, kept, 3, "acceptance bounds every source, hand-set included")
		assert.Equal(t, 1, dropped)
		for _, m := range kept {
			assert.NotEqual(t, outlier.MarkerUID, m.MarkerUID)
		}
	})
	t.Run("SeededSamplesStayReAcceptable", func(t *testing.T) {
		// The invariant the bound buys: whatever survives can be matched into the cluster
		// it seeded, so the migration never hands the matcher a link it would refuse.
		outlier := marker(far())
		outlier.SubjSrc = entity.SrcManual
		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), outlier}

		kept, _ := centroidSamples(group, registered, "jsubject00000008")
		require.NotEmpty(t, kept)

		embeddings := make(face.Embeddings, 0, len(kept))
		for _, m := range kept {
			embeddings = append(embeddings, m.Embeddings()[0])
		}

		midpoint, radius, _ := face.EmbeddingsMidpoint(embeddings)

		// Derived from the registered model rather than from face.AcceptDist, which reads
		// the process-wide thresholds another test may have pointed at a different model.
		accept := min(min(radius, registered.ClusterRadius)+registered.MatchDist, face.AcceptDistMax)

		for _, m := range kept {
			assert.LessOrEqual(t, m.Embeddings().Dist(midpoint), accept,
				"a seeded sample must be within the cluster's own accept distance")
		}
	})
	t.Run("SkipsIncomparableWidth", func(t *testing.T) {
		// Embeddings.Dist reports -1 when nothing is comparable, which must not read as
		// the closest possible sample and let a foreign vector define the centroid.
		odd := entity.Marker{MarkerUID: rnd.GenerateUID('m'), MarkerType: entity.MarkerFace}
		odd.SetEmbeddings(face.Embeddings{face.Embedding{1, 0}}, face.ModelSFace, face.EngineONNX)

		group := entity.Markers{marker(near(0)), marker(near(1)), marker(near(2)), odd}

		kept, _ := centroidSamples(group, registered, "jsubject00000005")
		assert.Len(t, kept, 3)
		for _, m := range kept {
			assert.NotEqual(t, odd.MarkerUID, m.MarkerUID, "an incomparable vector is not a sample")
		}
	})
}

func TestFaceMigrationSubjectCounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		subjects, assigned := faceMigrationSubjectCounts([]query.FaceMigrationIdentity{
			{MarkerUID: "m1", SubjUID: "s1", SubjSrc: entity.SrcManual},
			{MarkerUID: "m2", SubjUID: "s1", SubjSrc: entity.SrcAuto},
			{MarkerUID: "m3", SubjUID: "s2", SubjSrc: entity.SrcAuto},
		})
		assert.Equal(t, 2, subjects)
		assert.Equal(t, 3, assigned)
	})
	t.Run("NamedWithoutSubject", func(t *testing.T) {
		// Counting a name that has no subject row yet would report it as needing
		// attention, although the rebuild creates its subject right afterwards.
		subjects, assigned := faceMigrationSubjectCounts([]query.FaceMigrationIdentity{
			{MarkerUID: "m1", MarkerName: "Nameless", SubjSrc: entity.SrcManual},
		})
		assert.Zero(t, subjects)
		assert.Zero(t, assigned)
	})
	t.Run("Empty", func(t *testing.T) {
		subjects, assigned := faceMigrationSubjectCounts(nil)
		assert.Zero(t, subjects)
		assert.Zero(t, assigned)
	})
}

func TestMigrationCanceled(t *testing.T) {
	assert.NoError(t, migrationCanceled(context.Background(), nil))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, migrationCanceled(ctx, nil))
}

func TestFacesMigrateRerunError_Error(t *testing.T) {
	t.Run("IdentitiesChanged", func(t *testing.T) {
		// The rollback is the safe outcome, so the message has to describe the state it
		// leaves rather than the failure, and name the one action that resolves it.
		err := &FacesMigrateRerunError{Migrated: 12, Cause: query.ErrFaceMigrationIdentitiesChanged}

		assert.Contains(t, err.Error(), "a person assignment changed")
		assert.Contains(t, err.Error(), "nothing was lost")
		assert.Contains(t, err.Error(), "12 regenerated marker(s) stay unmatched")
		assert.Contains(t, err.Error(), "run again with the server stopped")
	})
	t.Run("Unwraps", func(t *testing.T) {
		// The identity case is the one a caller may want to tell apart from a storage error.
		err := error(&FacesMigrateRerunError{Migrated: 1, Cause: query.ErrFaceMigrationIdentitiesChanged})
		assert.ErrorIs(t, err, query.ErrFaceMigrationIdentitiesChanged)
	})
	t.Run("OtherCause", func(t *testing.T) {
		err := &FacesMigrateRerunError{Migrated: 0, Cause: errors.New("database is locked")}

		assert.Contains(t, err.Error(), "database is locked")
		assert.NotErrorIs(t, error(err), query.ErrFaceMigrationIdentitiesChanged)
	})
}

func TestMigrationProgress(t *testing.T) {
	// The loop is silent for the best part of an hour on a large library, so this is the only
	// signal a run is making progress. Every processed state counts, or a re-run that skips
	// most of the library would appear to stall.
	t.Run("CountsEveryProcessedState", func(t *testing.T) {
		plan := FacesMigratePlan{Markers: query.FaceMigrationMarkerCounts{Valid: 200}}
		result := FacesMigrateResult{Migrated: 40, Skipped: 30, Failed: 20, Retained: 10}

		done, total := migrationProgress(plan, result)

		assert.Equal(t, 100, done)
		assert.Equal(t, 200, total)
	})
	t.Run("NothingToDo", func(t *testing.T) {
		// Zero would divide by zero rather than report nothing.
		done, total := migrationProgress(FacesMigratePlan{}, FacesMigrateResult{})

		assert.Zero(t, done)
		assert.Zero(t, total)
		assert.NotPanics(t, func() { logMigrationProgress(FacesMigratePlan{}, FacesMigrateResult{}) })
	})
	t.Run("MoreProcessedThanPlanned", func(t *testing.T) {
		// The plan is counted before the run, so a marker added meanwhile must not report 101%.
		plan := FacesMigratePlan{Markers: query.FaceMigrationMarkerCounts{Valid: 10}}

		done, total := migrationProgress(plan, FacesMigrateResult{Migrated: 99})

		assert.Equal(t, 10, done)
		assert.Equal(t, 10, total)
	})
}

func TestFaces_useMigrationDetector(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)

	t.Run("LowersAndRestoresTheFloor", func(t *testing.T) {
		previous := face.ScoreThreshold

		restore, err := w.useMigrationDetector()
		require.NoError(t, err)
		require.NotNil(t, restore)

		assert.Equal(t, c.FaceMigrateScore(), face.ScoreThreshold)

		restore()

		assert.Equal(t, previous, face.ScoreThreshold)
	})
	t.Run("KeepsAConfiguredScore", func(t *testing.T) {
		// FACE_SCORE is a decision rather than a calibration, so the migration must not overrule
		// it the way it overrules the detector's own cutoff.
		previous := face.ScoreThreshold
		options := c.Options()
		configured := options.FaceScore
		options.FaceScore = 55
		t.Cleanup(func() { options.FaceScore = configured })

		restore, err := w.useMigrationDetector()
		require.NoError(t, err)

		assert.Equal(t, 55.0, face.ScoreThreshold)

		restore()

		assert.Equal(t, previous, face.ScoreThreshold)
	})
	t.Run("FollowsFaceMigrateScore", func(t *testing.T) {
		// The floor has to be tunable without a rebuild: it is the value the preview is
		// collecting data on, and a Docker image build per experiment is not a feedback loop.
		previous := face.ScoreThreshold
		options := c.Options()
		configured := options.FaceMigrateScore
		options.FaceMigrateScore = 25
		t.Cleanup(func() { options.FaceMigrateScore = configured })

		restore, err := w.useMigrationDetector()
		require.NoError(t, err)

		assert.Equal(t, 25.0, face.ScoreThreshold)
		assert.InDelta(t, 0.25, float64(face.ActiveEngineSettings().ONNX.ScoreThreshold), 0.0001)

		restore()

		assert.Equal(t, previous, face.ScoreThreshold)
	})
	t.Run("KeepsADisabledScore", func(t *testing.T) {
		// -1 switches the cutoff off, which is as much a decision as a number is. Treating it
		// as unset re-imposed a floor of 9 on an operator who had removed one.
		previous := face.ScoreThreshold
		options := c.Options()
		configured := options.FaceScore
		options.FaceScore = -1
		t.Cleanup(func() { options.FaceScore = configured })
		// FACE_MIGRATE_SCORE is unset, so the decision FACE_SCORE records carries into the run.

		restore, err := w.useMigrationDetector()
		require.NoError(t, err)

		assert.Equal(t, face.NoScoreThreshold, face.ScoreThreshold)

		restore()

		assert.Equal(t, previous, face.ScoreThreshold)
	})
	t.Run("MovesTheDetectionSessionToo", func(t *testing.T) {
		// The cutoff lives in the inference session as well as in the filter Detect applies
		// afterwards, so a run that moved only the second would have the first undo it.
		restore, err := w.useMigrationDetector()
		require.NoError(t, err)
		t.Cleanup(restore)

		settings := face.ActiveEngineSettings()

		assert.InDelta(t, c.FaceMigrateScore()/100, float64(settings.ONNX.ScoreThreshold), 0.0001)
	})
	t.Run("NilWorker", func(t *testing.T) {
		restore, err := (*Faces)(nil).useMigrationDetector()
		require.NoError(t, err)
		require.NotNil(t, restore)
		assert.NotPanics(t, restore)
	})
}
