package photoprism

import (
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// oneHotEmbedder returns a distinct unit vector per call, so migrated markers stay far
// enough apart to form separate clusters and the audit can normalize them.
type oneHotEmbedder struct {
	dims int
	n    int
}

// ModelName returns the target model this test embedder stands in for.
func (e *oneHotEmbedder) ModelName() face.ModelName { return face.ModelFaceNet }

// Dims returns the test embedding length.
func (e *oneHotEmbedder) Dims() int { return e.dims }

// CropSize returns the test crop size.
func (e *oneHotEmbedder) CropSize() (int, int) { return 16, 16 }

// Aligned reports that this embedder works on plain box crops.
func (e *oneHotEmbedder) Aligned() bool { return false }

// Run returns the next one-hot vector.
func (e *oneHotEmbedder) Run(image.Image) face.Embeddings {
	v := make(face.Embedding, e.dims)
	v[e.n%e.dims] = 1
	e.n++

	return face.Embeddings{v}
}

// Close releases the test embedder.
func (e *oneHotEmbedder) Close() error { return nil }

// alignedEmbedder stands in for a model that consumes landmarks, which is what makes the
// detector that placed them part of what a marker's vector depends on.
type alignedEmbedder struct {
	oneHotEmbedder
}

// Aligned reports that this embedder needs an aligned crop.
func (e *alignedEmbedder) Aligned() bool { return true }

// sameFaceEmbedder returns one person's faces as a loose group: about 0.73 apart from each
// other, so a single-sample cluster cannot reach them, and about 0.46 from their common
// centroid, so a cluster holding all of them can. That gap is what the sample set buys, and
// a test whose vectors sit closer together passes no matter how the cluster is seeded.
type sameFaceEmbedder struct {
	model face.ModelName
	dims  int
	n     int
}

// ModelName returns the target model this test embedder stands in for.
func (e *sameFaceEmbedder) ModelName() face.ModelName { return e.model }

// Dims returns the test embedding length.
func (e *sameFaceEmbedder) Dims() int { return e.dims }

// CropSize returns the test crop size.
func (e *sameFaceEmbedder) CropSize() (int, int) { return 16, 16 }

// Aligned reports that this embedder works on plain box crops.
func (e *sameFaceEmbedder) Aligned() bool { return false }

// Run returns the next vector of the same cluster.
func (e *sameFaceEmbedder) Run(image.Image) face.Embeddings {
	v := make([]float32, e.dims)
	v[0] = 1
	v[1+e.n%(e.dims-1)] = 0.6
	e.n++

	return face.NewEmbeddings([][]float32{v})
}

// Close releases the test embedder.
func (e *sameFaceEmbedder) Close() error { return nil }

// renamingEmbedder renames a marker the first time it runs, which is what a rename through
// the API looks like to a migration that snapshotted the identities before its batch loop.
type renamingEmbedder struct {
	oneHotEmbedder
	markerUID string
	renamed   bool
}

// Run renames the target marker once and then embeds as usual.
func (e *renamingEmbedder) Run(img image.Image) face.Embeddings {
	if !e.renamed {
		e.renamed = true
		_ = entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", e.markerUID).
			UpdateColumn("marker_name", "Someone Else").Error
	}

	return e.oneHotEmbedder.Run(img)
}

// stoppingEmbedder embeds like oneHotEmbedder and cancels the run once it has produced limit
// vectors, which leaves the checkpoints an interrupted migration leaves behind.
type stoppingEmbedder struct {
	oneHotEmbedder
	limit  int
	cancel context.CancelFunc
}

// Run returns the next one-hot vector and cancels the run once the limit is reached.
func (e *stoppingEmbedder) Run(img image.Image) face.Embeddings {
	v := e.oneHotEmbedder.Run(img)

	if e.n >= e.limit {
		e.cancel()
	}

	return v
}

// missingFaceEmbedder returns no vector for any crop, as an embedder that fails.
type missingFaceEmbedder struct {
	oneHotEmbedder
}

// Run returns no embeddings.
func (e *missingFaceEmbedder) Run(image.Image) face.Embeddings { return nil }

// addMigrateTestLibrary creates one named marker per file. Small markers start below the cluster
// size bar, which the extent a migration samples them at clears.
func addMigrateTestLibrary(t *testing.T, c *config.Config, files int, small bool) {
	t.Helper()

	for i := range files {
		f := addMigrateTestFile(t, c, fmt.Sprintf("%040d", 7000+i), true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, fmt.Sprintf("Person %d", i))

		if small {
			require.NoError(t, entity.UnscopedDb().Model(m).
				UpdateColumns(entity.Values{"size": 10, "thumb_size": 10}).Error)
			require.NoError(t, entity.UnscopedDb().Model(f).
				UpdateColumns(entity.Values{"file_width": 720, "file_height": 720}).Error)
		}
	}
}

// checkpointMigrateTestMarkers saves target vectors for the first n markers the way an aligned run
// does, which leaves the state an interrupted run of such a migration leaves behind.
func checkpointMigrateTestMarkers(t *testing.T, n int) (remaining []string) {
	t.Helper()

	var markers entity.Markers
	require.NoError(t, entity.UnscopedDb().Order("file_uid").Find(&markers).Error)
	require.GreaterOrEqual(t, len(markers), n)

	detector := face.ActiveDetector()
	require.NotEqual(t, face.DetectorNone, detector, "the detector decides which crops are stale")

	for i, m := range markers {
		if i >= n {
			remaining = append(remaining, m.MarkerUID)
			continue
		}

		v := make(face.Embedding, 4)
		v[i%4] = 1

		require.NoError(t, query.SaveFaceMigrationEmbeddings(face.ModelFaceNet, detector,
			map[string]face.Embeddings{m.MarkerUID: {v}},
			map[string]query.MigrationDetection{m.MarkerUID: {Size: face.ClusterSizeThreshold,
				Score: face.ClusterScore(detector) + 10, ThumbSize: face.ClusterSizeThreshold}}))
	}

	return remaining
}

// deleteMigrateTestFiles removes the last n file rows from the index, as a purge does when the
// originals are gone, and leaves their markers behind.
func deleteMigrateTestFiles(t *testing.T, n int) {
	t.Helper()

	var files entity.Files
	require.NoError(t, entity.UnscopedDb().Order("file_uid DESC").Limit(n).Find(&files).Error)
	require.Len(t, files, n)

	for i := range files {
		require.NoError(t, entity.Db().Where("file_uid = ?", files[i].FileUID).Delete(&entity.File{}).Error)
	}
}

// removeMigrateTestFiles removes the original and the cached thumbnail of the first n files, in the
// order a migration visits them, as on an unmounted volume with no cache.
func removeMigrateTestFiles(t *testing.T, c *config.Config, n int) {
	t.Helper()

	var files entity.Files
	require.NoError(t, entity.UnscopedDb().Order("file_uid").Limit(n).Find(&files).Error)
	require.Len(t, files, n)

	for _, file := range files {
		name, err := thumb.Sizes[thumb.Fit720].FileName(file.FileHash, c.ThumbCachePath())
		require.NoError(t, err)
		require.NoError(t, os.Remove(name))
		require.NoError(t, os.Remove(FileName(file.FileRoot, file.FileName)))
	}
}

// interruptedMigration runs a migration to FaceNet over the library that stops after migrated
// markers, and returns the markers the run left on their previous model.
func interruptedMigration(t *testing.T, w *Faces, files, migrated int) (remaining []string) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	plan := FacesMigratePlan{Target: face.ModelFaceNet}
	embedder := &stoppingEmbedder{oneHotEmbedder: oneHotEmbedder{dims: 4}, limit: migrated, cancel: cancel}
	result, err := w.migrate(ctx, plan, embedder, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, migrated, result.Migrated, "the first run checkpoints what it re-embedded")
	require.False(t, result.ModelRecorded)

	require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
		Where("embed_model <> ?", face.ModelFaceNet).Pluck("marker_uid", &remaining).Error)
	require.Len(t, remaining, files-migrated)

	return remaining
}

// newMigrateTestConfig returns an isolated config for a migration test.
// savedFaceModel returns the face model recorded in the config's "options.yml", or an empty
// string when the file names none. Reading the file is what tells a persisted setting apart
// from one this process is merely holding.
func savedFaceModel(t *testing.T, c *config.Config) string {
	t.Helper()

	b, err := os.ReadFile(c.OptionsYaml())

	if err != nil {
		return ""
	}

	values := map[string]any{}
	require.NoError(t, yaml.Unmarshal(b, &values))

	name, _ := values["FaceModel"].(string)

	return name
}

// snapshotFaceCalibration returns a function that restores the package values a change of
// embedding model moves, which a committed migration propagates for the rest of the process.
func snapshotFaceCalibration() func() {
	clusterSize, collisionDist, epsilon := face.ClusterSizeThreshold, face.CollisionDist, face.Epsilon
	clusterRadius, clusterDist, matchDist := face.ClusterRadius, face.ClusterDist, face.MatchDist

	return func() {
		face.ClusterSizeThreshold, face.CollisionDist, face.Epsilon = clusterSize, collisionDist, epsilon
		face.ClusterRadius, face.ClusterDist, face.MatchDist = clusterRadius, clusterDist, matchDist
	}
}

func newMigrateTestConfig(t *testing.T, name string) *config.Config {
	t.Helper()
	useTestDb(t, name)

	oldConfig := Config()

	c := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	require.NoError(t, c.CreateDirectories())

	SetConfig(c)
	restoreCalibration := snapshotFaceCalibration()
	t.Cleanup(func() {
		SetConfig(oldConfig)
		_ = c.CloseDb()
		oldConfig.RegisterDb()
		restoreCalibration()

		// Initializing a config reconfigures the process-wide embedder from its models
		// path, which is empty here, so the previous one has to be reinstated.
		restoreEmbedder(t)
	})

	// The isolated database is seeded from the cached test database, so migration would
	// otherwise walk the shared fixtures instead of the rows this test creates.
	for _, model := range []any{&entity.Marker{}, &entity.Face{}, &entity.Subject{}, &entity.File{}} {
		require.NoError(t, entity.UnscopedDb().Unscoped().Where("1 = 1").Delete(model).Error)
	}

	return c
}

// addMigrateTestFile creates an indexed file, optionally with a cached thumbnail so the
// migration can read a crop from it. Without one, every marker on it fails to re-embed.
func addMigrateTestFile(t *testing.T, c *config.Config, hash string, withThumb bool) *entity.File {
	t.Helper()

	f := &entity.File{
		FileUID:  rnd.GenerateUID('f'),
		PhotoUID: rnd.GenerateUID('p'),
		FileHash: hash,
		FileName: hash + ".jpg",
		FileRoot: entity.RootOriginals,
	}

	require.NoError(t, entity.Db().Create(f).Error)

	if withThumb {
		name, err := thumb.Sizes[thumb.Fit720].FileName(hash, c.ThumbCachePath())
		require.NoError(t, err)
		require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 720, 720)), name))

		// Cropping falls back to the original when the cache lookup misses, so both have to
		// exist for a marker on this file to be re-embedded.
		require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 720, 720)), FileName(f.FileRoot, f.FileName)))
	}

	return f
}

// addMigrateTestMarker creates a face marker with a stale embedding on the given file. It records a
// sample extent, so a case turns on the model or the detector rather than on that column.
func addMigrateTestMarker(t *testing.T, fileUID, subjSrc, name string) *entity.Marker {
	t.Helper()

	m := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        fileUID,
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		Size:           face.ClusterSizeThreshold,
		ThumbSize:      face.ClusterSizeThreshold,
		Score:          face.ClusterScore("") + 10,
		X:              0,
		Y:              0,
		W:              1,
		H:              1,
		EmbeddingsJSON: face.Embeddings{{0.1, 0.2, 0.3, 0.4}}.JSON(),
		EmbedModel:     face.ModelSFace,
		SubjSrc:        subjSrc,
		MarkerName:     name,
	}

	if subjSrc != "" && subjSrc != entity.SrcAuto {
		subj := entity.NewSubject(name, entity.SubjPerson, subjSrc)
		require.NotNil(t, subj)
		require.NoError(t, entity.Db().Create(subj).Error)
		m.SubjUID = subj.SubjUID
	}

	require.NoError(t, entity.Db().Create(m).Error)

	return m
}

// addMigrateTestSubjectMarker creates a face marker assigned to an existing subject by an
// automatic match, which is how most of a curated library's faces are attributed.
func addMigrateTestSubjectMarker(t *testing.T, fileUID, subjUID string, name string) *entity.Marker {
	t.Helper()

	m := addMigrateTestMarker(t, fileUID, entity.SrcAuto, name)
	require.NoError(t, entity.Db().Model(m).UpdateColumn("subj_uid", subjUID).Error)
	m.SubjUID = subjUID

	return m
}

// countFaceRows returns the number of stored face clusters.
func countFaceRows(t *testing.T) int {
	t.Helper()

	var n int64
	require.NoError(t, entity.UnscopedDb().Model(&entity.Face{}).Count(&n).Error)

	return int(n)
}

func TestFinalizeRefused(t *testing.T) {
	t.Run("NothingAttempted", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(FacesMigrateResult{}))
	})
	t.Run("AllFailed", func(t *testing.T) {
		assert.Contains(t, finalizeRefused(FacesMigrateResult{Failed: 15}), "none of 15")
	})
	t.Run("AboveRatio", func(t *testing.T) {
		assert.Contains(t, finalizeRefused(FacesMigrateResult{Migrated: 1, Failed: 9, AttemptedClusterable: 10, FailedClusterable: 9}), "9 of 10")
	})
	t.Run("WithinRatio", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(FacesMigrateResult{Migrated: 99, Failed: 1, AttemptedClusterable: 100, FailedClusterable: 1}))
	})
	t.Run("NoFailures", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(FacesMigrateResult{Migrated: 10}))
	})
	t.Run("StorageOutageOnALibraryWithFewClusterableMarkers", func(t *testing.T) {
		// The numerator counts clusterable failures, so the denominator has to count clusterable
		// attempts. Dividing by every attempted marker made the bound unreachable whenever the
		// clusterable share was under it - here 7 % - so a total outage finalized and blanked
		// every failed vector, which is the case the guard was written for.
		reason := finalizeRefused(FacesMigrateResult{
			Migrated:             10000,
			Failed:               190000,
			Unreadable:           190000,
			ReadErrors:           190000,
			AttemptedClusterable: 14000,
			FoundClusterable:     14000,
			FailedClusterable:    13300,
			FaultsClusterable:    13300,
		})

		assert.Contains(t, reason, "could not be re-embedded")
		assert.Contains(t, reason, "their file is missing or unreadable")
	})
	t.Run("NothingClusterableFallsBackToEveryMarker", func(t *testing.T) {
		// A library with no clusterable markers has no population to measure, so the bound must
		// fall back rather than pass by default - which is also what stops FACE_CLUSTER_SIZE
		// or FACE_CLUSTER_SCORE from switching the guard off.
		assert.NotEmpty(t, finalizeRefused(FacesMigrateResult{Migrated: 1, Failed: 99}))
		assert.Empty(t, finalizeRefused(FacesMigrateResult{Migrated: 99, Failed: 1}))
	})
	t.Run("FailuresBelowTheClusteringBars", func(t *testing.T) {
		// Most of what a detector change fails to re-find is too small or too low-scoring to
		// seed or join a cluster, so its vector changes nothing. Counting it refused two of
		// three real libraries over a loss the library could not have used.
		assert.Empty(t, finalizeRefused(FacesMigrateResult{Migrated: 41063, Failed: 7993, AttemptedClusterable: 41272, FailedClusterable: 209}))
	})
	t.Run("ResumedWithinTheBound", func(t *testing.T) {
		// A resumed run retries only what failed before, so it migrates nothing new; the markers
		// earlier runs moved to the target model are its successes.
		assert.Empty(t, finalizeRefused(FacesMigrateResult{
			Skipped: 1416, SkippedClusterable: 600, Failed: 168, AttemptedClusterable: 30, FailedClusterable: 30,
		}))
	})
	t.Run("ResumedAboveTheBound", func(t *testing.T) {
		reason := finalizeRefused(FacesMigrateResult{
			Skipped: 100, SkippedClusterable: 100, Failed: 50, AttemptedClusterable: 50, FailedClusterable: 50,
		})
		assert.Contains(t, reason, "50 of 150 marker(s) that clear the clustering thresholds")
	})
	t.Run("ResumedNothingClusterableFallsBackToEveryMarker", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(FacesMigrateResult{Skipped: 99, Failed: 1}))
		assert.NotEmpty(t, finalizeRefused(FacesMigrateResult{Skipped: 1, Failed: 99}))
	})
	t.Run("ResumedDecidesLikeOneRun", func(t *testing.T) {
		// However the successes are split between earlier runs and this one, the decision has to
		// be the one a single uninterrupted run with the same outcomes would reach.
		for failed := 0; failed <= 30; failed++ {
			for succeeded := 0; succeeded <= 120; succeeded += 7 {
				single := FacesMigrateResult{Migrated: succeeded, Failed: failed,
					AttemptedClusterable: succeeded + failed, FailedClusterable: failed}
				want := finalizeRefused(single) == ""

				for earlier := 0; earlier <= succeeded; earlier++ {
					resumed := FacesMigrateResult{Migrated: succeeded - earlier, Skipped: earlier, Failed: failed,
						AttemptedClusterable: succeeded - earlier + failed, FailedClusterable: failed, SkippedClusterable: earlier}

					if got := finalizeRefused(resumed) == ""; got != want {
						t.Fatalf("%d succeeded (%d earlier), %d failed: resumed finalizes %t, one run %t",
							succeeded, earlier, failed, got, want)
					}
				}
			}
		}
	})
	t.Run("PredictedUnreadableJudgedWithTheMigration", func(t *testing.T) {
		// A file the index no longer lists fails on every run, so it is judged with the whole
		// migration like a face that is not found again, not as a read error of this run.
		assert.Empty(t, finalizeRefused(FacesMigrateResult{
			Skipped: 1000, SkippedClusterable: 900, Failed: 5, Unreadable: 5, AttemptedClusterable: 5, FailedClusterable: 5,
		}))
	})
	t.Run("ResumedUnreadableIsRefused", func(t *testing.T) {
		// Earlier successes say nothing about a file this run could not read.
		reason := finalizeRefused(FacesMigrateResult{
			Skipped: 1000, SkippedClusterable: 900, Failed: 5, Unreadable: 5, ReadErrors: 5, AttemptedClusterable: 5,
			FailedClusterable: 5, FaultsClusterable: 5,
		})
		assert.Contains(t, reason, "none of 5 marker(s)")
		assert.Contains(t, reason, "their file is missing or unreadable")
	})
	t.Run("UnreadableJudgedByTheCurrentRun", func(t *testing.T) {
		// 3 of 20 attempted is above the bound for this run, although it is under 1 % of the library.
		reason := finalizeRefused(FacesMigrateResult{
			Migrated: 17, Skipped: 1000, SkippedClusterable: 900, Failed: 3, Unreadable: 3, ReadErrors: 3, AttemptedClusterable: 20,
			FoundClusterable: 20, FailedClusterable: 3, FaultsClusterable: 3,
		})
		assert.Contains(t, reason, "3 of 20 marker(s)")
		// A share of unreadable files within the bound of this run leaves the decision to the whole.
		assert.Empty(t, finalizeRefused(FacesMigrateResult{
			Migrated: 99, Skipped: 1000, SkippedClusterable: 900, Failed: 1, Unreadable: 1, ReadErrors: 1, AttemptedClusterable: 100,
			FoundClusterable: 100, FailedClusterable: 1, FaultsClusterable: 1,
		}))
	})
	t.Run("FreshRunWithoutSuccessesIsRefused", func(t *testing.T) {
		assert.Contains(t, finalizeRefused(FacesMigrateResult{Failed: 15, AttemptedClusterable: 15, FailedClusterable: 15}), "none of 15")
	})
	t.Run("NamesTheDetectorWhenNoFileFailed", func(t *testing.T) {
		reason := finalizeRefused(FacesMigrateResult{Migrated: 1, Failed: 9, AttemptedClusterable: 10, FailedClusterable: 9})
		assert.Contains(t, reason, "the detector did not find their face again")
		assert.NotContains(t, reason, "unreadable")
	})
	t.Run("NamesStorageWhenEveryFileFailed", func(t *testing.T) {
		reason := finalizeRefused(FacesMigrateResult{Migrated: 1, Failed: 9, Unreadable: 9, AttemptedClusterable: 10, FailedClusterable: 9})
		assert.Contains(t, reason, "their file is missing or unreadable")
		assert.NotContains(t, reason, "did not find")
	})
	t.Run("NamesBoth", func(t *testing.T) {
		reason := finalizeRefused(FacesMigrateResult{Migrated: 1, Failed: 9, Unreadable: 4, AttemptedClusterable: 10, FailedClusterable: 9})
		assert.Contains(t, reason, "5 because the detector did not find their face again")
		assert.Contains(t, reason, "4 because their file is missing or unreadable")
	})
}

func TestFaultsRefused(t *testing.T) {
	t.Run("NoFaults", func(t *testing.T) {
		assert.Empty(t, faultsRefused(FacesMigrateResult{Failed: 5, Unreadable: 5}))
	})
	t.Run("NothingMigrated", func(t *testing.T) {
		assert.Contains(t, faultsRefused(FacesMigrateResult{Skipped: 100, Failed: 1, Unreadable: 1, ReadErrors: 1}), "none of 1")
	})
	t.Run("AboveTheBoundOfThisRun", func(t *testing.T) {
		assert.Contains(t, faultsRefused(FacesMigrateResult{Migrated: 8, Failed: 2, Unreadable: 2, ReadErrors: 2,
			AttemptedClusterable: 10, FoundClusterable: 10, FailedClusterable: 2, FaultsClusterable: 2}), "2 of 10")
	})
	t.Run("AttemptsCountedAsFound", func(t *testing.T) {
		// Successes the save lifted over the bars must not hide read errors below them.
		reason := faultsRefused(FacesMigrateResult{Migrated: 3, Failed: 7, Unreadable: 7, ReadErrors: 7, AttemptedClusterable: 3})
		assert.Equal(t, "7 of 10 marker(s) could not be re-embedded, because their file is missing or unreadable", reason)
	})
	t.Run("EmbedErrors", func(t *testing.T) {
		// A face found and not embedded is a fault of this run, like a file it could not read.
		assert.Contains(t, faultsRefused(FacesMigrateResult{Skipped: 100, SkippedClusterable: 100, Failed: 1, EmbedErrors: 1}), "none of 1")
		assert.Equal(t, "2 of 10 marker(s) that clear the clustering thresholds could not be re-embedded, because their face could not be embedded",
			faultsRefused(FacesMigrateResult{Migrated: 8, Failed: 2, EmbedErrors: 2, FoundClusterable: 10, FaultsClusterable: 2}))
	})
	t.Run("NothingClusterableFallsBackToEveryMarker", func(t *testing.T) {
		assert.Contains(t, faultsRefused(FacesMigrateResult{Migrated: 8, Failed: 2, Unreadable: 2, ReadErrors: 2}), "2 of 10 marker(s) could not")
		assert.Empty(t, faultsRefused(FacesMigrateResult{Migrated: 99, Failed: 1, Unreadable: 1, ReadErrors: 1}))
	})
}

func TestFaultCause(t *testing.T) {
	t.Run("ReadErrors", func(t *testing.T) {
		assert.Equal(t, ", because their file is missing or unreadable", faultCause(FacesMigrateResult{ReadErrors: 2}))
	})
	t.Run("EmbedErrors", func(t *testing.T) {
		assert.Equal(t, ", because their face could not be embedded", faultCause(FacesMigrateResult{EmbedErrors: 2}))
	})
	t.Run("Both", func(t *testing.T) {
		assert.Contains(t, faultCause(FacesMigrateResult{ReadErrors: 1, EmbedErrors: 1}), "unreadable or their face could not be embedded")
	})
}

func TestFailureRatioRefused(t *testing.T) {
	t.Run("WithinTheBound", func(t *testing.T) {
		assert.Empty(t, failureRatioRefused(1, 10, "", ", because"))
	})
	t.Run("AboveTheBound", func(t *testing.T) {
		reason := failureRatioRefused(2, 10, " that clear the clustering thresholds", ", because the detector did not find their face again")
		assert.Equal(t, "2 of 10 marker(s) that clear the clustering thresholds could not be re-embedded, because the detector did not find their face again", reason)
	})
	t.Run("NoPopulation", func(t *testing.T) {
		assert.Empty(t, failureRatioRefused(0, 0, "", ""))
	})
}

func TestMigrationFailureCause(t *testing.T) {
	t.Run("NoFailures", func(t *testing.T) {
		assert.Empty(t, migrationFailureCause(FacesMigrateResult{Migrated: 10}))
	})
	t.Run("MoreUnreadableThanFailed", func(t *testing.T) {
		// Cannot happen, but the subtraction would report a negative count if it did.
		assert.Equal(t, ", because their file is missing or unreadable", migrationFailureCause(FacesMigrateResult{Failed: 2, Unreadable: 5}))
	})
	t.Run("EmbedErrors", func(t *testing.T) {
		assert.Equal(t, ", because their face could not be embedded", migrationFailureCause(FacesMigrateResult{Failed: 3, EmbedErrors: 3}))
	})
	t.Run("AllThree", func(t *testing.T) {
		assert.Equal(t, ", 1 because the detector did not find their face again, 2 because their face could not be embedded, "+
			"and 3 because their file is missing or unreadable",
			migrationFailureCause(FacesMigrateResult{Failed: 6, EmbedErrors: 2, Unreadable: 3}))
	})
}

func TestFaces_migrate(t *testing.T) {
	t.Run("ResumeFinalizesWithinTheBound", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateresume")
		w := NewFaces(c)

		// 12 of 13 markers reached the target model in an earlier run; detection does not find the
		// last face again, which is 1 of 13 for the migration as a whole and within the bound.
		addMigrateTestLibrary(t, c, 13, false)
		remaining := checkpointMigrateTestMarkers(t, 12)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &alignedEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var incomplete *FacesMigrateIncompleteError
		require.ErrorAs(t, err, &incomplete, "the resumed run finalizes and reports what it lost")
		assert.Zero(t, result.Migrated)
		assert.Equal(t, 12, result.Skipped)
		assert.Equal(t, 1, result.Failed)
		assert.Zero(t, result.ReadErrors+result.EmbedErrors, "a face not found again is not a fault")
		assert.True(t, result.ModelRecorded)
		assert.Equal(t, face.ModelFaceNet, savedFaceModel(t, c))

		lost := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&lost, "marker_uid = ?", remaining[0]).Error)
		assert.Empty(t, lost.EmbeddingsJSON, "the vector that could not be regenerated is cleared")
		assert.NotEmpty(t, lost.SubjUID, "the person assignment is kept")
		assert.NotEmpty(t, lost.MarkerName)
	})
	t.Run("ResumeAboveTheBoundIsRefused", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateresumerefused")
		w := NewFaces(c)

		// 8 of 13 in an earlier run, so the 5 faces not found again are above the bound.
		addMigrateTestLibrary(t, c, 13, false)
		remaining := checkpointMigrateTestMarkers(t, 8)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &alignedEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var aborted *FacesMigrateAbortedError
		require.ErrorAs(t, err, &aborted)
		assert.Contains(t, aborted.Reason, "5 of 13 marker(s) that clear the clustering thresholds")
		assert.Equal(t, 8, aborted.Unmatched, "the markers earlier runs regenerated stay unmatched")
		assert.Contains(t, aborted.Error(), "5 marker(s) whose face was not found again fail again with the same configuration")
		assert.False(t, result.ModelRecorded)
		assert.Zero(t, countFaceRows(t))

		kept := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&kept, "marker_uid = ?", remaining[0]).Error)
		assert.NotEmpty(t, kept.EmbeddingsJSON, "a refused run keeps every previous vector")
	})
	t.Run("ResumeWithEmbedErrorsIsRefused", func(t *testing.T) {
		// A face found and not embedded is a fault of this run, which earlier successes do not dilute.
		c := newMigrateTestConfig(t, "migrateresumeembed")
		w := NewFaces(c)

		addMigrateTestLibrary(t, c, 13, false)
		interruptedMigration(t, w, 13, 12)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &missingFaceEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var aborted *FacesMigrateAbortedError
		require.ErrorAs(t, err, &aborted)
		assert.Equal(t, 1, result.EmbedErrors)
		assert.Contains(t, aborted.Reason, "because their face could not be embedded")
		assert.Contains(t, aborted.Error(), "1 marker(s) that could not be read or embedded may succeed once the cause is fixed")
		assert.False(t, result.ModelRecorded)
	})
	t.Run("ResumeWithUnreadableFilesIsRefused", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateresumeunreadable")
		w := NewFaces(c)

		addMigrateTestLibrary(t, c, 13, false)
		remaining := interruptedMigration(t, w, 13, 12)

		// The original and its thumbnail cannot be read, as on an unmounted volume with no cache.
		var marker entity.Marker
		require.NoError(t, entity.UnscopedDb().First(&marker, "marker_uid = ?", remaining[0]).Error)
		var file entity.File
		require.NoError(t, entity.UnscopedDb().First(&file, "file_uid = ?", marker.FileUID).Error)
		name, err := thumb.Sizes[thumb.Fit720].FileName(file.FileHash, c.ThumbCachePath())
		require.NoError(t, err)
		require.NoError(t, os.Remove(name))
		require.NoError(t, os.Remove(FileName(file.FileRoot, file.FileName)))

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var aborted *FacesMigrateAbortedError
		require.ErrorAs(t, err, &aborted, "earlier successes must not let an unreadable file through")
		assert.Equal(t, 1, result.ReadErrors)
		assert.Equal(t, 1, result.Unreadable)
		assert.Contains(t, aborted.Reason, "their file is missing or unreadable")
		assert.Contains(t, aborted.Error(), "may succeed once the cause is fixed")
		assert.False(t, result.ModelRecorded)
	})
	t.Run("ResumeWithADeletedFileFinalizes", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateresumedeleted")
		w := NewFaces(c)

		addMigrateTestLibrary(t, c, 13, false)
		remaining := interruptedMigration(t, w, 13, 12)

		// A file the index no longer lists fails on every run, so it is judged with the migration.
		var marker entity.Marker
		require.NoError(t, entity.UnscopedDb().First(&marker, "marker_uid = ?", remaining[0]).Error)
		require.NoError(t, entity.Db().Where("file_uid = ?", marker.FileUID).Delete(&entity.File{}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var incomplete *FacesMigrateIncompleteError
		require.ErrorAs(t, err, &incomplete)
		assert.Equal(t, 1, result.Unreadable)
		assert.Zero(t, result.ReadErrors)
		assert.True(t, result.ModelRecorded)
	})
	t.Run("ResumeDecidesLikeOneRun", func(t *testing.T) {
		// The same library and the same per-marker outcomes, once in one run and once interrupted
		// and resumed, must reach the same decision - including markers whose saved extent lifts
		// them over the cluster size bar. Files the index no longer lists fail in either.
		for _, tc := range []struct {
			name    string
			files   int
			deleted int
			earlier int
			small   bool
			refused bool
		}{
			{"Finalizes", 13, 1, 6, false, false},
			{"Refused", 13, 5, 4, false, true},
			// The failures sit below the size bar and the saved successes above it.
			{"SmallMarkers", 10, 7, 2, true, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				plan := FacesMigratePlan{Target: face.ModelFaceNet}
				opt := FacesMigrateOptions{Target: face.ModelFaceNet}

				one := newMigrateTestConfig(t, "migrateonerun")
				addMigrateTestLibrary(t, one, tc.files, tc.small)
				deleteMigrateTestFiles(t, tc.deleted)
				single, singleErr := NewFaces(one).migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt,
					FacesMigrateResult{Target: face.ModelFaceNet})
				require.Equal(t, tc.files-tc.deleted, single.Migrated)
				require.Equal(t, tc.deleted, single.Failed)

				if tc.small {
					require.Zero(t, single.FailedClusterable, "small markers fail below the size bar")
					require.Equal(t, single.Migrated, single.AttemptedClusterable, "the saved extent lifts a success over it")
				}

				var singleAborted *FacesMigrateAbortedError
				singleRefused := errors.As(singleErr, &singleAborted)
				require.Equal(t, tc.refused, singleRefused, "one run: %v", singleErr)

				two := newMigrateTestConfig(t, "migratetworuns")
				w := NewFaces(two)
				addMigrateTestLibrary(t, two, tc.files, tc.small)
				deleteMigrateTestFiles(t, tc.deleted)
				interruptedMigration(t, w, tc.files, tc.earlier)
				resumed, resumedErr := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})

				var resumedAborted *FacesMigrateAbortedError
				assert.Equal(t, singleRefused, errors.As(resumedErr, &resumedAborted), "one run: %v, resumed: %v", singleErr, resumedErr)
				assert.Equal(t, single.AttemptedClusterable, resumed.AttemptedClusterable+resumed.SkippedClusterable)
				assert.Equal(t, single.FailedClusterable, resumed.FailedClusterable)
			})
		}
	})
	t.Run("OutageOnSmallMarkersIsRefused", func(t *testing.T) {
		// Read errors are judged by the attempts as found, so successes the save lifts over the
		// size bar cannot hide an outage among markers that were below it.
		c := newMigrateTestConfig(t, "migratesmalloutage")
		w := NewFaces(c)

		addMigrateTestLibrary(t, c, 10, true)
		removeMigrateTestFiles(t, c, 7)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		var aborted *FacesMigrateAbortedError
		require.ErrorAs(t, err, &aborted)
		assert.Equal(t, 3, result.Migrated)
		assert.Equal(t, 7, result.ReadErrors)
		assert.Contains(t, aborted.Reason, "7 of 10 marker(s) could not be re-embedded")
	})
	t.Run("AlignedReadErrorsAreCounted", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migratealignedread")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "8888888888888888888888888888888888888888", false)
		addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &alignedEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.Error(t, err)
		assert.Equal(t, 1, result.ReadErrors)
		assert.Equal(t, 1, result.FaultsClusterable)
		assert.Equal(t, 1, result.Unreadable)
	})
	t.Run("RetainedCropFailureIsNotAReadError", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateretainedread")
		w := NewFaces(c)

		// On the target model with an unsettled extent, so it only needs a new crop and keeps its
		// vector when the crop fails.
		f := addMigrateTestFile(t, c, "9999999999999999999999999999999999999999", false)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")
		require.NoError(t, entity.UnscopedDb().Model(m).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "thumb_size": 0}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Retained)
		assert.Zero(t, result.Unreadable)
		assert.Zero(t, result.ReadErrors)
	})
	t.Run("ForcedRunNamesEmbedErrors", func(t *testing.T) {
		// A forced run finalizes through faults, and its closing message names them as recoverable.
		c := newMigrateTestConfig(t, "migrateforcedembed")
		w := NewFaces(c)

		addMigrateTestLibrary(t, c, 2, false)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &missingFaceEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet, Force: true}, FacesMigrateResult{Target: face.ModelFaceNet})

		var incomplete *FacesMigrateIncompleteError
		require.ErrorAs(t, err, &incomplete)
		assert.Equal(t, 2, result.EmbedErrors)
		assert.Equal(t, 2, incomplete.EmbedErrors)
		assert.Contains(t, err.Error(), "2 lost marker(s) were found but could not be embedded")
	})
	t.Run("RetainedEmbedFailureIsNotAFault", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateretainedembed")
		w := NewFaces(c)

		// The crop is read and not embedded, but the marker keeps its vector, so it is not a fault.
		f := addMigrateTestFile(t, c, "9898989898989898989898989898989898989898", true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")
		require.NoError(t, entity.UnscopedDb().Model(m).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "thumb_size": 0}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &missingFaceEmbedder{oneHotEmbedder{dims: 4}},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Retained)
		assert.Zero(t, result.EmbedErrors)

		kept := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&kept, "marker_uid = ?", m.MarkerUID).Error)
		assert.NotEmpty(t, kept.EmbeddingsJSON, "a retained marker keeps its vector")
	})
	t.Run("IdentityChangedMidRun", func(t *testing.T) {
		// The finalize rolls back when a person assignment moved under it. What the
		// operator must not be left with is a bare error: the clusters are still the old
		// model's while the markers this run regenerated are the new one's.
		c := newMigrateTestConfig(t, "migraterename")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "3333333333333333333333333333333333333333", true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		embedder := &renamingEmbedder{oneHotEmbedder: oneHotEmbedder{dims: 4}, markerUID: m.MarkerUID}
		result, err := w.migrate(context.Background(), plan, embedder,
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.Error(t, err)
		assert.IsType(t, &FacesMigrateRerunError{}, err)
		assert.ErrorIs(t, err, query.ErrFaceMigrationIdentitiesChanged)
		assert.Contains(t, err.Error(), "run again")
		assert.Positive(t, result.Migrated, "the rollback happens after markers were regenerated")
		assert.False(t, result.ModelRecorded, "a rolled-back run records no model, so it asks for no restart")

		// The rollback has to leave the clusters alone, which is what makes re-running safe.
		assert.Zero(t, countFaceRows(t))
		assert.NotEqual(t, face.ModelFaceNet, savedFaceModel(t, c))
	})
	t.Run("RefusesTotalFailure", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migraterefuse")
		w := NewFaces(c)

		// No cached thumbnail, so every marker on this file fails to re-embed, which is
		// what an unmounted originals directory looks like from here.
		f := addMigrateTestFile(t, c, "1111111111111111111111111111111111111111", false)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.Error(t, err)
		assert.IsType(t, &FacesMigrateAbortedError{}, err)
		assert.Zero(t, result.Migrated)
		assert.False(t, result.ModelRecorded, "a refused run records no model, so it asks for no restart")

		// A run that did not replace the clusters must not move the setting either.
		assert.NotEqual(t, face.ModelFaceNet, savedFaceModel(t, c))

		// Nothing may have been replaced.
		assert.Zero(t, countFaceRows(t))

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.NotEmpty(t, stored.EmbeddingsJSON)
		assert.Equal(t, "Jane Doe", stored.MarkerName)
		assert.Equal(t, m.SubjUID, stored.SubjUID)
	})
	t.Run("ForceOverridesRefusal", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateforce")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "2222222222222222222222222222222222222222", false)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet, Force: true}, FacesMigrateResult{Target: face.ModelFaceNet})

		// Force finalizes, so the stale vector is cleared rather than kept.
		require.Error(t, err)
		assert.IsType(t, &FacesMigrateIncompleteError{}, err)
		assert.True(t, result.ModelRecorded, "an incomplete run that finalized still records the model")

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.Empty(t, stored.EmbeddingsJSON)
	})
	t.Run("LiftsRejectedMatches", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migraterejected")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "3333333333333333333333333333333333333333", true)
		rejected := addMigrateTestMarker(t, f.FileUID, entity.SrcAuto, "")
		require.NoError(t, entity.Db().Model(rejected).UpdateColumn("subj_src", entity.SrcManual).Error)
		require.True(t, entity.FindMarker(rejected.MarkerUID).RejectedMatch())

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Migrated)
		assert.Equal(t, 1, result.LiftedRejections)

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", rejected.MarkerUID).Error)
		assert.Equal(t, entity.SrcAuto, stored.SubjSrc)
		assert.False(t, stored.RejectedMatch())
	})
	t.Run("SameModelKeepsRejectedMatches", func(t *testing.T) {
		// A re-run to the model the library already uses re-crops markers without a sample extent,
		// and a rejection made under that model has to survive it.
		c := newMigrateTestConfig(t, "migraterejectedsame")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "6666666666666666666666666666666666666666", true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcAuto, "")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid = ?", m.MarkerUID).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "thumb_size": -1, "subj_src": entity.SrcManual}).Error)
		require.True(t, entity.FindMarker(m.MarkerUID).RejectedMatch())

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		require.Equal(t, 1, result.Migrated, "the marker is re-cropped")
		assert.Zero(t, result.LiftedRejections)

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.Equal(t, entity.SrcManual, stored.SubjSrc)
	})
	t.Run("FailedKeepsRejectedMatches", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migraterejectedfailed")
		w := NewFaces(c)

		ok := addMigrateTestFile(t, c, "8888888888888888888888888888888888888888", true)
		bad := addMigrateTestFile(t, c, "9999999999999999999999999999999999999999", false)
		embedded := addMigrateTestMarker(t, ok.FileUID, entity.SrcAuto, "")
		failed := addMigrateTestMarker(t, bad.FileUID, entity.SrcAuto, "")
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).Where("marker_uid IN (?)", []string{embedded.MarkerUID, failed.MarkerUID}).
			UpdateColumn("subj_src", entity.SrcManual).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet, Force: true}, FacesMigrateResult{Target: face.ModelFaceNet})

		var incomplete *FacesMigrateIncompleteError
		require.ErrorAs(t, err, &incomplete)
		assert.Equal(t, 1, result.FailedManual)
		assert.Equal(t, 1, result.LiftedRejections)

		for uid, want := range map[string]string{embedded.MarkerUID: entity.SrcAuto, failed.MarkerUID: entity.SrcManual} {
			stored := entity.Marker{}
			require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", uid).Error)
			assert.Equal(t, want, stored.SubjSrc, uid)
		}
	})
	t.Run("Success", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migratesuccess")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "3333333333333333333333333333333333333333", true)
		manual := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")
		auto := addMigrateTestMarker(t, f.FileUID, entity.SrcAuto, "Auto Person")
		xmp := addMigrateTestMarker(t, f.FileUID, entity.SrcXmp, "Xmp Person")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Equal(t, 3, result.Migrated)
		assert.Zero(t, result.Failed)

		// The vectors are the target's now, so the setting has to follow them: a start that
		// read the previous model would hide the whole library from matching.
		assert.Equal(t, face.ModelFaceNet, c.FaceModel())
		assert.Equal(t, face.ModelFaceNet, savedFaceModel(t, c))

		// The re-clustering at the end of the same run is gated on the block being cleared.
		assert.False(t, face.EmbeddingsBlocked())

		models, err := query.MarkerEmbeddingModels()
		require.NoError(t, err)
		require.Len(t, models, 1)
		assert.Equal(t, face.ModelFaceNet, models[0].EmbedModel)

		// Every identity must survive, whichever source recorded it.
		storedManual := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedManual, "marker_uid = ?", manual.MarkerUID).Error)
		assert.Equal(t, "Jane Doe", storedManual.MarkerName)
		assert.Equal(t, manual.SubjUID, storedManual.SubjUID)

		storedAuto := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedAuto, "marker_uid = ?", auto.MarkerUID).Error)
		assert.Equal(t, "Auto Person", storedAuto.MarkerName)

		storedXmp := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedXmp, "marker_uid = ?", xmp.MarkerUID).Error)
		assert.Equal(t, "Xmp Person", storedXmp.MarkerName)
		assert.NotEmpty(t, storedXmp.EmbeddingsJSON)
	})
	t.Run("KeepsAutomaticAssignments", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateassignments")
		useTestEmbedder(t, face.ModelSFace)
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "5555555555555555555555555555555555555555", true)
		manual := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")
		require.NotEmpty(t, manual.SubjUID)

		// A curated library holds one hand-named face per person and many the matcher
		// assigned, which is the sample set the replacement cluster has to keep.
		automatic := entity.Markers{
			*addMigrateTestSubjectMarker(t, f.FileUID, manual.SubjUID, "Jane Doe"),
			*addMigrateTestSubjectMarker(t, f.FileUID, manual.SubjUID, "Jane Doe"),
			*addMigrateTestSubjectMarker(t, f.FileUID, manual.SubjUID, "Jane Doe"),
		}

		plan := FacesMigratePlan{Target: face.ModelSFace}
		embedder := &sameFaceEmbedder{model: face.ModelSFace, dims: 128}
		result, err := w.migrate(context.Background(), plan, embedder,
			FacesMigrateOptions{Target: face.ModelSFace}, FacesMigrateResult{Target: face.ModelSFace})

		require.NoError(t, err)
		assert.Equal(t, 4, result.Migrated)
		assert.Zero(t, result.Failed)
		assert.Equal(t, 1, result.RebuiltSubjects)
		assert.Zero(t, result.AttentionSubjects)

		// The re-clustering and the audit at the end of this run read these, and a distance means
		// nothing outside the model it was measured in: asserted within the migrating process,
		// because a restart propagates them and would pass without the migration doing so.
		target := face.FindEmbeddingModel(face.ModelSFace)
		require.NotNil(t, target)
		assert.Equal(t, target.ClusterDist, face.ClusterDist)
		assert.Equal(t, target.ClusterRadius, face.ClusterRadius)
		assert.Equal(t, target.MatchDist, face.MatchDist)

		// Not one assignment may be dropped: losing them empties the person's page and no
		// later matching run brings them back.
		for _, m := range append(entity.Markers{*manual}, automatic...) {
			stored := entity.Marker{}
			require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
			assert.Equal(t, manual.SubjUID, stored.SubjUID, "marker %s lost its subject", m.MarkerUID)
			assert.Equal(t, "Jane Doe", stored.MarkerName)
			assert.NotEmpty(t, stored.FaceID, "marker %s was not linked to a cluster", m.MarkerUID)
		}

		// The rebuilt cluster has to carry all four samples, because a one-sample cluster
		// is too narrow to accept the faces the person already had.
		cluster := entity.Face{}
		require.NoError(t, entity.UnscopedDb().First(&cluster, "subj_uid = ?", manual.SubjUID).Error)
		assert.Equal(t, face.ModelSFace, cluster.EmbedModel)
		assert.Equal(t, 4, cluster.Samples)

		// A migration deletes every cluster and rebuilds it, so whatever the schema migration
		// raised is gone by then and each replacement has to classify itself. Left unset, the
		// person would be missing from the "face:1" search filter that reads this column.
		assert.Equal(t, int(face.RegularFace), cluster.FaceKind)
	})
	t.Run("RendersACropThumbnail", func(t *testing.T) {
		// A library indexed at a lower thumbnail limit has renditions too narrow for its face
		// crops, and the detail is in the original: the run renders what it needs rather than
		// embedding upscaled pixels, which nothing downstream could tell apart afterwards.
		c := newMigrateTestConfig(t, "migraterenders")
		w := NewFaces(c)

		cached, onDemand := thumb.SizeCached, thumb.SizeOnDemand
		t.Cleanup(func() { thumb.SizeCached, thumb.SizeOnDemand = cached, onDemand })
		thumb.SizeCached, thumb.SizeOnDemand = 720, 720

		f := newRenderTestFile(t, c, "4a4d444444444444444444444444444444444447", 2000, 1500)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		// 160/0.1 asks for 1600 px, which only a wider rendition than the cache holds can supply.
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", m.MarkerUID).
			UpdateColumns(entity.Values{"w": 0.1, "h": 0.1}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4},
			FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Equal(t, 1, result.Migrated)
		assert.Equal(t, 1, result.RenderedThumbs)
		assert.True(t, crop.CachedSizeExists(thumb.Sizes[thumb.Fit1920], f.FileHash, c.ThumbCachePath()),
			"the rendition the crop needs has to be written under the indexed hash")
	})
	t.Run("LegacyBlankModel", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migratelegacy")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "5555555555555555555555555555555555555555", true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		// Blank provenance means FaceNet, so migrating to FaceNet must skip this marker
		// and the finalize must not blank the vector it just declared valid.
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", m.MarkerUID).
			UpdateColumns(entity.Values{"embed_model": ""}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		result, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet}, FacesMigrateResult{Target: face.ModelFaceNet})

		require.NoError(t, err)
		assert.Zero(t, result.Migrated)
		assert.Equal(t, 1, result.Skipped)

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.NotEmpty(t, stored.EmbeddingsJSON)
	})
	t.Run("DetectorChangeKeepsUnfoundVectors", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migratedetector")
		w := NewFaces(c)

		detector := face.ActiveDetector()
		require.NotEmpty(t, detector)
		require.NotEqual(t, face.DetectorNone, detector, "the detector decides which crops are stale")

		f := addMigrateTestFile(t, c, "6666666666666666666666666666666666666666", true)
		recrop := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")
		stale := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "John Doe")

		// The first marker already holds a target-model vector and only needs a new crop; the
		// second holds one no run can compare. Detection finds neither in a blank thumbnail,
		// which is what tells them apart: only the second has nothing left to keep.
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", recrop.MarkerUID).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "detect_model": "some-other-detector"}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		opt := FacesMigrateOptions{Target: face.ModelFaceNet, Force: true}
		result, err := w.migrate(context.Background(), plan, &alignedEmbedder{oneHotEmbedder{dims: 4}}, opt, FacesMigrateResult{Target: face.ModelFaceNet})

		// The stale marker could not be re-embedded, so the run is incomplete: a retained
		// marker must not mask that, since only one of the two kept anything.
		var incomplete *FacesMigrateIncompleteError
		require.ErrorAs(t, err, &incomplete)
		assert.Equal(t, 1, incomplete.Failed)
		assert.Zero(t, result.Skipped, "a crop another detector placed is not up to date")
		assert.Equal(t, 1, result.Retained)
		assert.Equal(t, 1, result.Failed)

		kept := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&kept, "marker_uid = ?", recrop.MarkerUID).Error)
		assert.NotEmpty(t, kept.EmbeddingsJSON, "a marker detection cannot find again keeps its usable vector")

		lost := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&lost, "marker_uid = ?", stale.MarkerUID).Error)
		assert.Empty(t, lost.EmbeddingsJSON, "a vector of the previous model is cleared whether it could be replaced or not")
	})
	t.Run("MissingSampleExtentTerminates", func(t *testing.T) {
		// A library already on the target model is skipped whole without this condition, so the
		// column stays unset for good. The second run is the point: a sampling that cannot measure
		// one records the attempt, or a run would re-embed the same marker forever.
		c := newMigrateTestConfig(t, "migrateextent")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "6666666666666666666666666666666666666666", true)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", m.MarkerUID).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "thumb_size": -1}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		opt := FacesMigrateOptions{Target: face.ModelFaceNet}

		first, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		assert.Equal(t, 1, first.Migrated, "a marker with no recorded extent is re-embedded")

		second, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		assert.Zero(t, second.Migrated, "the second run has nothing left to do")
		assert.Equal(t, 1, second.Skipped)

		// The test file records no dimensions, so nothing can be measured from it - which is what
		// makes it the case that has to settle rather than repeat.
		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.Equal(t, entity.ThumbSizeUnmeasured, stored.ThumbSize)
		assert.NotEmpty(t, stored.EmbeddingsJSON)
	})
	t.Run("FailedReembeddingTerminates", func(t *testing.T) {
		// The other half of termination: a marker the sampling reached and could not re-embed keeps
		// its vector, so nothing records an extent for it and it would be stale on every run. The
		// file has no thumb, which is how this harness makes every re-embedding on it fail.
		c := newMigrateTestConfig(t, "migrateunreadable")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "7777777777777777777777777777777777777777", false)
		m := addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", m.MarkerUID).
			UpdateColumns(entity.Values{"embed_model": face.ModelFaceNet, "thumb_size": -1}).Error)

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		opt := FacesMigrateOptions{Target: face.ModelFaceNet}

		_, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)

		assert.NotEmpty(t, stored.EmbeddingsJSON, "a failed re-embedding must keep the vector it holds")
		assert.True(t, stored.ThumbSizeSettled(), "and must not be sampled again on every future run")

		second, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		assert.Equal(t, 1, second.Skipped, "the second run has nothing left to attempt")
	})
	t.Run("Idempotent", func(t *testing.T) {
		c := newMigrateTestConfig(t, "migrateidempotent")
		w := NewFaces(c)

		f := addMigrateTestFile(t, c, "4444444444444444444444444444444444444444", true)
		addMigrateTestMarker(t, f.FileUID, entity.SrcManual, "Jane Doe")

		plan := FacesMigratePlan{Target: face.ModelFaceNet}
		opt := FacesMigrateOptions{Target: face.ModelFaceNet}

		first, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		require.Equal(t, 1, first.Migrated)

		second, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, opt, FacesMigrateResult{Target: face.ModelFaceNet})
		require.NoError(t, err)
		assert.Zero(t, second.Migrated)
		assert.Equal(t, 1, second.Skipped)
		assert.Zero(t, second.Failed)
	})
}
