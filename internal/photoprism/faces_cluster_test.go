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

// baseEmbedder holds the embedder that the package test config installed.
var baseEmbedder face.EmbedderSettings

// captureEmbedderSettings records the embedder installed by the package test config, so
// a test that replaces the process-wide embedder has a value known to be good to put
// back. ONNX models load from an explicit file, so restoring by name alone would leave
// the package without an embedder and fail every later test that expects one.
func captureEmbedderSettings(c *config.Config) {
	baseEmbedder = face.EmbedderSettings{
		Name:      face.ConfiguredModel(),
		Model:     face.FindEmbeddingModel(face.ConfiguredModel()),
		ModelPath: c.FaceModelPath(),
		Threads:   c.FaceModelThreads(),
	}
}

// restoreEmbedder reinstates the embedder that the package test config installed.
//
// Reading the name back from the global would capture whatever an earlier test left
// there, so one missed restore would spread to every test that follows. A restore that
// does not take fails the test for the same reason.
func restoreEmbedder(t *testing.T) {
	t.Helper()

	assert.NoError(t, face.ConfigureEmbedder(baseEmbedder))
	assert.Equal(t, baseEmbedder.Name, face.ConfiguredModel())
}

// useTestEmbedder configures a working embedding model for the duration of a test, so
// clustering does not depend on whether the ONNX runtime is present in the environment.
func useTestEmbedder(t *testing.T, name face.ModelName) {
	t.Helper()

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: name}))

	t.Cleanup(func() {
		restoreEmbedder(t)
	})
}

func TestFaces_Cluster(t *testing.T) {
	t.Run("ForceTrue", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, face.ModelSFace)

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
		useTestEmbedder(t, face.ModelSFace)

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

		require.Error(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace, Model: face.FindEmbeddingModel(face.ModelSFace)}))
		t.Cleanup(func() {
			restoreEmbedder(t)
		})

		r, err := NewFaces(c).Cluster(FacesOptions{Force: true, Threshold: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding model failed to load")
		assert.Empty(t, r)
	})
	t.Run("RefusesMixedDimensions", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, face.ModelSFace)

		// No fixture records an embedding model, so stamping these two selects exactly
		// them and keeps the mixed set under the test's control.
		for i, values := range []face.Embeddings{{{0.1, 0.2}}, {{0.3, 0.4, 0.5}}} {
			m := entity.Marker{
				MarkerUID:      rnd.GenerateUID('m'),
				MarkerType:     entity.MarkerFace,
				MarkerSrc:      entity.SrcImage,
				Size:           100,
				Score:          face.ClusterScore("") + 10,
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
