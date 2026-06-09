package vector

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMean(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		assert.InDelta(t, 2.5, Mean(Vector{1, 2, 3, 4}), 0.00001)
	})
	t.Run("Single", func(t *testing.T) {
		assert.InDelta(t, 7.0, Mean(Vector{7}), 0.00001)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, math.IsNaN(Mean(Vector{})))
	})
}

func TestGeometricMean(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		assert.InDelta(t, 4.0, GeometricMean(Vector{2, 8}), 0.00001)
	})
	t.Run("PowersOfThree", func(t *testing.T) {
		assert.InDelta(t, math.Pow(3, 1.5), GeometricMean(Vector{1, 3, 9, 27}), 0.00001)
	})
	t.Run("Method", func(t *testing.T) {
		assert.InDelta(t, 4.0, Vector{2, 8}.GeometricMean(), 0.00001)
	})
	t.Run("ContainsZero", func(t *testing.T) {
		// A zero element drives the product to 0, so the geometric mean is 0.
		assert.InDelta(t, 0.0, GeometricMean(Vector{2, 0, 8}), 0.00001)
	})
	t.Run("LeadingZero", func(t *testing.T) {
		assert.InDelta(t, 0.0, GeometricMean(Vector{0, 4, 9}), 0.00001)
	})
	t.Run("Negative", func(t *testing.T) {
		assert.True(t, math.IsNaN(GeometricMean(Vector{-1, 2})))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, math.IsNaN(GeometricMean(Vector{})))
	})
}

func TestHarmonicMean(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		assert.InDelta(t, 3.0/1.75, HarmonicMean(Vector{1, 2, 4}), 0.00001)
	})
	t.Run("Equal", func(t *testing.T) {
		assert.InDelta(t, 2.0, HarmonicMean(Vector{2, 2}), 0.00001)
	})
	t.Run("Method", func(t *testing.T) {
		assert.InDelta(t, 3.0/1.75, Vector{1, 2, 4}.HarmonicMean(), 0.00001)
	})
	t.Run("ContainsZero", func(t *testing.T) {
		assert.True(t, math.IsNaN(HarmonicMean(Vector{1, 0, 2})))
	})
	t.Run("Negative", func(t *testing.T) {
		assert.True(t, math.IsNaN(HarmonicMean(Vector{1, -2})))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, math.IsNaN(HarmonicMean(Vector{})))
	})
}

func TestVector_WeightedMean(t *testing.T) {
	t.Run("Values", func(t *testing.T) {
		r, err := Vector{1, 2, 4}.WeightedMean(Vector{1, 0, 1})
		assert.NoError(t, err)
		assert.InDelta(t, 2.5, r, 0.00001)
	})
	t.Run("LengthMismatch", func(t *testing.T) {
		r, err := Vector{1, 2, 4}.WeightedMean(Vector{1, 1})
		assert.Error(t, err)
		assert.True(t, math.IsNaN(r))
	})
}
