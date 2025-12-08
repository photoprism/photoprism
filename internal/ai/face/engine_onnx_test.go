package face

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	onnxruntime "github.com/yalue/onnxruntime_go"
)

// TestONNXSharedLibraryCandidates_Defaults verifies default search ordering when no explicit path is provided.
func TestONNXSharedLibraryCandidates_Defaults(t *testing.T) {
	t.Cleanup(func() { onnxExecutableVar = os.Executable })
	onnxExecutableVar = func() (string, error) {
		return filepath.Join("/opt/photoprism", "bin", "photoprism"), nil
	}

	candidates := onnxSharedLibraryCandidates("")
	require.NotEmpty(t, candidates)
	require.Equal(t, "libonnxruntime.so", candidates[0])
	require.Contains(t, candidates, filepath.Join("/opt/photoprism", "lib", "libonnxruntime.so"))
}

// TestONNXSharedLibraryCandidates_ExplicitFirst ensures explicit paths remain the first candidate.
func TestONNXSharedLibraryCandidates_ExplicitFirst(t *testing.T) {
	t.Cleanup(func() { onnxExecutableVar = os.Executable })
	onnxExecutableVar = func() (string, error) { return "/tmp/photoprism", nil }

	explicit := "/custom/libonnxruntime.so"
	candidates := onnxSharedLibraryCandidates(explicit)
	require.NotEmpty(t, candidates)
	require.Equal(t, explicit, candidates[0])
}

func TestDeriveONNXLayout(t *testing.T) {
	outputs := make([]onnxruntime.InputOutputInfo, 9)
	outputs[0] = onnxruntime.InputOutputInfo{Dimensions: onnxruntime.Shape{1, 3, 3}}

	fmc, anchors, useKps, batched, err := deriveONNXLayout(outputs)
	require.NoError(t, err)
	require.Equal(t, 3, fmc)
	require.Equal(t, 2, anchors)
	require.True(t, useKps)
	require.True(t, batched)

	_, _, _, _, err = deriveONNXLayout(make([]onnxruntime.InputOutputInfo, 1))
	require.Error(t, err)
}

func TestStridesForFeatureMaps(t *testing.T) {
	require.Equal(t, []int{8, 16, 32, 64, 128}, stridesForFeatureMaps(5))
	require.Equal(t, []int{8, 16, 32}, stridesForFeatureMaps(3))
}

func TestONNXEngineAnchorCentersCaches(t *testing.T) {
	engine := &onnxEngine{centerCache: make(map[anchorCacheKey][]float32)}
	centers1 := engine.anchorCenters(2, 2, 8, 2)
	require.Len(t, centers1, 16)
	centers2 := engine.anchorCenters(2, 2, 8, 2)
	// The cache should return the same backing array.
	require.Equal(t, &centers1[0], &centers2[0])
}

func TestONNXEngineBuildBlob(t *testing.T) {
	engine := &onnxEngine{inputWidth: 4, inputHeight: 4}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	blob, scale, err := engine.buildBlob(img)
	require.NoError(t, err)
	require.Len(t, blob, 4*4*3)
	require.InDelta(t, (255-onnxInputMean)/onnxInputStd, blob[0], 1e-3)
	require.InDelta(t, (0-onnxInputMean)/onnxInputStd, blob[16], 1e-3)
	require.Equal(t, float32(4), scale)
}
