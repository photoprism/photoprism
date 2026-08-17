package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// restoreEmbedderSettings returns the settings that reinstate the currently configured
// embedder. ONNX models load from an explicit file, so restoring by name alone would
// leave the package without an embedder and fail every later test that expects one.
func restoreEmbedderSettings(c *config.Config) face.EmbedderSettings {
	return face.EmbedderSettings{
		Name:      face.ConfiguredModel(),
		Model:     face.FindEmbeddingModel(face.ConfiguredModel()),
		ModelPath: c.FaceModelPath(),
		Threads:   c.FaceEngineThreads(),
	}
}

// useTestEmbedder configures a working embedding model for the duration of a test, so
// clustering does not depend on whether the ONNX runtime is present in the environment.
func useTestEmbedder(t *testing.T, c *config.Config, name face.ModelName) {
	t.Helper()

	restore := restoreEmbedderSettings(c)
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: name}))

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(restore)
	})
}

func TestFaces_Cluster(t *testing.T) {
	t.Run("ForceTrue", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, c, face.ModelSFace)

		m := NewFaces(c)

		opt := FacesOptions{
			Force:     true,
			Threshold: 1,
		}

		r, err := m.Cluster(opt)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})
	t.Run("ForceFalse", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, c, face.ModelSFace)

		m := NewFaces(c)

		opt := FacesOptions{
			Force:     false,
			Threshold: 1,
		}

		r, err := m.Cluster(opt)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})
	t.Run("RefusesWhenEmbedderFailed", func(t *testing.T) {
		c := config.TestConfig()

		restore := restoreEmbedderSettings(c)
		require.Error(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace, Model: face.FindEmbeddingModel(face.ModelSFace)}))
		t.Cleanup(func() {
			_ = face.ConfigureEmbedder(restore)
		})

		r, err := NewFaces(c).Cluster(FacesOptions{Force: true, Threshold: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding model failed to load")
		assert.Empty(t, r)
	})
	t.Run("RefusesMixedDimensions", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, c, face.ModelSFace)

		// No fixture records an embedding model, so stamping these two selects exactly
		// them and keeps the mixed set under the test's control.
		for i, values := range []face.Embeddings{{{0.1, 0.2}}, {{0.3, 0.4, 0.5}}} {
			m := entity.Marker{
				MarkerUID:      rnd.GenerateUID('m'),
				MarkerType:     entity.MarkerFace,
				MarkerSrc:      entity.SrcImage,
				Size:           100,
				Score:          50,
				EmbeddingsJSON: values.JSON(),
				EmbedModel:     face.ModelSFace,
			}

			require.NoError(t, entity.Db().Create(&m).Error, "marker %d", i)

			t.Cleanup(func() {
				entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID)
			})
		}

		r, err := NewFaces(c).Cluster(FacesOptions{Force: true, Threshold: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "different lengths")
		assert.Empty(t, r)
	})
}
