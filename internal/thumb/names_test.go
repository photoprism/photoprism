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

func TestVision_DefaultsAndBounds(t *testing.T) {
	// Exact 720 returns Fit720
	sz := Vision(720)
	assert.Equal(t, SizeFit720, sz)
	// Below 224 selects the smallest square tile >= resolution
	assert.Equal(t, SizeTile100, Vision(100))
	// Next square tile at or above resolution
	assert.Equal(t, SizeTile384, Vision(300))
	assert.Equal(t, SizeTile500, Vision(500))
}
