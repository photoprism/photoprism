package photoprism

import (
	"context"
	"image"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/thumb"
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

// newMigrateTestConfig returns an isolated config for a migration test.
func newMigrateTestConfig(t *testing.T, name string) *config.Config {
	t.Helper()
	useTestDb(t, name)

	oldConfig := Config()

	// Initializing a config reconfigures the process-wide embedder from its models path,
	// which is empty here, so the previous one has to be reinstated afterwards.
	restore := restoreEmbedderSettings(oldConfig)

	c := config.NewMinimalTestConfigWithDb(name, filepath.Join(t.TempDir(), "storage"))
	require.NoError(t, c.CreateDirectories())

	SetConfig(c)
	t.Cleanup(func() {
		SetConfig(oldConfig)
		oldConfig.RegisterDb()
		_ = face.ConfigureEmbedder(restore)
	})

	// The isolated database is seeded from the cached test database, so migration would
	// otherwise walk the shared fixtures instead of the rows this test creates.
	for _, model := range []any{&entity.Marker{}, &entity.Face{}, &entity.Subject{}, &entity.File{}} {
		require.NoError(t, entity.UnscopedDb().Unscoped().Delete(model).Error)
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

// addMigrateTestMarker creates a face marker with a stale embedding on the given file.
func addMigrateTestMarker(t *testing.T, fileUID, subjSrc, name string) *entity.Marker {
	t.Helper()

	m := &entity.Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		FileUID:        fileUID,
		MarkerType:     entity.MarkerFace,
		MarkerSrc:      entity.SrcImage,
		Size:           100,
		Score:          50,
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

// countFaceRows returns the number of stored face clusters.
func countFaceRows(t *testing.T) int {
	t.Helper()

	var n int
	require.NoError(t, entity.UnscopedDb().Model(&entity.Face{}).Count(&n).Error)

	return n
}

func TestFinalizeRefused(t *testing.T) {
	t.Run("NothingAttempted", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(0, 0))
	})
	t.Run("AllFailed", func(t *testing.T) {
		assert.Contains(t, finalizeRefused(0, 15), "none of 15")
	})
	t.Run("AboveRatio", func(t *testing.T) {
		assert.Contains(t, finalizeRefused(1, 9), "9 of 10")
	})
	t.Run("WithinRatio", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(99, 1))
	})
	t.Run("NoFailures", func(t *testing.T) {
		assert.Empty(t, finalizeRefused(10, 0))
	})
}

func TestFaces_migrate(t *testing.T) {
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
		_, err := w.migrate(context.Background(), plan, &oneHotEmbedder{dims: 4}, FacesMigrateOptions{Target: face.ModelFaceNet, Force: true}, FacesMigrateResult{Target: face.ModelFaceNet})

		// Force finalizes, so the stale vector is cleared rather than kept.
		require.Error(t, err)
		assert.IsType(t, &FacesMigrateIncompleteError{}, err)

		stored := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&stored, "marker_uid = ?", m.MarkerUID).Error)
		assert.Empty(t, stored.EmbeddingsJSON)
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

		models, err := query.MarkerEmbeddingModels()
		require.NoError(t, err)
		require.Len(t, models, 1)
		assert.Equal(t, face.ModelFaceNet, models[0].EmbedModel)

		// The manual identity must survive, and the automatic one must be cleared.
		storedManual := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedManual, "marker_uid = ?", manual.MarkerUID).Error)
		assert.Equal(t, "Jane Doe", storedManual.MarkerName)
		assert.Equal(t, manual.SubjUID, storedManual.SubjUID)

		storedAuto := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedAuto, "marker_uid = ?", auto.MarkerUID).Error)
		assert.Empty(t, storedAuto.MarkerName)

		storedXmp := entity.Marker{}
		require.NoError(t, entity.UnscopedDb().First(&storedXmp, "marker_uid = ?", xmp.MarkerUID).Error)
		assert.Equal(t, "Xmp Person", storedXmp.MarkerName)
		assert.NotEmpty(t, storedXmp.EmbeddingsJSON)
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
