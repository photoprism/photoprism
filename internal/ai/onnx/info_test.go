package onnx

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInput_IsDynamic(t *testing.T) {
	t.Run("Dynamic", func(t *testing.T) {
		assert.True(t, (&Input{}).IsDynamic())
	})
	t.Run("Fixed", func(t *testing.T) {
		assert.False(t, (&Input{Width: 112, Height: 112}).IsDynamic())
	})
	t.Run("Nil", func(t *testing.T) {
		var input *Input
		assert.False(t, input.IsDynamic())
	})
}

func TestInput_Merge(t *testing.T) {
	t.Run("FillsEmpty", func(t *testing.T) {
		input := &Input{Width: 112}
		input.Merge(&Input{
			Name:          "data",
			Width:         640,
			Height:        112,
			Layout:        LayoutNCHW,
			ColorOrder:    BGR,
			Normalization: Uniform(127.5, 128),
			Resize:        Resize{Mode: ResizePad},
		})

		assert.Equal(t, "data", input.Name)
		assert.Equal(t, 112, input.Width, "an established value must not be overwritten")
		assert.Equal(t, 112, input.Height)
		assert.Equal(t, LayoutNCHW, input.Layout)
		assert.Equal(t, BGR, input.ColorOrder)
		assert.Equal(t, Uniform(127.5, 128), input.Normalization)
		assert.Equal(t, ResizePad, input.Resize.Mode)
	})
	t.Run("NilOther", func(t *testing.T) {
		input := &Input{Width: 112}
		input.Merge(nil)
		assert.Equal(t, 112, input.Width)
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var input *Input
		assert.NotPanics(t, func() { input.Merge(&Input{Width: 112}) })
	})
}

func TestOutput_Merge(t *testing.T) {
	t.Run("FillsEmpty", func(t *testing.T) {
		output := &Output{Width: 128}
		output.Merge(&Output{Name: "embedding", Width: 512, Count: 1, Logits: Bool(true)})

		assert.Equal(t, "embedding", output.Name)
		assert.Equal(t, 128, output.Width)
		assert.Equal(t, 1, output.Count)
		assert.True(t, output.OutputsLogits())
	})
	t.Run("PreservesExplicitFalse", func(t *testing.T) {
		output := &Output{Logits: Bool(false)}
		output.Merge(&Output{Logits: Bool(true)})

		assert.False(t, output.OutputsLogits())
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var output *Output
		assert.NotPanics(t, func() { output.Merge(&Output{Width: 512}) })
	})
}

func TestModelInfo_FilePath(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &ModelInfo{File: "sface.onnx"}
		assert.Equal(t, filepath.Join("/models", "sface.onnx"), m.FilePath("/models"))
	})
	t.Run("MissingFile", func(t *testing.T) {
		assert.Empty(t, (&ModelInfo{}).FilePath("/models"))
	})
	t.Run("Nil", func(t *testing.T) {
		var m *ModelInfo
		assert.Empty(t, m.FilePath("/models"))
	})
}

func TestModelInfo_InputSize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		width, height := (&ModelInfo{Input: &Input{Width: 112, Height: 96}}).InputSize()
		assert.Equal(t, 112, width)
		assert.Equal(t, 96, height)
	})
	t.Run("MissingInput", func(t *testing.T) {
		width, height := (&ModelInfo{}).InputSize()
		assert.Equal(t, 0, width)
		assert.Equal(t, 0, height)
	})
}

func TestModelInfo_OutputWidth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, 128, (&ModelInfo{Output: &Output{Width: 128}}).OutputWidth())
	})
	t.Run("MissingOutput", func(t *testing.T) {
		assert.Equal(t, 0, (&ModelInfo{}).OutputWidth())
	})
}

func TestModelInfo_Merge(t *testing.T) {
	t.Run("FillsEmpty", func(t *testing.T) {
		m := &ModelInfo{File: "sface.onnx", Input: &Input{Width: 112}}
		m.Merge(&ModelInfo{
			File:         "other.onnx",
			Source:       "https://example.com/sface.onnx",
			SHA256:       "0ba9fbfa",
			License:      "Apache-2.0",
			Quantization: "fp32",
			Input:        &Input{Name: "data", Height: 112},
			Output:       &Output{Name: "embedding", Width: 128},
		})

		assert.Equal(t, "sface.onnx", m.File)
		assert.Equal(t, "https://example.com/sface.onnx", m.Source)
		assert.Equal(t, "0ba9fbfa", m.SHA256)
		assert.Equal(t, "Apache-2.0", m.License)
		assert.Equal(t, "fp32", m.Quantization)
		assert.Equal(t, "data", m.Input.Name)
		assert.Equal(t, 112, m.Input.Height)
		assert.Equal(t, 128, m.OutputWidth())
	})
	t.Run("NilOther", func(t *testing.T) {
		m := &ModelInfo{File: "sface.onnx"}
		m.Merge(nil)
		assert.Equal(t, "sface.onnx", m.File)
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var m *ModelInfo
		assert.NotPanics(t, func() { m.Merge(&ModelInfo{File: "sface.onnx"}) })
	})
}
