package face

import (
	"image"
	"image/color"
	"math"
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

// providerTestCrop returns a deterministic face-sized crop. A fixed pattern is used rather than
// a photo so the comparison does not depend on installed test data, and rather than a zero
// image so the two providers are compared on values that actually exercise the graph.
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
		// Embeddings are persisted, so a vector computed under one provider must be
		// comparable with one computed under the other. Where no GPU is present the CUDA
		// embedder falls back to the CPU, and the assertion still holds.
		cpu := newProviderTestEmbedder(t, ModelSFace, onnx.ProviderCPU)
		gpu := newProviderTestEmbedder(t, ModelSFace, onnx.ProviderCUDA)

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

		// The providers run the same graph in FP32 but not with the same kernels, so the
		// vectors agree to floating-point noise rather than bit for bit. They are persisted
		// and compared by distance, so what has to hold is that the difference is orders of
		// magnitude below the smallest threshold any comparison uses.
		assert.Less(t, dist, MatchDistDefault/1000)
	})
}
