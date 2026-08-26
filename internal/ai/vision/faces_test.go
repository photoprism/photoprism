package vision

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// mustEndpoint returns the face model endpoint URL configured for the test, so a case that
// depends on one is skipped rather than passing for the wrong reason.
func mustEndpoint(t *testing.T) string {
	t.Helper()

	uri, method := Config.Model(ModelTypeFace).Endpoint()

	if method == "" {
		t.Skip("vision: skipping, no endpoint method is configured")
	}

	return uri
}

func TestDetectFaces(t *testing.T) {
	fileName, err := filepath.Abs(filepath.Join("..", "face", "testdata", "1.jpg"))
	require.NoError(t, err)

	origConfig := Config
	t.Cleanup(func() { Config = origConfig })

	Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}}

	modelsPath, err := filepath.Abs(filepath.Join("..", "..", "..", "assets", "models"))
	require.NoError(t, err)

	detectorPath := face.DefaultDetector().Path(modelsPath)

	if _, statErr := os.Stat(detectorPath); statErr != nil {
		t.Skipf("faces: skipping, %s is not available", filepath.Base(detectorPath))
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

		result, detectErr := DetectFaces(fileName, 20, 0, false, 0)

		require.NoError(t, detectErr)

		if len(result) == 0 {
			t.Skip("faces: skipping, the detector found no face to record")
		}

		assert.True(t, result[0].Embeddings.Empty(), "a paused instance must not embed")
	})
	t.Run("PausedIgnoresTheEndpoint", func(t *testing.T) {
		// A configured endpoint is no exemption: its vectors are stamped with the model this
		// instance is configured for, so they would land in the same second space.
		t.Cleanup(func() {
			Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}}
			face.UnblockEmbeddings()
		})

		Config = &ConfigValues{Models: Models{{
			Name:    "facenet",
			Type:    ModelTypeFace,
			Service: Service{Uri: "http://127.0.0.1:1/vision/face", Method: "POST"},
		}}}

		require.NotEmpty(t, mustEndpoint(t))

		face.BlockEmbeddings("12 marker(s) use sface, but this instance is configured for facenet")

		result, detectErr := DetectFaces(fileName, 20, 0, false, 0)

		// An endpoint that was called would fail against a closed port, so no error is what
		// proves it was not.
		require.NoError(t, detectErr)

		if len(result) > 0 {
			assert.True(t, result[0].Embeddings.Empty())
		}
	})
	t.Run("MissingFilename", func(t *testing.T) {
		_, detectErr := DetectFaces("", 20, 0, false, 0)
		require.Error(t, detectErr)
	})
	t.Run("NoFaceModel", func(t *testing.T) {
		Config = &ConfigValues{Models: Models{}}
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		_, detectErr := DetectFaces(fileName, 20, 0, false, 0)
		require.Error(t, detectErr)
	})
}
