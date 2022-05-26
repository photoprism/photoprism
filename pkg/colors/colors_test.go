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
	assert.Equal(t, "C", Magenta.Hex())
	assert.Equal(t, "7", Cyan.Hex())
}

func TestColors_Hex(t *testing.T) {
	result := Colors{Orange, Lime, Black}.Hex()
	assert.Equal(t, "DA0", result)
}

func TestColor_Uint8(t *testing.T) {
	assert.Equal(t, int8(7), Cyan.ID())
}
