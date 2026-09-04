package onnx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInfoFromMetadata verifies PhotoPrism model metadata parsing.
func TestInfoFromMetadata(t *testing.T) {
	values := map[string]string{
		"source":        "https://example.com/model.safetensors",
		"layout":        "NCHW",
		"colorOrder":    "RGB",
		"mean":          "[123.675, 116.28, 103.53]",
		"stdDev":        "[58.395, 57.12, 57.375]",
		"cropPct":       "0.95",
		"interpolation": "bicubic",
		"outputWidth":   "1000",
		"outputCount":   "1",
		"logits":        "true",
	}

	info, err := InfoFromMetadata(values)
	require.NoError(t, err)
	require.NotNil(t, info.Input)
	require.NotNil(t, info.Output)
	assert.Equal(t, "https://example.com/model.safetensors", info.Source)
	assert.Equal(t, LayoutNCHW, info.Input.Layout)
	assert.Equal(t, RGB, info.Input.ColorOrder)
	assert.InDelta(t, 123.675, info.Input.Normalization.Mean[0], 1e-3)
	assert.InDelta(t, 58.395, info.Input.Normalization.StdDev[0], 1e-3)
	assert.Equal(t, ResizeCenterCrop, info.Input.Resize.Mode)
	assert.Equal(t, float32(0.95), info.Input.Resize.CropRatio)
	assert.Equal(t, InterpolationBicubic, info.Input.Resize.Interpolation)
	assert.Equal(t, 1000, info.Output.Width)
	assert.Equal(t, 1, info.Output.Count)
	assert.NotNil(t, info.Output.Logits)
	assert.True(t, info.Output.OutputsLogits())

	info.Input.Width = 224
	info.Input.Height = 224
	CompleteResizeMetadata(info)
	assert.Equal(t, 236, info.Input.Resize.ShortEdge)
}

// TestInfoFromMetadataErrors verifies malformed semantic metadata is rejected.
func TestInfoFromMetadataErrors(t *testing.T) {
	t.Run("ExplicitProbabilities", func(t *testing.T) {
		info, err := InfoFromMetadata(map[string]string{"logits": "false"})
		require.NoError(t, err)
		require.NotNil(t, info.Output.Logits)
		assert.False(t, info.Output.OutputsLogits())
	})
	t.Run("Channels", func(t *testing.T) {
		_, err := InfoFromMetadata(map[string]string{"mean": "[0.5, 0.5]"})
		require.Error(t, err)
	})
	t.Run("UnitValuesRemainUnscaled", func(t *testing.T) {
		info, err := InfoFromMetadata(map[string]string{"mean": "[0, 0, 0]", "stdDev": "[1, 1, 1]"})
		require.NoError(t, err)
		assert.Equal(t, Uniform(0, 1), info.Input.Normalization)
	})
	t.Run("Layout", func(t *testing.T) {
		_, err := InfoFromMetadata(map[string]string{"layout": "CHWN"})
		require.Error(t, err)
	})
	t.Run("Crop", func(t *testing.T) {
		_, err := InfoFromMetadata(map[string]string{"cropPct": "2"})
		require.Error(t, err)
	})
}
