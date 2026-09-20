package face

import (
	"image"
	"image/color"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
)

// newProviderTestEmbedder returns an embedder for the named model on the given execution
// provider, or skips the test when its weights have not been installed. An unavailable
// provider falls back to the CPU rather than failing, so this loads on any machine.
func newProviderTestEmbedder(t *testing.T, name ModelName, provider onnx.Provider) Embedder {
	t.Helper()

	m := FindEmbeddingModel(name)
	require.NotNil(t, m)

	if !m.Installed(embeddingModelsPath) {
		t.Skipf("faces: %s is not installed", name)
	}

	embedder, err := NewONNXEmbedder(EmbedderSettings{
		Name:      name,
		Model:     m,
		ModelPath: m.FilePath(embeddingModelsPath),
		Threads:   1,
		Provider:  provider,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = embedder.Close() })

	return embedder
}

// requireProviderApplied skips the test unless the loader actually got the provider asked for.
// Without a device the CUDA loaders fall back to the CPU, and a parity test would then compare
// the CPU with itself and pass no matter what the GPU path does - so it must skip visibly
// rather than report a result it did not measure.
func requireProviderApplied(t *testing.T, applied, want onnx.Provider) {
	t.Helper()

	if applied != want {
		t.Skipf("faces: skipping, the %s execution provider is unavailable here (running on the %s)", want, applied)
	}
}

// providerTestCrop returns a deterministic face-sized image. A fixed pattern is used rather
// than a photo so the comparison does not depend on installed test data, and rather than a zero
// image so the two providers are compared on values that exercise the graph.
func providerTestCrop(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{
				R: uint8((x*7 + y*3) % 256),
				G: uint8((x*5 + y*11) % 256),
				B: uint8((x*13 + y*17) % 256),
				A: 255,
			})
		}
	}

	return img
}

func TestONNXEmbedder_ProviderParity(t *testing.T) {
	t.Run("SameVector", func(t *testing.T) {
		// Embeddings are persisted and compared by distance, so a vector computed under one
		// provider must be comparable with one computed under the other. They are not bit
		// identical: the providers run the same FP32 graph with different kernels, so what has
		// to hold is that they agree far below the smallest threshold any comparison uses.
		cpu := newProviderTestEmbedder(t, ModelSFace, onnx.ProviderCPU)
		gpu := newProviderTestEmbedder(t, ModelSFace, onnx.ProviderCUDA)

		gpuEmbedder, ok := gpu.(*onnxEmbedder)
		require.True(t, ok)
		requireProviderApplied(t, gpuEmbedder.provider, onnx.ProviderCUDA)

		width, height := cpu.CropSize()
		crop := providerTestCrop(width, height)

		cpuEmb := cpu.Run(crop)
		gpuEmb := gpu.Run(crop)

		require.Len(t, cpuEmb, 1)
		require.Len(t, gpuEmb, 1)

		cpuVec, gpuVec := cpuEmb[0], gpuEmb[0]
		require.NotEmpty(t, cpuVec)
		require.Len(t, gpuVec, len(cpuVec))

		var maxDiff float64

		for i := range cpuVec {
			maxDiff = math.Max(maxDiff, math.Abs(cpuVec[i]-gpuVec[i]))
		}

		dist := cpuVec.Dist(gpuVec)
		t.Logf("faces: largest per-dimension difference %g over %d dimensions, distance %g (match %g, cluster %g)",
			maxDiff, len(cpuVec), dist, MatchDistDefault, ClusterDistDefault)

		assert.Less(t, dist, MatchDistDefault/1000)
	})
}

// providerTestImage is a bundled photograph with faces, used to compare what the two providers
// actually persist rather than raw tensor values.
const providerTestImage = "testdata/1.jpg"

func TestONNXEngine_ProviderParity(t *testing.T) {
	t.Run("SameDetections", func(t *testing.T) {
		// The detector's warm-up geometry is derived rather than fixed, and its landmarks
		// decide the crop every embedding is computed from, so a provider difference here
		// would move vectors much further than the embedder's own kernel noise.
		if _, err := os.Stat(detectorModelPath); err != nil {
			t.Skipf("faces: %s is not installed", detectorModelPath)
		}

		if _, err := os.Stat(providerTestImage); err != nil {
			t.Skipf("faces: %s is not available", providerTestImage)
		}

		newEngine := func(provider onnx.Provider) DetectionEngine {
			engine, err := NewONNXEngine(ONNXOptions{ModelPath: detectorModelPath, Threads: 1, Provider: provider})
			require.NoError(t, err)
			t.Cleanup(func() { _ = engine.Close() })

			return engine
		}

		cpu := newEngine(onnx.ProviderCPU)
		gpu := newEngine(onnx.ProviderCUDA)

		gpuEngine, ok := gpu.(*onnxEngine)
		require.True(t, ok)
		requireProviderApplied(t, gpuEngine.provider, onnx.ProviderCUDA)

		cpuFaces, err := cpu.Detect(providerTestImage, 20)
		require.NoError(t, err)
		require.NotEmpty(t, cpuFaces)

		gpuFaces, err := gpu.Detect(providerTestImage, 20)
		require.NoError(t, err)

		require.Len(t, gpuFaces, len(cpuFaces))

		// Scores and landmark areas are rounded to integers before anything is persisted, so
		// this is the granularity at which a provider difference would actually reach a user.
		for i := range cpuFaces {
			assert.Equal(t, cpuFaces[i].Score, gpuFaces[i].Score, "face %d score", i)
			assert.Equal(t, cpuFaces[i].Area, gpuFaces[i].Area, "face %d area", i)
			assert.Equal(t, cpuFaces[i].Eyes, gpuFaces[i].Eyes, "face %d eyes", i)
			assert.Equal(t, cpuFaces[i].Landmarks, gpuFaces[i].Landmarks, "face %d landmarks", i)
		}

		t.Logf("faces: %d detections identical under both providers", len(cpuFaces))
	})
}
