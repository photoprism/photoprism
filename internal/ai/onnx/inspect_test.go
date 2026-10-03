package onnx

import (
	"os"
	"path/filepath"
	"testing"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var detectorModelPath, _ = filepath.Abs("../../../assets/models/yunet/face_detection_yunet_2026may.onnx")
var embeddingModelPath, _ = filepath.Abs("../../../assets/models/sface/face_recognition_sface_2021dec.onnx")
var labelModelPath, _ = filepath.Abs("../../../assets/models/efficientformerv2_s1/efficientformerv2_s1.onnx")

// requireRuntime skips a test when the ONNX Runtime or the model it needs is unavailable,
// which is the case in build environments that did not run "make dep".
func requireRuntime(t *testing.T, modelPath string) {
	t.Helper()

	if _, err := os.Stat(modelPath); err != nil {
		t.Skipf("onnx: skipping, %s is not available", filepath.Base(modelPath))
	}

	if err := EnsureRuntime(""); err != nil {
		t.Skipf("onnx: skipping, %s", err)
	}
}

func TestInspect(t *testing.T) {
	t.Run("Detector", func(t *testing.T) {
		requireRuntime(t, detectorModelPath)

		info, err := Inspect(detectorModelPath, nil)
		require.NoError(t, err)
		assert.Equal(t, filepath.Base(detectorModelPath), info.File)
		assert.NotEmpty(t, info.Input.Name)
		assert.Equal(t, LayoutNCHW, info.Input.Layout)
		// The bundled detector is a dynamic export, so a spatial axis reads as unset and the
		// geometry to run it at comes from the registry rather than from the graph.
		assert.Zero(t, info.Input.Width)
		assert.Zero(t, info.Input.Height)
	})
	t.Run("EmbeddingModel", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)

		info, err := Inspect(embeddingModelPath, nil)
		require.NoError(t, err)
		assert.Equal(t, 112, info.Input.Width)
		assert.Equal(t, 112, info.Input.Height)
		assert.Equal(t, LayoutNCHW, info.Input.Layout)
		assert.Equal(t, 128, info.Output.Width)
		assert.Equal(t, 1, info.Output.Count)
	})
	t.Run("LabelModel", func(t *testing.T) {
		requireRuntime(t, labelModelPath)

		info, err := Inspect(labelModelPath, nil)
		require.NoError(t, err)
		assert.Equal(t, 224, info.Input.Width)
		assert.Equal(t, 224, info.Input.Height)
		assert.Equal(t, 1000, info.Output.Width)
		assert.Equal(t, 1, info.Output.Count)
		t.Logf("input %q, output %q", info.Input.Name, info.Output.Name)
	})
	t.Run("PreprocessingStaysUnset", func(t *testing.T) {
		// Channel order, normalization, and the resize convention are not present in a
		// graph, so inspection must leave them for the registry to supply.
		requireRuntime(t, embeddingModelPath)

		info, err := Inspect(embeddingModelPath, nil)
		require.NoError(t, err)
		assert.Equal(t, OrderUndefined, info.Input.ColorOrder)
		assert.True(t, info.Input.Normalization.IsZero())
		assert.True(t, info.Input.Resize.IsZero())
	})
	t.Run("MissingFile", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)

		_, err := Inspect(filepath.Join(t.TempDir(), "missing.onnx"), nil)
		require.Error(t, err)
	})
}

func TestMetadata(t *testing.T) {
	t.Run("NoPhotoPrismKeys", func(t *testing.T) {
		// The bundled models are mirrored upstream artifacts, so they carry no
		// photoprism.* entries; only models we export ourselves do.
		requireRuntime(t, embeddingModelPath)

		values, err := Metadata(embeddingModelPath)
		require.NoError(t, err)
		assert.Empty(t, values)
	})
	t.Run("MissingFile", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)

		_, err := Metadata(filepath.Join(t.TempDir(), "missing.onnx"))
		require.Error(t, err)
	})
	t.Run("LabelExport", func(t *testing.T) {
		requireRuntime(t, labelModelPath)

		values, err := Metadata(labelModelPath)
		require.NoError(t, err)
		assert.NotEmpty(t, values["source"])
		assert.NotEmpty(t, values["checkpoint"])
		assert.Equal(t, "17", values["opset"])

		info, err := InfoFromMetadata(values)
		require.NoError(t, err)
		assert.Equal(t, values["source"], info.Source)
		assert.Equal(t, LayoutNCHW, info.Input.Layout)
		assert.Equal(t, RGB, info.Input.ColorOrder)
		assert.Equal(t, ResizeCenterCrop, info.Input.Resize.Mode)
		assert.Equal(t, InterpolationBicubic, info.Input.Resize.Interpolation)
		t.Logf("metadata: %#v", values)
	})
}

func TestInputGeometry(t *testing.T) {
	t.Run("NCHW", func(t *testing.T) {
		width, height, layout := InputGeometry(onnxruntime.Shape{1, 3, 112, 96})
		assert.Equal(t, 96, width)
		assert.Equal(t, 112, height)
		assert.Equal(t, LayoutNCHW, layout)
	})
	t.Run("NHWC", func(t *testing.T) {
		width, height, layout := InputGeometry(onnxruntime.Shape{1, 112, 96, 3})
		assert.Equal(t, 96, width)
		assert.Equal(t, 112, height)
		assert.Equal(t, LayoutNHWC, layout)
	})
	t.Run("DynamicSpatialAxes", func(t *testing.T) {
		width, height, layout := InputGeometry(onnxruntime.Shape{1, 3, -1, -1})
		assert.Equal(t, 0, width)
		assert.Equal(t, 0, height)
		assert.Equal(t, LayoutNCHW, layout)
	})
	t.Run("DynamicChannelAxis", func(t *testing.T) {
		// A dynamic channel axis cannot be told apart from a dynamic spatial one, so the
		// layout stays undefined rather than being guessed.
		width, height, layout := InputGeometry(onnxruntime.Shape{1, -1, -1, -1})
		assert.Equal(t, 0, width)
		assert.Equal(t, 0, height)
		assert.Equal(t, LayoutUndefined, layout)
	})
	t.Run("TooFewAxes", func(t *testing.T) {
		_, _, layout := InputGeometry(onnxruntime.Shape{1, 3, 112})
		assert.Equal(t, LayoutUndefined, layout)
	})
}

func TestTensorAxis(t *testing.T) {
	t.Run("Static", func(t *testing.T) {
		assert.Equal(t, 512, tensorAxis(onnxruntime.Shape{1, 512}, 1))
	})
	t.Run("Dynamic", func(t *testing.T) {
		assert.Equal(t, 0, tensorAxis(onnxruntime.Shape{1, -1}, 1))
	})
	t.Run("OutOfRange", func(t *testing.T) {
		assert.Equal(t, 0, tensorAxis(onnxruntime.Shape{1, 512}, 5))
		assert.Equal(t, 0, tensorAxis(onnxruntime.Shape{1, 512}, -1))
	})
}
