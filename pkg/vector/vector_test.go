package vector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVector(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		v, err := NewVector([]int{1, 2, 3, 4, 6, 5})
		assert.IsType(t, Vector{}, v)
		assert.NoError(t, err)
	})
	t.Run("Float32", func(t *testing.T) {
		v, err := NewVector([]float32{1.0, 2.1, 3.54, 4.9, 6.666666, 5.33333333})
		assert.IsType(t, Vector{}, v)
		assert.NoError(t, err)
	})
	t.Run("Float64", func(t *testing.T) {
		v, err := NewVector([]float64{1.0, 2.1, 3.54, 4.9, 6.666666, 5.33333333})
		assert.IsType(t, Vector{}, v)
		assert.NoError(t, err)
	})
	t.Run("Uint8", func(t *testing.T) {
		v, err := NewVector([]uint8{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{1, 2, 3}, v)
	})
	t.Run("Uint16", func(t *testing.T) {
		v, err := NewVector([]uint16{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{1, 2, 3}, v)
	})
	t.Run("Uint32", func(t *testing.T) {
		v, err := NewVector([]uint32{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{1, 2, 3}, v)
	})
	t.Run("Uint64", func(t *testing.T) {
		v, err := NewVector([]uint64{1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{1, 2, 3}, v)
	})
	t.Run("Int8", func(t *testing.T) {
		v, err := NewVector([]int8{-1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{-1, 2, 3}, v)
	})
	t.Run("Int16", func(t *testing.T) {
		v, err := NewVector([]int16{-1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{-1, 2, 3}, v)
	})
	t.Run("Int32", func(t *testing.T) {
		v, err := NewVector([]int32{-1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{-1, 2, 3}, v)
	})
	t.Run("Int64", func(t *testing.T) {
		v, err := NewVector([]int64{-1, 2, 3})
		assert.NoError(t, err)
		assert.Equal(t, Vector{-1, 2, 3}, v)
	})
	t.Run("Vector", func(t *testing.T) {
		src := Vector{1, 2, 3}
		v, err := NewVector(src)
		assert.NoError(t, err)
		assert.Equal(t, src, v)
	})
	t.Run("String", func(t *testing.T) {
		v, err := NewVector([]string{"a", "b", "c"})
		assert.IsType(t, Vector{}, v)
		assert.Error(t, err)
	})
}

func TestNewVector_Float64Copy(t *testing.T) {
	// NewVector must return a vector independent of the source slice.
	src := []float64{1, 2, 3}
	v, err := NewVector(src)
	assert.NoError(t, err)
	src[0] = 99
	assert.Equal(t, Vector{1, 2, 3}, v)
}

func TestNullVector(t *testing.T) {
	v := NullVector(3)
	assert.Equal(t, 3, v.Dim())
	assert.Equal(t, Vector{0, 0, 0}, v)
}

func TestVector_Copy(t *testing.T) {
	a := Vector{1, 2, 3}
	b := a.Copy()
	b[0] = 99
	assert.Equal(t, Vector{1, 2, 3}, a, "modifying the copy must not change the original")
	assert.Equal(t, Vector{99, 2, 3}, b)
}

func TestVector_Dim(t *testing.T) {
	assert.Equal(t, 3, Vector{1, 2, 3}.Dim())
	assert.Equal(t, 0, Vector{}.Dim())
}

func TestVector_Sum(t *testing.T) {
	assert.InDelta(t, 6.0, Vector{1, 2, 3}.Sum(), 0.00001)
	assert.InDelta(t, 0.0, Vector{}.Sum(), 0.00001)
}
