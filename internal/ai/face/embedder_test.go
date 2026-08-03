package face

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEmbedder is a stand-in Embedder that records whether it has been closed.
type testEmbedder struct {
	name   ModelName
	dims   int
	closed bool
}

// ModelName returns the configured model name.
func (e *testEmbedder) ModelName() ModelName { return e.name }

// Dims returns the configured embedding length.
func (e *testEmbedder) Dims() int { return e.dims }

// CropSize returns a fixed input size.
func (e *testEmbedder) CropSize() (int, int) { return 112, 112 }

// Aligned reports that the test embedder wants aligned crops.
func (e *testEmbedder) Aligned() bool { return true }

// Run returns a single embedding of the configured length.
func (e *testEmbedder) Run(img image.Image) Embeddings {
	if img == nil {
		return nil
	}

	values := make([]float32, e.dims)
	values[0] = 1

	return NewEmbeddings([][]float32{values})
}

// Close marks the embedder as closed.
func (e *testEmbedder) Close() error {
	e.closed = true
	return nil
}

// restoreEmbedder resets the active embedder and configured model after a test.
func restoreEmbedder(t *testing.T) {
	prev := ActiveEmbedder()
	prevName := ConfiguredModel()

	t.Cleanup(func() {
		embedderMu.Lock()
		activeEmbedder = prev
		configuredModel = prevName
		embedderMu.Unlock()
	})
}

func TestUseEmbedder(t *testing.T) {
	restoreEmbedder(t)

	t.Run("ReturnsPrevious", func(t *testing.T) {
		first := &testEmbedder{name: ModelSFace, dims: 128}
		second := &testEmbedder{name: ModelArcFaceR50, dims: 512}
		UseEmbedder(first)
		assert.Equal(t, first, UseEmbedder(second))
		assert.Equal(t, second, ActiveEmbedder())
	})
	t.Run("Clears", func(t *testing.T) {
		UseEmbedder(nil)
		assert.Nil(t, ActiveEmbedder())
	})
}

func TestConfiguredModel(t *testing.T) {
	restoreEmbedder(t)

	t.Run("Normalized", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: "FaceNet", Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.Equal(t, ModelFaceNet, ConfiguredModel())
	})
	t.Run("DerivedFromModel", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.Equal(t, ModelFaceNet, ConfiguredModel())
	})
	t.Run("NoModel", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{}))
		assert.Equal(t, ModelNone, ConfiguredModel())
	})
}

func TestEmbeddingsDisabled(t *testing.T) {
	restoreEmbedder(t)

	t.Run("None", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelNone}))
		assert.True(t, EmbeddingsDisabled())
	})
	t.Run("FaceNet", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.False(t, EmbeddingsDisabled())
	})
}

func TestExpectedDims(t *testing.T) {
	restoreEmbedder(t)

	t.Run("FromActiveEmbedder", func(t *testing.T) {
		UseEmbedder(&testEmbedder{name: ModelSFace, dims: 128})
		assert.Equal(t, 128, ExpectedDims())
	})
	t.Run("FromConfiguredModel", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.Equal(t, 512, ExpectedDims())
	})
	t.Run("Default", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelNone}))
		assert.Equal(t, len(NullEmbedding), ExpectedDims())
	})
}

func TestConfigureEmbedder(t *testing.T) {
	restoreEmbedder(t)

	t.Run("TensorFlowModelStaysWithVision", func(t *testing.T) {
		// FaceNet is instantiated by the vision subsystem, so no embedder is installed.
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.Nil(t, ActiveEmbedder())
	})
	t.Run("ClosesPrevious", func(t *testing.T) {
		prev := &testEmbedder{name: ModelSFace, dims: 128}
		UseEmbedder(prev)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelNone}))
		assert.True(t, prev.closed)
	})
	t.Run("MissingONNXPath", func(t *testing.T) {
		// A failed initialization must not leave the requested name configured: callers
		// fall back to another runtime and would record its vectors under this model.
		err := ConfigureEmbedder(EmbedderSettings{Name: ModelSFace, Model: FindEmbeddingModel(ModelSFace)})
		require.Error(t, err)
		assert.Nil(t, ActiveEmbedder())
		assert.Equal(t, ModelNone, ConfiguredModel())
		assert.Equal(t, "", EmbeddingModelName())
		assert.True(t, EmbeddingsDisabled())
		assert.True(t, SamplesComparable())
	})
}

func TestEmbeddingModelName(t *testing.T) {
	restoreEmbedder(t)

	t.Run("FromActiveEmbedder", func(t *testing.T) {
		UseEmbedder(&testEmbedder{name: ModelSFace, dims: 128})
		assert.Equal(t, ModelSFace, EmbeddingModelName())
	})
	t.Run("FromConfiguredModel", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.Equal(t, ModelFaceNet, EmbeddingModelName())
	})
	t.Run("UnknownStaysEmpty", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelNone}))
		assert.Equal(t, "", EmbeddingModelName())
	})
}

func TestSamplesComparable(t *testing.T) {
	restoreEmbedder(t)

	t.Run("FaceNet", func(t *testing.T) {
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.True(t, SamplesComparable())
	})
	t.Run("Unknown", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelNone}))
		assert.True(t, SamplesComparable())
	})
	t.Run("OtherModelSameDims", func(t *testing.T) {
		// ArcFace also returns 512 values, so only the model name can rule it out.
		UseEmbedder(&testEmbedder{name: ModelArcFaceR50, dims: 512})
		assert.False(t, SamplesComparable())
	})
	t.Run("OtherModelOtherDims", func(t *testing.T) {
		UseEmbedder(&testEmbedder{name: ModelSFace, dims: 128})
		assert.False(t, SamplesComparable())
	})
}

func TestEmbeddingSamplesGatedByModel(t *testing.T) {
	restoreEmbedder(t)

	prevChildren, prevBackground := SkipChildren, IgnoreBackground
	SkipChildren, IgnoreBackground = true, true

	t.Cleanup(func() { SkipChildren, IgnoreBackground = prevChildren, prevBackground })

	require.NotEmpty(t, Children)
	require.NotEmpty(t, Background)

	child := make(Embedding, len(Children[0].Embedding))
	copy(child, Children[0].Embedding)

	background := make(Embedding, len(Background[0].Embedding))
	copy(background, Background[0].Embedding)

	t.Run("AppliesToFaceNet", func(t *testing.T) {
		UseEmbedder(nil)
		require.NoError(t, ConfigureEmbedder(EmbedderSettings{Name: ModelFaceNet, Model: FindEmbeddingModel(ModelFaceNet)}))
		assert.True(t, child.IsChild())
		assert.True(t, background.IsBackground())
	})
	t.Run("InactiveForOtherModel", func(t *testing.T) {
		// The same 512-value vectors must no longer be judged by FaceNet-space samples.
		UseEmbedder(&testEmbedder{name: ModelArcFaceR50, dims: 512})
		assert.False(t, child.IsChild())
		assert.False(t, background.IsBackground())
	})
}
