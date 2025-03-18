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
	t.Run("String", func(t *testing.T) {
		v, err := NewVector([]string{"a", "b", "c"})
		assert.IsType(t, Vector{}, v)
		assert.Error(t, err)
	})
}
