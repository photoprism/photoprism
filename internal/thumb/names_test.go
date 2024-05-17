package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName_Jpeg(t *testing.T) {
	t.Run("ResamplePng, FillCenter", func(t *testing.T) {
		assert.Equal(t, "tile_50.jpg", Tile50.Jpeg())
	})
}

func TestFind(t *testing.T) {
	t.Run("2048", func(t *testing.T) {
		name, size := Find(2048)
		assert.Equal(t, Fit1920, name)
		assert.Equal(t, 1920, size.Width)
		assert.Equal(t, 1200, size.Height)
	})

	t.Run("1900", func(t *testing.T) {
		name, size := Find(1900)
		assert.Equal(t, Fit1280, name)
		assert.Equal(t, 1280, size.Width)
		assert.Equal(t, 1024, size.Height)
	})
}
