package classify

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

var assetsPath = fs.Abs("../../../assets")
var samplesPath = filepath.Join(assetsPath, "samples")
var modelsPath = filepath.Join(assetsPath, "models")

// TestSoftmax verifies numerically stable probability conversion.
func TestSoftmax(t *testing.T) {
	t.Run("Stable", func(t *testing.T) {
		probabilities, err := softmax([]float32{10000, 9999, -10000})
		require.NoError(t, err)
		require.Len(t, probabilities, 3)
		assert.Greater(t, probabilities[0], probabilities[1])
		assert.InDelta(t, 1, probabilitySum(probabilities), 1e-6)
	})
	t.Run("NonFinite", func(t *testing.T) {
		_, err := softmax([]float32{0, float32(math.Inf(1))})
		require.Error(t, err)
	})
	t.Run("Empty", func(t *testing.T) {
		_, err := softmax(nil)
		require.Error(t, err)
	})
}

// TestValidateProbabilities verifies probability range and sum checks.
func TestValidateProbabilities(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		require.NoError(t, validateProbabilities([]float32{0.25, 0.75}))
	})
	t.Run("WrongSum", func(t *testing.T) {
		require.Error(t, validateProbabilities([]float32{0.25, 0.5}))
	})
	t.Run("Negative", func(t *testing.T) {
		require.Error(t, validateProbabilities([]float32{-0.25, 1.25}))
	})
}

// TestResizeInput verifies the model-specific geometry conventions.
func TestResizeInput(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 400, 200))

	t.Run("CenterCrop", func(t *testing.T) {
		result := resizeInput(source, 224, 224, onnx.Resize{
			Mode:          onnx.ResizeCenterCrop,
			ShortEdge:     236,
			Interpolation: onnx.InterpolationBicubic,
		})
		assert.Equal(t, image.Rect(0, 0, 224, 224), result.Bounds())
	})
	t.Run("Stretch", func(t *testing.T) {
		result := resizeInput(source, 224, 224, onnx.Resize{Mode: onnx.ResizeStretch})
		assert.Equal(t, image.Rect(0, 0, 224, 224), result.Bounds())
	})
	t.Run("Pad", func(t *testing.T) {
		result := resizeInput(source, 224, 224, onnx.Resize{Mode: onnx.ResizePad})
		assert.Equal(t, image.Rect(0, 0, 224, 224), result.Bounds())
	})
}

// TestModelBuildBlob verifies channel order, layout, and normalization.
func TestModelBuildBlob(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})

	model := NewModel(Settings{Info: &onnx.ModelInfo{Input: &onnx.Input{
		Width:         1,
		Height:        1,
		Layout:        onnx.LayoutNCHW,
		ColorOrder:    onnx.BGR,
		Normalization: onnx.Uniform(0, 10),
		Resize:        onnx.Resize{Mode: onnx.ResizeStretch},
	}}})
	model.mean = model.meta.Input.Normalization.Mean
	model.scales = model.meta.Input.Normalization.Scales()

	blob, err := model.buildBlob(img)
	require.NoError(t, err)
	assert.Equal(t, []float32{3, 2, 1}, blob)
}

// TestRegisteredModelBuildBlob verifies the default model's complete preprocessing tensor.
func TestRegisteredModelBuildBlob(t *testing.T) {
	description := DefaultModel()
	require.NotNil(t, description)
	model := NewModel(Settings{Info: description.ONNX})
	model.mean = model.meta.Input.Normalization.Mean
	model.scales = model.meta.Input.Normalization.Scales()

	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.NRGBA{R: 255, G: 128, B: 0, A: 255}}, image.Point{}, draw.Src)
	blob, err := model.buildBlob(img)
	require.NoError(t, err)
	planeSize := 224 * 224
	require.Len(t, blob, planeSize*onnx.Channels)
	assert.InDelta(t, (255-imageNetMean[0])/imageNetStdDev[0], blob[0], 1e-6)
	assert.InDelta(t, (128-imageNetMean[1])/imageNetStdDev[1], blob[planeSize], 1e-6)
	assert.InDelta(t, (0-imageNetMean[2])/imageNetStdDev[2], blob[2*planeSize], 1e-6)
}

// TestRegisteredModelBuildBlobCenterCrop verifies the short-edge resize removes side bands.
func TestRegisteredModelBuildBlobCenterCrop(t *testing.T) {
	description := DefaultModel()
	require.NotNil(t, description)
	model := NewModel(Settings{Info: description.ONNX})
	model.mean = model.meta.Input.Normalization.Mean
	model.scales = model.meta.Input.Normalization.Scales()

	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	draw.Draw(img, image.Rect(0, 0, 100, 200), &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(100, 0, 300, 200), &image.Uniform{C: color.NRGBA{G: 255, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(300, 0, 400, 200), &image.Uniform{C: color.NRGBA{B: 255, A: 255}}, image.Point{}, draw.Src)

	blob, err := model.buildBlob(img)
	require.NoError(t, err)
	planeSize := 224 * 224
	for _, pixel := range []int{0, planeSize / 2, planeSize - 1} {
		assert.InDelta(t, (0-imageNetMean[0])/imageNetStdDev[0], blob[pixel], 1e-6)
		assert.InDelta(t, (255-imageNetMean[1])/imageNetStdDev[1], blob[planeSize+pixel], 1e-6)
		assert.InDelta(t, (0-imageNetMean[2])/imageNetStdDev[2], blob[2*planeSize+pixel], 1e-6)
	}
}

// TestRegisteredModelBuildBlobFixture pins the complete default preprocessing tensor.
func TestRegisteredModelBuildBlobFixture(t *testing.T) {
	description := DefaultModel()
	require.NotNil(t, description)
	model := NewModel(Settings{Info: description.ONNX})
	model.mean = model.meta.Input.Normalization.Mean
	model.scales = model.meta.Input.Normalization.Scales()

	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			img.SetNRGBA(x, y, color.NRGBA{
				R: uint8((x*13 + y*3) % 256),
				G: uint8((x*5 + y*17) % 256),
				B: uint8((x*19 + y*7) % 256),
				A: 255,
			})
		}
	}

	blob, err := model.buildBlob(img)
	require.NoError(t, err)
	digest := sha256.New()
	buffer := make([]byte, 4)
	for _, value := range blob {
		binary.LittleEndian.PutUint32(buffer, math.Float32bits(value))
		_, err = digest.Write(buffer)
		require.NoError(t, err)
	}

	assert.Equal(t, "617e74633e89df115e3ebb1acb4552b1a52bff8e3438e0f8991f0a92d7da969c", hex.EncodeToString(digest.Sum(nil)))
}

// TestReadLabels verifies exact newline-separated vocabulary loading.
func TestReadLabels(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "labels.txt")
	require.NoError(t, os.WriteFile(fileName, []byte("tench fish\ngoldfish\n"), fs.ModeFile))

	labels, err := readLabels(fileName)
	require.NoError(t, err)
	assert.Equal(t, []string{"tench fish", "goldfish"}, labels)
}

// TestModelLoadLabels verifies custom widths and mismatched vocabularies.
func TestModelLoadLabels(t *testing.T) {
	t.Run("CustomWidth", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "custom-21k.txt")
		require.NoError(t, os.WriteFile(fileName, []byte("first\nsecond\nthird\n"), fs.ModeFile))
		model := &Model{labelPath: fileName, meta: &onnx.ModelInfo{Output: &onnx.Output{Width: 3}}}
		require.NoError(t, model.loadLabels())
		assert.Equal(t, []string{"first", "second", "third"}, model.labels)
		require.NoError(t, model.validateLabels())
	})
	t.Run("Mismatch", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "labels.txt")
		require.NoError(t, os.WriteFile(fileName, []byte("first\nsecond\n"), fs.ModeFile))
		model := &Model{labelPath: fileName, meta: &onnx.ModelInfo{Output: &onnx.Output{Width: 3}}}
		require.Error(t, model.loadLabels())
	})
	t.Run("Missing", func(t *testing.T) {
		model := &Model{labelPath: filepath.Join(t.TempDir(), "missing.txt"), meta: &onnx.ModelInfo{Output: &onnx.Output{Width: 3}}}
		require.Error(t, model.loadLabels())
		assert.Empty(t, model.labels)
	})
}

// TestModelValidateLabels verifies canonical ImageNet guards.
func TestModelValidateLabels(t *testing.T) {
	t.Run("Background", func(t *testing.T) {
		labels := make([]string, ImageNetClasses)
		labels[0] = "background"
		model := &Model{
			canonicalOrder: true,
			labels:         labels,
			meta:           &onnx.ModelInfo{Output: &onnx.Output{Width: ImageNetClasses}},
		}
		require.Error(t, model.validateLabels())
	})
	t.Run("ClassCount", func(t *testing.T) {
		model := &Model{
			canonicalOrder: true,
			labels:         []string{"one", "two"},
			meta:           &onnx.ModelInfo{Output: &onnx.Output{Width: 2}},
		}
		require.Error(t, model.validateLabels())
	})
}

// TestModelBestLabels verifies existing label-rule mapping remains in force.
func TestModelBestLabels(t *testing.T) {
	model := &Model{labels: []string{"tench fish", "goldfish", "great white shark", "tiger shark", "hammerhead", "electric ray", "stingray", "cock", "hen"}}
	probabilities := make([]float32, len(model.labels))
	probabilities[8] = 0.7
	probabilities[1] = 0.5

	result := model.bestLabels(probabilities, 10)
	require.NotEmpty(t, result)
	assert.Equal(t, "chicken", result[0].Name)
	assert.Equal(t, SrcImage, result[0].Source)
}

// TestRegisteredModelIntegration verifies a bundled ONNX graph can classify a fixture.
func TestRegisteredModelIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ONNX model integration test in short mode")
	}

	for _, name := range AutoModelPreference {
		t.Run(string(name), func(t *testing.T) {
			model := requireRegisteredModel(t, name)
			defer func() { require.NoError(t, model.Close()) }()

			result, err := model.File(filepath.Join(samplesPath, "dog_orange.jpg"), 10)
			require.NoError(t, err)
			require.NotEmpty(t, result)
			assert.Equal(t, "dog", result[0].Name)
		})
	}
}

// TestCustomModelMetadataIntegration verifies a registry-free model can describe its preprocessing.
func TestCustomModelMetadataIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ONNX model integration test in short mode")
	}

	description := DefaultModel()
	require.NotNil(t, description)
	modelPath := description.ONNX.FilePath(filepath.Join(modelsPath, string(description.Name)))
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skip("default ONNX model is not installed")
	}

	model := NewModel(Settings{
		Name:      "custom_metadata",
		ModelPath: modelPath,
		LabelPath: filepath.Join(modelsPath, "nasnet", "labels.txt"),
	})
	defer func() { require.NoError(t, model.Close()) }()

	result, err := model.File(filepath.Join(samplesPath, "dog_orange.jpg"), 10)
	require.NoError(t, err)
	require.NotEmpty(t, result)
	assert.Equal(t, "dog", result[0].Name)
	assert.Equal(t, onnx.ResizeCenterCrop, model.meta.Input.Resize.Mode)
	assert.Equal(t, 236, model.meta.Input.Resize.ShortEdge)
}

// TestModelRunConcurrent verifies deterministic output under parallel indexing.
func TestModelRunConcurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping ONNX model integration test in short mode")
	}

	model := requireRegisteredModel(t, DefaultModelName())
	defer func() { require.NoError(t, model.Close()) }()

	cases := []struct {
		file string
		want string
	}{
		{file: "dog_orange.jpg", want: "dog"},
		{file: "cat_224.jpeg", want: "cat"},
		{file: "cat_720.jpeg", want: "cat"},
		{file: "zebra_green_brown.jpg", want: "zebra"},
	}
	images := make([][]byte, len(cases))
	for i := range cases {
		data, err := os.ReadFile(filepath.Join(samplesPath, cases[i].file)) //nolint:gosec // bundled test fixture
		require.NoError(t, err)
		images[i] = data
	}

	const rounds = 2
	results := make([][]Labels, len(cases))
	errors := make([]error, len(cases))
	var wait sync.WaitGroup
	start := make(chan struct{})

	for i := range cases {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			for range rounds {
				var result Labels
				result, errors[index] = model.Run(images[index], 10)
				if errors[index] != nil {
					return
				}
				results[index] = append(results[index], result)
			}
		}(i)
	}

	close(start)
	wait.Wait()
	for i := range cases {
		require.NoError(t, errors[i])
		require.Len(t, results[i], rounds)
		for _, result := range results[i] {
			require.NotEmpty(t, result, cases[i].file)
			assert.Equal(t, cases[i].want, result[0].Name, cases[i].file)
		}
	}
}

// TestModelDisabledAndErrors verifies disabled models and invalid inputs fail safely.
func TestModelDisabledAndErrors(t *testing.T) {
	disabled := NewModel(Settings{Disabled: true})
	require.NoError(t, disabled.Init())
	assert.False(t, disabled.ModelLoaded())
	result, err := disabled.Run([]byte("not an image"), 10)
	require.NoError(t, err)
	assert.Nil(t, result)
	require.NoError(t, disabled.Close())

	missing := NewModel(Settings{ModelPath: filepath.Join(t.TempDir(), "missing.onnx")})
	require.Error(t, missing.Init())
	assert.False(t, missing.ModelLoaded())

	model := requireRegisteredModel(t, DefaultModelName())
	defer func() { require.NoError(t, model.Close()) }()
	result, err = model.Run([]byte("not an image"), 10)
	require.Error(t, err)
	assert.Empty(t, result)
}

// BenchmarkModelRun measures inference on the selected bundled classifier.
func BenchmarkModelRun(b *testing.B) {
	model := NewRegisteredModel(modelsPath, DefaultModelName(), false)
	if model == nil || model.Init() != nil {
		b.Skip("bundled ONNX model is unavailable")
	}
	defer func() { _ = model.Close() }()

	data, err := os.ReadFile(filepath.Join(samplesPath, "dog_orange.jpg")) //nolint:gosec // bundled test fixture
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err = model.Run(data, 10); err != nil {
			b.Fatal(err)
		}
	}
}

// requireRegisteredModel loads a registry entry or skips when its artifact is absent.
func requireRegisteredModel(t *testing.T, name ModelName) *Model {
	t.Helper()

	model := NewRegisteredModel(modelsPath, name, false)
	if model == nil {
		t.Skipf("classify: model %s is not registered", name)
	}

	if _, err := os.Stat(model.modelPath); err != nil {
		t.Skipf("classify: model %s is not installed", name)
	}

	require.NoError(t, model.Init())

	return model
}

// probabilitySum returns the sum of all probabilities.
func probabilitySum(probabilities []float32) float64 {
	var result float64
	for _, probability := range probabilities {
		result += float64(probability)
	}

	return result
}
