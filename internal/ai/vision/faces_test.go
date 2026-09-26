package vision

import (
	"net/http"
	"net/http/httptest"
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

		result, detectErr := DetectFaces(fileName, 20, 0, false, 0, nil)

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

		result, detectErr := DetectFaces(fileName, 20, 0, false, 0, nil)

		// An endpoint that was called would fail against a closed port, so no error is what
		// proves it was not.
		require.NoError(t, detectErr)

		if len(result) > 0 {
			assert.True(t, result[0].Embeddings.Empty())
		}
	})
	t.Run("RendersTheCropSourceBeforeEmbedding", func(t *testing.T) {
		// Both embedding paths select the rendition they crop from by statting the cache, so one
		// rendered afterwards is one no vector was drawn from.
		var detected face.Faces
		var embedded bool

		result, detectErr := DetectFaces(fileName, 20, 0, false, 0, func(faces face.Faces) {
			detected = faces
			embedded = !faces[0].Embeddings.Empty()
		})

		require.NoError(t, detectErr)

		if len(result) == 0 {
			t.Skip("faces: skipping, the detector found no face to render for")
		}

		require.Len(t, detected, len(result), "the detections decide how wide the rendition has to be")
		assert.False(t, embedded, "the crops must not have been taken yet")
	})
	t.Run("PausedRendersNothing", func(t *testing.T) {
		// A paused instance takes no crop, so rendering one would be work for nobody.
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		called := false

		_, detectErr := DetectFaces(fileName, 20, 0, false, 0, func(faces face.Faces) { called = true })

		require.NoError(t, detectErr)
		assert.False(t, called)
	})
	t.Run("DisabledEmbeddingsRenderNothing", func(t *testing.T) {
		// The same for an instance configured to embed nothing at all.
		prev := face.ConfiguredModel()
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
		t.Cleanup(func() {
			require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: prev}))
		})

		called := false

		_, detectErr := DetectFaces(fileName, 20, 0, false, 0, func(faces face.Faces) { called = true })

		require.NoError(t, detectErr)
		assert.False(t, called)
	})
	t.Run("MissingFilename", func(t *testing.T) {
		_, detectErr := DetectFaces("", 20, 0, false, 0, nil)
		require.Error(t, detectErr)
	})
	t.Run("NoFaceModel", func(t *testing.T) {
		Config = &ConfigValues{Models: Models{}}
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		_, detectErr := DetectFaces(fileName, 20, 0, false, 0, nil)
		require.Error(t, detectErr)
	})
}

func TestEmbedFaces(t *testing.T) {
	fileName, err := filepath.Abs(filepath.Join("..", "face", "testdata", "1.jpg"))
	require.NoError(t, err)

	origConfig := Config
	t.Cleanup(func() { Config = origConfig })

	Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}}

	faces := face.Faces{{Rows: 100, Cols: 100, Area: face.NewArea("face", 50, 50, 20)}}

	t.Run("NoFaces", func(t *testing.T) {
		called := false

		require.NoError(t, EmbedFaces(fileName, nil, false, func(face.Faces) { called = true }))
		assert.False(t, called)
	})
	t.Run("Paused", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet, but this instance is configured for sface")

		called := false

		require.NoError(t, EmbedFaces(fileName, faces, false, func(face.Faces) { called = true }))
		assert.False(t, called, "a paused instance must not render a crop source")
		assert.True(t, faces[0].Embeddings.Empty())
	})
	t.Run("MissingFilename", func(t *testing.T) {
		require.Error(t, EmbedFaces("", faces, false, nil))
	})
	t.Run("NotConfigured", func(t *testing.T) {
		Config = nil
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		require.Error(t, EmbedFaces(fileName, faces, false, nil))
	})
	t.Run("NoFaceModel", func(t *testing.T) {
		Config = &ConfigValues{Models: Models{}}
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		require.Error(t, EmbedFaces(fileName, faces, false, nil))
	})
	t.Run("EndpointRefused", func(t *testing.T) {
		// A service that refuses the request returns an error, so no markers are updated.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code":403,"error":"Forbidden","result":{}}`))
		}))
		defer server.Close()

		Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace, Service: Service{
			Uri: server.URL, Method: http.MethodPost, RequestFormat: ApiFormatVision, ResponseFormat: ApiFormatVision,
		}}}}
		t.Cleanup(func() { Config = &ConfigValues{Models: Models{{Name: "facenet", Type: ModelTypeFace}}} })

		// The crops are cached next to a copy of the image, which the request reads them from.
		tmpFile := filepath.Join(t.TempDir(), "1.jpg")
		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		require.NoError(t, os.WriteFile(tmpFile, data, 0o600))

		refused := face.Faces{{Rows: 100, Cols: 100, Area: face.NewArea("face", 50, 50, 20)}}
		require.EqualError(t, EmbedFaces(tmpFile, refused, true, nil), "Forbidden (status code 403)")
		assert.True(t, refused[0].Embeddings.Empty())
	})
}
