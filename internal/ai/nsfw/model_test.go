package nsfw

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
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

var testModelsPath, _ = filepath.Abs("../../../assets/models")

// testModel returns a detector description that can exercise preprocessing without a graph.
func testModel(layout onnx.Layout, order onnx.ColorOrder, logits bool) *Model {
	model := NewModel(Settings{
		Name: "test",
		Info: &onnx.ModelInfo{
			Input: &onnx.Input{Width: 1, Height: 1, Layout: layout, ColorOrder: order,
				Normalization: onnx.Normalization{StdDev: [onnx.Channels]float32{1, 1, 1}},
				Resize:        onnx.Resize{Mode: onnx.ResizeStretch, Interpolation: onnx.InterpolationLinear}},
			Output: &onnx.Output{Width: 2, Count: 1, Logits: onnx.Bool(logits)},
		},
		Reduction:        ReductionSoftmaxUnsafe,
		UnsafeClassIndex: 1,
	})
	model.mean = model.meta.Input.Normalization.Mean
	model.scales = model.meta.Input.Normalization.Scales()
	return model
}

// TestModelReduction verifies every supported output contract and its validation.
func TestModelReduction(t *testing.T) {
	t.Run("SoftmaxUnsafe", func(t *testing.T) {
		model := testModel(onnx.LayoutNCHW, onnx.RGB, true)
		score, err := model.reduceOutput([]float32{0, 2})
		require.NoError(t, err)
		assert.InDelta(t, 0.880797, score, 1e-6)
	})
	t.Run("ProbabilityUnsafe", func(t *testing.T) {
		model := testModel(onnx.LayoutNCHW, onnx.RGB, false)
		score, err := model.reduceOutput([]float32{0.2, 0.8})
		require.NoError(t, err)
		assert.InDelta(t, 0.8, score, 1e-6)
	})
	t.Run("NeutralComplement", func(t *testing.T) {
		model := testModel(onnx.LayoutNCHW, onnx.RGB, true)
		model.meta.Output.Width = 4
		model.reduction = ReductionNeutralComplement
		model.neutralClassIndex = 0
		score, err := model.reduceOutput([]float32{2, 0, 0, 0})
		require.NoError(t, err)
		assert.InDelta(t, 0.288765, score, 1e-6)
	})
	t.Run("SigmoidUnsafe", func(t *testing.T) {
		model := testModel(onnx.LayoutNCHW, onnx.RGB, true)
		model.meta.Output.Width = 1
		model.reduction = ReductionSigmoidUnsafe
		score, err := model.reduceOutput([]float32{2})
		require.NoError(t, err)
		assert.InDelta(t, 0.880797, score, 1e-6)
	})
	t.Run("WrongWidth", func(t *testing.T) {
		_, err := testModel(onnx.LayoutNCHW, onnx.RGB, true).reduceOutput([]float32{0})
		require.Error(t, err)
	})
	t.Run("NonFinite", func(t *testing.T) {
		_, err := testModel(onnx.LayoutNCHW, onnx.RGB, true).reduceOutput([]float32{0, float32(math.Inf(1))})
		require.Error(t, err)
	})
	t.Run("InvalidProbability", func(t *testing.T) {
		_, err := testModel(onnx.LayoutNCHW, onnx.RGB, false).reduceOutput([]float32{0.5, 0.6})
		require.Error(t, err)
	})
}

// TestModelBuildBlob verifies channel order, layout, and normalization.
func TestModelBuildBlob(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 10, G: 20, B: 30, A: 255})

	t.Run("NCHWRGB", func(t *testing.T) {
		blob, err := testModel(onnx.LayoutNCHW, onnx.RGB, true).buildBlob(img)
		require.NoError(t, err)
		assert.Equal(t, []float32{10, 20, 30}, blob)
	})
	t.Run("NHWCBGR", func(t *testing.T) {
		blob, err := testModel(onnx.LayoutNHWC, onnx.BGR, true).buildBlob(img)
		require.NoError(t, err)
		assert.Equal(t, []float32{30, 20, 10}, blob)
	})
}

// TestResizeModelInput verifies center cropping does not depend on SubImage support.
func TestResizeModelInput(t *testing.T) {
	img := struct{ image.Image }{Image: image.NewNRGBA(image.Rect(0, 0, 4, 6))}
	resized := resizeModelInput(img, 2, 2, onnx.Resize{Mode: onnx.ResizeCenterCrop, ShortEdge: 2})
	assert.Equal(t, image.Rect(0, 0, 2, 2), resized.Bounds())
}

// TestModelDisabled verifies every disabled entry point reports no decision.
func TestModelDisabled(t *testing.T) {
	disabled := NewModel(Settings{Disabled: true})
	require.NoError(t, disabled.Init())
	assert.False(t, disabled.ModelLoaded())

	result, err := disabled.File("missing.jpg", DefaultThreshold)
	require.NoError(t, err)
	assert.True(t, result.IsUnavailable())
	result, err = disabled.Url("https://example.com/image.jpg", DefaultThreshold)
	require.NoError(t, err)
	assert.True(t, result.IsUnavailable())
	result, err = disabled.Run([]byte("not an image"), DefaultThreshold)
	require.NoError(t, err)
	assert.True(t, result.IsUnavailable())
	require.NoError(t, disabled.Close())
}

// TestModelFailures verifies missing graphs and unreadable inputs never report safe.
func TestModelFailures(t *testing.T) {
	missing := NewModel(Settings{ModelPath: filepath.Join(t.TempDir(), "missing.onnx")})
	require.Error(t, missing.Init())
	assert.False(t, missing.ModelLoaded())

	result, err := missing.File(filepath.Join(t.TempDir(), "missing.jpg"), DefaultThreshold)
	require.Error(t, err)
	assert.False(t, result.IsSafe())
	assert.True(t, result.IsUnavailable())

	fileName := filepath.Join(t.TempDir(), "invalid.jpg")
	require.NoError(t, os.WriteFile(fileName, []byte("not an image"), fs.ModeFile))
	result, err = missing.File(fileName, DefaultThreshold)
	require.Error(t, err)
	assert.False(t, result.IsSafe())
}

// TestModelNilReceiver verifies a missing detector reports no decision.
func TestModelNilReceiver(t *testing.T) {
	var model *Model
	result, err := model.Run(nil, DefaultThreshold)
	require.NoError(t, err)
	assert.True(t, result.IsUnavailable())
	assert.False(t, result.IsSafe())
}

// TestRegisteredModels verifies registry descriptions are complete and independent.
func TestRegisteredModels(t *testing.T) {
	for name, description := range Models {
		t.Run(string(name), func(t *testing.T) {
			require.NotNil(t, description.ONNX)
			require.NotNil(t, description.ONNX.Input)
			require.NotNil(t, description.ONNX.Output)
			assert.Len(t, description.ONNX.SHA256, 64)
			assert.NotEmpty(t, description.ONNX.Source)
			assert.NotEmpty(t, description.ONNX.License)
			model := NewRegisteredModel("/models", name, false)
			require.NotNil(t, model)
			assert.Equal(t, filepath.Join("/models", string(name), description.ONNX.File), model.modelPath)
			require.NoError(t, model.validateDescription())
		})
	}
}

// TestRegisteredModelInference verifies the bundled graph accepts JPEG and PNG input.
func TestRegisteredModelInference(t *testing.T) {
	model := NewRegisteredModel(testModelsPath, DefaultModelName(), false)
	if model == nil || !fs.FileExists(model.modelPath) {
		t.Skip("nsfw: default ONNX model is not installed")
	}
	require.NoError(t, model.Init())
	defer func() { require.NoError(t, model.Close()) }()

	jpegResult, err := model.File(filepath.Join("testdata", "cat_brown.jpg"), 0.75)
	require.NoError(t, err)
	assert.False(t, jpegResult.IsUnavailable())
	t.Logf("cat_brown.jpg unsafe score: %.6f", jpegResult.Score)

	img := image.NewNRGBA(image.Rect(0, 0, 32, 24))
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, img))
	pngResult, err := model.Run(encoded.Bytes(), 0.75)
	require.NoError(t, err)
	assert.False(t, pngResult.IsUnavailable())
}

// TestRegisteredModelConcurrentInference verifies shared-session results stay deterministic.
func TestRegisteredModelConcurrentInference(t *testing.T) {
	model := NewRegisteredModel(testModelsPath, DefaultModelName(), false)
	if model == nil || !fs.FileExists(model.modelPath) {
		t.Skip("nsfw: default ONNX model is not installed")
	}
	require.NoError(t, model.Init())
	defer func() { require.NoError(t, model.Close()) }()

	const workers = 4
	scores := make([]float32, workers)
	errors := make([]error, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := range workers {
		go func(index int) {
			defer wait.Done()
			result, err := model.File(filepath.Join("testdata", "cat_brown.jpg"), 0.75)
			scores[index], errors[index] = result.Score, err
		}(i)
	}
	wait.Wait()
	for i := range workers {
		require.NoError(t, errors[i])
		assert.InDelta(t, scores[0], scores[i], 1e-7)
	}
}
