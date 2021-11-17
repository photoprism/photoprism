package crop

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName_Jpeg(t *testing.T) {
	t.Run("Tile320", func(t *testing.T) {
		assert.Equal(t, "tile_320.jpg", Tile320.Jpeg())
	})
	t.Run("Tile50", func(t *testing.T) {
		assert.Equal(t, "tile_50.jpg", Tile50.Jpeg())
	})
}
