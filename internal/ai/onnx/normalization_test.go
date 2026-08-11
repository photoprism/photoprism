package onnx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniform(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		n := Uniform(127.5, 128)
		assert.Equal(t, [Channels]float32{127.5, 127.5, 127.5}, n.Mean)
		assert.Equal(t, [Channels]float32{128, 128, 128}, n.StdDev)
	})
	t.Run("Identity", func(t *testing.T) {
		n := Uniform(0, 1)
		assert.False(t, n.IsZero())
		assert.Equal(t, [Channels]float32{1, 1, 1}, n.Scales())
	})
}

func TestNormalization_Scales(t *testing.T) {
	t.Run("Reciprocal", func(t *testing.T) {
		assert.Equal(t, [Channels]float32{1.0 / 128, 1.0 / 128, 1.0 / 128}, Uniform(127.5, 128).Scales())
	})
	t.Run("PerChannel", func(t *testing.T) {
		n := Normalization{StdDev: [Channels]float32{2, 4, 8}}
		assert.Equal(t, [Channels]float32{0.5, 0.25, 0.125}, n.Scales())
	})
	t.Run("ZeroStaysUnscaled", func(t *testing.T) {
		assert.Equal(t, [Channels]float32{1, 1, 1}, Normalization{}.Scales())
	})
}

func TestNormalization_IsZero(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, Normalization{}.IsZero())
	})
	t.Run("MeanOnly", func(t *testing.T) {
		assert.False(t, Normalization{Mean: [Channels]float32{127.5}}.IsZero())
	})
}
