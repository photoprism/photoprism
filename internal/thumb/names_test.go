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
		assert.Equal(t, Fit2048, name)
		assert.Equal(t, 2048, size.Width)
		assert.Equal(t, 2048, size.Height)
	})

	t.Run("2000", func(t *testing.T) {
		name, size := Find(2000)
		assert.Equal(t, Fit1920, name)
		assert.Equal(t, 1920, size.Width)
		assert.Equal(t, 1200, size.Height)
	})
}
