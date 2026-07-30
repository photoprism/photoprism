package face

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/pkg/fs"
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

func TestONNXEngineDetectLandmarks(t *testing.T) {
	prev := UseEngine(nil)
	t.Cleanup(func() {
		if current := UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	err := ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{ModelPath: detectorModelPath, Threads: 1},
	})
	if err != nil {
		t.Skipf("faces: skipping detector-dependent test: %s", err)
	}

	fileName := filepath.Join("testdata", "1.jpg")
	faces, err := Detect(fileName, 20)
	require.NoError(t, err)
	require.NotEmpty(t, faces)

	f := &faces[0]

	t.Run("EyesAndLandmarks", func(t *testing.T) {
		require.Len(t, f.Eyes, 2)
		require.Len(t, f.Landmarks, NumLandmarks-2)
		assert.Equal(t, "eye_l", f.Eyes[0].Name)
		assert.Equal(t, "nose", f.Landmarks[0].Name)
	})
	t.Run("WithinImage", func(t *testing.T) {
		for _, areas := range []Areas{f.Eyes, f.Landmarks} {
			for _, a := range areas {
				assert.GreaterOrEqual(t, a.Col, 0)
				assert.GreaterOrEqual(t, a.Row, 0)
				assert.Less(t, a.Col, f.Cols)
				assert.Less(t, a.Row, f.Rows)
			}
		}
	})
	t.Run("EyesAboveMouth", func(t *testing.T) {
		// A frontal portrait must place both eyes above both mouth corners.
		assert.Less(t, f.Eyes[0].Row, f.Landmarks[1].Row)
		assert.Less(t, f.Eyes[1].Row, f.Landmarks[2].Row)
	})
	t.Run("LeftEyeLeftOfRightEye", func(t *testing.T) {
		assert.Less(t, f.Eyes[0].Col, f.Eyes[1].Col)
	})
	t.Run("AlignedCrop", func(t *testing.T) {
		img, _, imgErr := fs.DecodeImageFile(fileName)
		require.NoError(t, imgErr)

		out, cropErr := AlignedCrop(img, f, ArcFaceTemplateSize, ArcFaceTemplateSize)
		require.NoError(t, cropErr)
		assert.Equal(t, ArcFaceTemplateSize, out.Bounds().Dx())
		assert.Equal(t, ArcFaceTemplateSize, out.Bounds().Dy())
	})
}

func TestONNXEngineDetect(t *testing.T) {
	t.Run("MalformedTiffIfdOffset", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "evil.tiff")
		require.NoError(t, os.WriteFile(fileName, []byte{0x49, 0x49, 0x2a, 0x00, 0xff, 0xff, 0xff, 0xff}, fs.ModeFile))

		faces, err := (&onnxEngine{}).Detect(fileName, 20)

		assert.Empty(t, faces)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid TIFF: IFD offset")
	})
}
