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

func TestDetectorModel(t *testing.T) {
	t.Run("Fields", func(t *testing.T) {
		// Preprocessing cannot be read from a graph, so every registered detector has to
		// carry it: an entry that left it unset would run under another detector's channel
		// order and quietly detect worse.
		for _, d := range Detectors {
			require.NotNil(t, d.ONNX, d.Name)
			require.NotNil(t, d.ONNX.Input, d.Name)
			assert.NotEmpty(t, d.ONNX.File, d.Name)
			assert.Positive(t, d.ONNX.Input.Width, d.Name)
			assert.Positive(t, d.ONNX.Input.Height, d.Name)
			assert.True(t, d.ONNX.Input.ColorOrder.Valid(), d.Name)
			assert.NotEmpty(t, d.ONNX.License, d.Name)
		}
	})
	t.Run("InstalledArtifactsMatch", func(t *testing.T) {
		// Every detector that happens to be installed, rather than only the one that ships:
		// a stale opt-in download decodes as the registered artifact and is not detectable
		// from its detections.
		for _, d := range Detectors {
			path := d.Path(assetsModelsPath)

			if _, err := os.Stat(path); err != nil {
				continue
			}

			assert.NoError(t, d.ONNX.VerifyChecksum(path), d.Name)
		}
	})
}

func TestONNXEngineBuildBlob(t *testing.T) {
	input := FindDetector(DetectorSCRFD).ONNX.Input

	engine := &onnxEngine{
		inputWidth:  4,
		inputHeight: 4,
		colorOrder:  input.ColorOrder,
		mean:        input.Normalization.Mean,
		scales:      input.Normalization.Scales(),
	}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	blob, scale, err := engine.buildBlob(img)
	require.NoError(t, err)
	require.Len(t, blob, 4*4*3)
	// Stated as literals rather than as the registered mean and standard deviation, so
	// that changing the registry changes what this test measures rather than what it
	// compares against.
	require.InDelta(t, (255-127.5)/128.0, blob[0], 1e-3)
	require.InDelta(t, (0-127.5)/128.0, blob[16], 1e-3)
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

// TestDetectorRecall pins how many faces the detector finds per test image, so a preprocessing
// change cannot pass unnoticed: a wrong channel order still returns plausible detections, and
// nothing else here would fail. Images pinned to zero are negatives, so the table guards against
// new false positives as well as lost recall. 19.jpg is the exception and is pinned to what the
// detector does: it holds flowers, three of which clear the shipping score threshold.
func TestDetectorRecall(t *testing.T) {
	pinned := []struct {
		fileName string
		faces    int
	}{
		{"1.jpg", 2},
		{"2.jpg", 1},
		{"3.jpg", 1},
		{"4.jpg", 1},
		{"5.jpg", 1},
		{"6.jpg", 1},
		{"7.jpg", 0},
		{"8.jpg", 0},
		{"9.jpg", 0},
		{"10.jpg", 0},
		{"11.jpg", 0},
		{"12.jpg", 1},
		{"13.jpg", 0},
		{"14.jpg", 0},
		{"15.jpg", 0},
		{"16.jpg", 2},
		{"17.jpg", 2},
		{"18.jpg", 2},
		{"19.jpg", 3},
	}

	prev := UseEngine(nil)
	t.Cleanup(func() {
		if current := UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	if err := ConfigureEngine(EngineSettings{
		Name: EngineONNX,
		ONNX: ONNXOptions{ModelPath: detectorModelPath, Threads: 1},
	}); err != nil {
		if _, statErr := os.Stat(detectorModelPath); statErr != nil {
			t.Skipf("faces: skipping detector-dependent test, %s is not available", filepath.Base(detectorModelPath))
		}

		t.Fatalf("faces: failed to initialize detector: %s", err)
	}

	for _, c := range pinned {
		t.Run(c.fileName, func(t *testing.T) {
			faces, err := Detect(filepath.Join("testdata", c.fileName), 20)
			require.NoError(t, err)
			assert.Len(t, faces, c.faces)
		})
	}
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
	t.Run("ClosedSession", func(t *testing.T) {
		// Reconfiguring the detector closes the session while workers may still hold it,
		// so a detection that arrives afterwards has to report it rather than use it.
		fileName, err := filepath.Abs("testdata/1.jpg")
		require.NoError(t, err)

		engine := &onnxEngine{inputWidth: 640, inputHeight: 640}
		require.NoError(t, engine.Close())

		faces, err := engine.Detect(fileName, 20)

		assert.Empty(t, faces)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})
}

func TestONNXEngineClose(t *testing.T) {
	t.Run("Idempotent", func(t *testing.T) {
		engine := &onnxEngine{}
		require.NoError(t, engine.Close())
		require.NoError(t, engine.Close())
	})
}

func TestDetectorInputSize(t *testing.T) {
	detector := DefaultDetector()
	defaultWidth, defaultHeight := detector.ONNX.InputSize()

	t.Run("DeclaredGeometry", func(t *testing.T) {
		w, h, err := detectorInputSize(detector, defaultWidth, defaultHeight)
		require.NoError(t, err)
		assert.Equal(t, defaultWidth, w)
		assert.Equal(t, defaultHeight, h)
	})
	t.Run("DynamicAxes", func(t *testing.T) {
		// A dynamic axis reports zero, so the registered size stands in for it.
		w, h, err := detectorInputSize(detector, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, defaultWidth, w)
		assert.Equal(t, defaultHeight, h)
	})
	t.Run("DynamicWidthOnly", func(t *testing.T) {
		w, h, err := detectorInputSize(detector, 0, 480)
		require.NoError(t, err)
		assert.Equal(t, defaultWidth, w)
		assert.Equal(t, 480, h)
	})
	t.Run("AtLimit", func(t *testing.T) {
		w, h, err := detectorInputSize(detector, maxDetectorInputSize, maxDetectorInputSize)
		require.NoError(t, err)
		assert.Equal(t, maxDetectorInputSize, w)
		assert.Equal(t, maxDetectorInputSize, h)
	})
	t.Run("WidthTooLarge", func(t *testing.T) {
		// The blob is sized from these values, so an axis this wide is not an export that
		// can be run.
		_, _, err := detectorInputSize(detector, maxDetectorInputSize+1, defaultHeight)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds")
	})
	t.Run("HeightTooLarge", func(t *testing.T) {
		_, _, err := detectorInputSize(detector, defaultWidth, 65536)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds")
	})
}
