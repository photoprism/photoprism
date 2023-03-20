package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrientation(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, 0, Orientation(0))
	})

	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, 1, Orientation(1))
		assert.Equal(t, 3, Orientation(3))
		assert.Equal(t, 5, Orientation(5))
		assert.Equal(t, 7, Orientation(7))
		assert.Equal(t, 8, Orientation(8))
	})

	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, 0, Orientation(-1))
		assert.Equal(t, 0, Orientation(9))
		assert.Equal(t, 0, Orientation(2000))
	})
}
