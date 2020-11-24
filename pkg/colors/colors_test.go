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

func TestChroma_Hex(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("chroma 15", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, "F", perception.Chroma.Hex())
	})
	t.Run("chroma 155", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 155}
		assert.Equal(t, "9B", perception.Chroma.Hex())
	})
}

func TestChroma_Value(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("chroma 15", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, uint8(0xf), perception.Chroma.Value())
	})
	t.Run("chroma 155", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 155}
		assert.Equal(t, uint8(0x9b), perception.Chroma.Value())
	})
}

func TestChroma_Uint(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("chroma 15", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, uint(0xf), perception.Chroma.Uint())
	})
	t.Run("chroma 155", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 155}
		assert.Equal(t, uint(0x9b), perception.Chroma.Uint())
	})
}

func TestChroma_Int(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("chroma 15", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, 15, perception.Chroma.Int())
	})
	t.Run("chroma 155", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 155}
		assert.Equal(t, 155, perception.Chroma.Int())
	})
}
