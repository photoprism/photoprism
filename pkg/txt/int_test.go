package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := Int("")
		assert.Equal(t, 0, result)
	})

	t.Run("non-numeric", func(t *testing.T) {
		result := Int("Screenshot")
		assert.Equal(t, 0, result)
	})

	t.Run("zero", func(t *testing.T) {
		result := Int("0")
		assert.Equal(t, 0, result)
	})

	t.Run("int", func(t *testing.T) {
		result := Int("123")
		assert.Equal(t, 123, result)
	})

	t.Run("negative int", func(t *testing.T) {
		result := Int("-123")
		assert.Equal(t, -123, result)
	})
}
func TestIsUInt(t *testing.T) {
	assert.False(t, IsUInt(""))
	assert.False(t, IsUInt("12 3"))
	assert.True(t, IsUInt("123"))
}

func TestUInt(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		result := UInt("")
		assert.Equal(t, uint(0), result)
	})

	t.Run("non-numeric", func(t *testing.T) {
		result := UInt("Screenshot")
		assert.Equal(t, uint(0), result)
	})

	t.Run("zero", func(t *testing.T) {
		result := UInt("0")
		assert.Equal(t, uint(0), result)
	})

	t.Run("int", func(t *testing.T) {
		result := UInt("123")
		assert.Equal(t, uint(0x7b), result)
	})

	t.Run("negative int", func(t *testing.T) {
		result := UInt("-123")
		assert.Equal(t, uint(0), result)
	})
}
