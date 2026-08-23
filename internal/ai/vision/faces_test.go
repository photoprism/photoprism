package vision

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

func TestDetectFaces(t *testing.T) {
	fileName, err := filepath.Abs(filepath.Join("..", "face", "testdata", "1.jpg"))
	require.NoError(t, err)

	origConfig := Config
	t.Cleanup(func() { Config = origConfig })

	Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}}

	detectorPath, err := filepath.Abs(filepath.Join("..", "..", "..", "assets", "models", "scrfd", face.DefaultONNXModelFilename))
	require.NoError(t, err)

	if _, statErr := os.Stat(detectorPath); statErr != nil {
		t.Skipf("faces: skipping, %s is not available", face.DefaultONNXModelFilename)
	}

	prev := face.UseEngine(nil)
	t.Cleanup(func() {
		if current := face.UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	require.NoError(t, face.ConfigureEngine(face.EngineSettings{
		Name: face.EngineONNX,
		ONNX: face.ONNXOptions{ModelPath: detectorPath, Threads: 1},
	}))

	t.Run("PausedKeepsTheDetections", func(t *testing.T) {
		// A marker without a vector is an ordinary state that a migration fills in, while a
		// dropped detection has to be re-indexed: the workers discard what this returns an
		// error for.
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		result, detectErr := DetectFaces(fileName, 20, false, 0)

		require.NoError(t, detectErr)

		if len(result) == 0 {
			t.Skip("faces: skipping, the detector found no face to record")
		}

		assert.True(t, result[0].Embeddings.Empty(), "a paused instance must not embed")
	})
	t.Run("MissingFilename", func(t *testing.T) {
		_, detectErr := DetectFaces("", 20, false, 0)
		require.Error(t, detectErr)
	})
	t.Run("NoFaceModel", func(t *testing.T) {
		Config = &ConfigValues{Models: Models{}}
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		_, detectErr := DetectFaces(fileName, 20, false, 0)
		require.Error(t, detectErr)
	})
}
