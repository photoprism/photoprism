package photoprism

import (
	"context"
	"image"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
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

// Run returns a deterministic test embedding.
func (e *migrationTestEmbedder) Run(image.Image) face.Embeddings {
	return face.Embeddings{make(face.Embedding, e.dims)}
}

// Close releases the test embedder.
func (e *migrationTestEmbedder) Close() error { return nil }

func TestFacesMigrateIncompleteError_Error(t *testing.T) {
	err := (&FacesMigrateIncompleteError{Failed: 3}).Error()
	assert.Contains(t, err, "3 failed")
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
}

func TestFaces_MigrateDryRun(t *testing.T) {
	w := NewFaces(config.TestConfig())
	result, err := w.Migrate(context.Background(), FacesMigrateOptions{DryRun: true})
	require.NoError(t, err)
	assert.Equal(t, face.NormalizeModelName(w.conf.FaceModel()), result.Target)
	assert.Equal(t, 0, result.Migrated)

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

func TestFaces_migrateFaceFile(t *testing.T) {
	embedder := &migrationTestEmbedder{name: face.ModelFaceNet, dims: 512}
	w := NewFaces(config.TestConfig())

	migrated, skipped, failed, detected, err := w.migrateFaceFile(embedder, face.ModelFaceNet, "missing-file")
	require.NoError(t, err)
	assert.Zero(t, migrated)
	assert.Zero(t, skipped)
	assert.Empty(t, failed)
	assert.False(t, detected)

	_, _, _, _, err = w.migrateFaceFile(embedder, face.ModelFaceNet, "")
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

	_, _, failed, _, err = w.migrateFaceFile(embedder, face.ModelFaceNet, marker.FileUID)
	require.Error(t, err)
	assert.Equal(t, []string{marker.MarkerUID}, failed)
}

func TestFaceMigrationMarkerUIDs(t *testing.T) {
	markers := entity.Markers{{MarkerUID: "m1"}, {MarkerUID: "m2"}}
	assert.Equal(t, []string{"m1", "m2"}, faceMigrationMarkerUIDs(markers))
	assert.Empty(t, faceMigrationMarkerUIDs(nil))
}

func TestFaces_cropMigrationEmbeddings(t *testing.T) {
	embedder := &migrationTestEmbedder{name: face.ModelFaceNet, dims: 4}
	w := NewFaces(config.TestConfig())
	file := &entity.File{FileHash: "0123456789012345678901234567890123456789", FileName: "missing.jpg", FileRoot: entity.RootOriginals}
	markers := entity.Markers{{MarkerUID: "m1", MarkerType: entity.MarkerFace, W: 0.5, H: 0.5}}

	result, err := w.cropMigrationEmbeddings(embedder, file, markers)
	require.NoError(t, err)
	assert.Empty(t, result)

	_, err = w.cropMigrationEmbeddings(embedder, nil, markers)
	require.Error(t, err)

	successConf := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, successConf.CreateDirectories())
	successWorker := NewFaces(successConf)
	hash := "0123456789012345678901234567890123456789"
	thumbName, err := thumb.Sizes[thumb.Fit720].FileName(hash, successConf.ThumbCachePath())
	require.NoError(t, err)
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 64, 64)), thumbName))
	result, err = successWorker.cropMigrationEmbeddings(
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

	result, err := w.detectMigrationEmbeddings(embedder, nil, nil, nil)
	require.Error(t, err)
	assert.Empty(t, result)
}

func TestMigrationDetectionThumb(t *testing.T) {
	thumbPath := t.TempDir()
	hash := "0123456789012345678901234567890123456789"
	name, err := thumb.Sizes[thumb.Fit720].FileName(hash, thumbPath)
	require.NoError(t, err)
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 8, 8)), name))

	result, err := migrationDetectionThumb(thumbPath, &entity.File{FileHash: hash})
	require.NoError(t, err)
	assert.Equal(t, name, result)

	_, err = migrationDetectionThumb(thumbPath, nil)
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
	assert.Equal(t, detected[0].Area, result["m1"].Area)
	assert.Equal(t, detected[1].Area, result["m2"].Area)
	assert.Empty(t, matchMigrationDetections(nil, detected))
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
	marker := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        "fs6sg6bw45bnlqdw",
		MarkerType:     entity.MarkerFace,
		SubjUID:        subjectUID,
		SubjSrc:        entity.SrcManual,
		EmbedModel:     target,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		W:              0.1,
		H:              0.1,
	}
	require.NoError(t, entity.Db().Create(marker).Error)
	t.Cleanup(func() { entity.UnscopedDb().Delete(marker) })

	result, rebuilt, err := buildFaceMigrationClusters(target)
	require.NoError(t, err)
	assert.Equal(t, rebuilt, len(result))
	found := false
	for _, cluster := range result {
		found = found || cluster.Face.SubjUID == subjectUID
	}
	assert.True(t, found)

	_, _, err = buildFaceMigrationClusters("unknown")
	require.Error(t, err)
}

func TestMigrationCanceled(t *testing.T) {
	assert.NoError(t, migrationCanceled(context.Background(), nil))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, migrationCanceled(ctx, nil))
}
