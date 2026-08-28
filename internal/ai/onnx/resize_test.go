package onnx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResize_IsZero(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, Resize{}.IsZero())
	})
	t.Run("ModeOnly", func(t *testing.T) {
		assert.False(t, Resize{Mode: ResizePad}.IsZero())
	})
	t.Run("ShortEdgeConvention", func(t *testing.T) {
		assert.False(t, Resize{Mode: ResizeCenterCrop, ShortEdge: 236, CropRatio: 0.95}.IsZero())
	})
}
