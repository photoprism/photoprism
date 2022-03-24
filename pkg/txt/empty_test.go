package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, IsEmpty("non"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("0"))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("-1"))
	})
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("NULL"))
	})
	t.Run("*", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("*"))
	})
	t.Run("%", func(t *testing.T) {
		assert.Equal(t, true, IsEmpty("%"))
	})
}

func TestNotEmpty(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, true, NotEmpty("non"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0"))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("-1"))
	})
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("NULL"))
	})
	t.Run("*", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("*"))
	})
	t.Run("%", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("%"))
	})
}
