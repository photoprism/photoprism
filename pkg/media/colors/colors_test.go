package colors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColors_List(t *testing.T) {
	allColors := All.List()

	assert.Equal(t, "purple", allColors[0]["Slug"])
	assert.Equal(t, "Purple", allColors[0]["Name"])
	assert.Equal(t, "magenta", allColors[1]["Slug"])
	assert.Equal(t, "Magenta", allColors[1]["Name"])
	assert.Equal(t, "#EF5350", allColors[3]["Example"])

	t.Logf("colors: %+v", allColors)
}

func TestColor_Hex(t *testing.T) {
	assert.Equal(t, "0", Color(-1).Hex())
	assert.Equal(t, "0", Black.Hex())
	assert.Equal(t, "C", Magenta.Hex())
	assert.Equal(t, "7", Cyan.Hex())
	assert.Equal(t, "F", Pink.Hex())
	assert.Equal(t, "F", Color(15).Hex())
	assert.Equal(t, "0", Color(16).Hex())
	assert.Equal(t, "0", Color(17).Hex())
}

func TestColors_Hex(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		assert.Equal(t, "5CFED3BA98762410", All.Hex())
	})
	t.Run("OrangeLimeBlack", func(t *testing.T) {
		assert.Equal(t, "DA0", Colors{Orange, Lime, Black}.Hex())
	})
}

func TestColor_ID(t *testing.T) {
	assert.Equal(t, int16(7), Cyan.ID())
}
