package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVector_Normalize(t *testing.T) {
	t.Run("UnitLength", func(t *testing.T) {
		v := Vector{3, 4}
		v.Normalize()
		assert.InDelta(t, 0.6, v[0], 1e-9)
		assert.InDelta(t, 0.8, v[1], 1e-9)
		assert.InDelta(t, 1.0, v.EuclideanNorm(), 1e-9)
	})
	t.Run("Negative", func(t *testing.T) {
		v := Vector{-3, 4}
		v.Normalize()
		assert.InDelta(t, -0.6, v[0], 1e-9)
		assert.InDelta(t, 0.8, v[1], 1e-9)
	})
	t.Run("Single", func(t *testing.T) {
		v := Vector{5}
		v.Normalize()
		assert.InDelta(t, 1.0, v[0], 1e-9)
	})
	t.Run("ZeroVector", func(t *testing.T) {
		v := Vector{0, 0}
		v.Normalize()
		assert.Equal(t, Vector{0, 0}, v, "a zero vector must be left unchanged")
	})
	t.Run("Empty", func(t *testing.T) {
		v := Vector{}
		v.Normalize()
		assert.Equal(t, Vector{}, v)
	})
}

func TestVector_Normalized(t *testing.T) {
	t.Run("ReturnsCopy", func(t *testing.T) {
		orig := Vector{3, 4}
		got := orig.Normalized()
		// Receiver stays unchanged.
		assert.InDelta(t, 3.0, orig[0], 1e-9)
		assert.InDelta(t, 4.0, orig[1], 1e-9)
		// Returned vector is normalized.
		assert.InDelta(t, 0.6, got[0], 1e-9)
		assert.InDelta(t, 0.8, got[1], 1e-9)
	})
	t.Run("ZeroVector", func(t *testing.T) {
		orig := Vector{0, 0}
		got := orig.Normalized()
		assert.Equal(t, Vector{0, 0}, got)
		assert.Equal(t, Vector{0, 0}, orig)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, Vector{}, Vector{}.Normalized())
	})
}
