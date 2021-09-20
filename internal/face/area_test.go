package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var area1 = NewArea("face1", 400, 250, 200)
var area2 = NewArea("face2", 100, 100, 50)
var area3 = NewArea("face3", 900, 500, 25)
var area4 = NewArea("face4", 110, 110, 60)

func TestArea_TopLeft(t *testing.T) {
	t.Run("area1", func(t *testing.T) {
		x, y := area1.TopLeft()
		assert.Equal(t, 300, x)
		assert.Equal(t, 150, y)
	})
	t.Run("area2", func(t *testing.T) {
		x, y := area2.TopLeft()
		assert.Equal(t, 75, x)
		assert.Equal(t, 75, y)
	})
	t.Run("area3", func(t *testing.T) {
		x, y := area3.TopLeft()
		assert.Equal(t, 888, x)
		assert.Equal(t, 488, y)
	})
	t.Run("area4", func(t *testing.T) {
		x, y := area4.TopLeft()
		assert.Equal(t, 80, x)
		assert.Equal(t, 80, y)
	})
}
