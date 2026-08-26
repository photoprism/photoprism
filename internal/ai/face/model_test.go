package face

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
)

var assetsModelsPath, _ = filepath.Abs("../../../assets/models")
var modelPath = filepath.Join(assetsModelsPath, "facenet")

// detectorModelPath names the detector a build ships rather than a fixed directory, so the
// detector-dependent tests keep running against whatever "make dep-models" installs.
var detectorModelPath = DefaultDetector().Path(assetsModelsPath)

func TestNet(t *testing.T) {
	prev := UseEngine(nil)
	t.Cleanup(func() {
		current := UseEngine(prev)
		if current != nil {
			_ = current.Close()
		}
	})

	err := ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{
			ModelPath: detectorModelPath,
			Threads:   1,
		},
	})
	if err != nil {
		// Only a missing model may skip this test, as a detector that fails to
		// initialize despite being present indicates a broken ONNX Runtime or a
		// binding that does not match the installed library version.
		if _, statErr := os.Stat(detectorModelPath); statErr != nil {
			t.Skipf("faces: skipping detector-dependent test, %s is not available", filepath.Base(detectorModelPath))
		}

		t.Fatalf("faces: failed to initialize detector: %s", err)
	}
	require.Equal(t, EngineONNX, ActiveEngineName())

	faceNet := NewModel(ModelFaceNet, modelPath, "testdata/cache", 160, nil, false)
	detectedFiles := 0
	embeddedFaces := 0

	if err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || filepath.Base(filepath.Dir(fileName)) != "testdata" {
			return nil
		}

		t.Run(fileName, func(t *testing.T) {
			baseName := filepath.Base(fileName)

			faces, err := faceNet.Detect(fileName, 20, false, -1)

			if err != nil {
				t.Fatal(err)
			}

			if len(faces) > 0 {
				detectedFiles++
			}

			for i, f := range faces {
				if len(f.Embeddings) == 0 {
					continue
				}

				embeddedFaces++
				magnitude := f.Embeddings[0].Magnitude()
				assert.InDeltaf(t, 1.0, magnitude, 0.02, "embedding %d in %s should stay normalized", i, baseName)
			}
		})

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	assert.Greater(t, detectedFiles, 0)
	assert.Greater(t, embeddedFaces, 0)
}

func TestModelDims(t *testing.T) {
	t.Run("FromGraphMetadata", func(t *testing.T) {
		meta := &tensorflow.ModelInfo{Output: &tensorflow.ModelOutput{NumOutputs: 256}}
		assert.Equal(t, 256, modelDims(ModelFaceNet, meta))
	})
	t.Run("FromRegistry", func(t *testing.T) {
		assert.Equal(t, 512, modelDims(ModelFaceNet, &tensorflow.ModelInfo{}))
	})
	t.Run("UnknownModel", func(t *testing.T) {
		assert.Equal(t, len(NullEmbedding), modelDims("custom", &tensorflow.ModelInfo{}))
	})
}

func TestModel_Embedder(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		m := NewModel("", modelPath, "testdata/cache", 0, nil, true)
		width, height := m.CropSize()
		assert.Equal(t, ModelFaceNet, m.ModelName())
		assert.Equal(t, 512, m.Dims())
		assert.Equal(t, CropSize.Width, width)
		assert.Equal(t, CropSize.Height, height)
		assert.False(t, m.Aligned())
	})
	t.Run("CustomModel", func(t *testing.T) {
		meta := &tensorflow.ModelInfo{Output: &tensorflow.ModelOutput{NumOutputs: 128}}
		m := NewModel("custom", modelPath, "testdata/cache", 160, meta, true)
		assert.Equal(t, ModelName("custom"), m.ModelName())
		assert.Equal(t, 128, m.Dims())
	})
	t.Run("CloseWithoutSession", func(t *testing.T) {
		m := NewModel(ModelFaceNet, modelPath, "testdata/cache", 160, nil, true)
		require.NoError(t, m.Close())
	})
	t.Run("ImplementsEmbedder", func(t *testing.T) {
		var embedder Embedder = NewModel(ModelFaceNet, modelPath, "testdata/cache", 160, nil, true)
		assert.Equal(t, ModelFaceNet, embedder.ModelName())
	})
}
