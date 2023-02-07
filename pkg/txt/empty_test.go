package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmpty(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, true, Empty(""))
	})
	t.Run("EnNew", func(t *testing.T) {
		assert.Equal(t, false, Empty(EnNew))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, false, Empty("     new "))
	})
	t.Run("Uppercase", func(t *testing.T) {
		assert.Equal(t, false, Empty("NEW"))
	})
	t.Run("Lowercase", func(t *testing.T) {
		assert.Equal(t, false, Empty("new"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, false, Empty("New"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, Empty("non"))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, true, Empty("0"))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, true, Empty("-1"))
	})
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, true, Empty("0000:00:00 00:00:00"))
	})
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, true, Empty("nil"))
	})
	t.Run("NaN", func(t *testing.T) {
		assert.Equal(t, true, Empty("NaN"))
	})
	t.Run("NULL", func(t *testing.T) {
		assert.Equal(t, true, Empty("NULL"))
	})
	t.Run("*", func(t *testing.T) {
		assert.Equal(t, true, Empty("*"))
	})
	t.Run("%", func(t *testing.T) {
		assert.Equal(t, true, Empty("%"))
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
	t.Run("Date", func(t *testing.T) {
		assert.Equal(t, false, NotEmpty("0000:00:00 00:00:00"))
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

func TestEmptyTime(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		assert.True(t, EmptyTime(""))
	})
	t.Run("0000-00-00 00-00-00", func(t *testing.T) {
		assert.True(t, EmptyTime("0000-00-00 00-00-00"))
	})
	t.Run("0000:00:00 00:00:00", func(t *testing.T) {
		assert.True(t, EmptyTime("0000:00:00 00:00:00"))
	})
	t.Run("0000-00-00 00:00:00", func(t *testing.T) {
		assert.True(t, EmptyTime("0000-00-00 00:00:00"))
	})
	t.Run("0001-01-01 00:00:00 +0000 UTC", func(t *testing.T) {
		assert.True(t, EmptyTime("0001-01-01 00:00:00 +0000 UTC"))
	})
}
