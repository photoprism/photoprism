package colors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChroma_Percent(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("ChromaFifteen", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, int16(15), perception.Chroma.Percent())
	})
	t.Run("ChromaNum127", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 127}
		assert.Equal(t, int16(100), perception.Chroma.Percent())
	})
}

func TestChroma_Uint(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("ChromaFifteen", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, uint(0xf), perception.Chroma.Uint())
	})
	t.Run("ChromaNum127", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 127}
		assert.Equal(t, uint(100), perception.Chroma.Uint())
	})
}

func TestChroma_Int(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("ChromaFifteen", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, 15, perception.Chroma.Int())
	})
	t.Run("ChromaOne", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: -1}
		assert.Equal(t, 0, perception.Chroma.Int())
	})
	t.Run("ChromaNum127", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: -127}
		assert.Equal(t, 0, perception.Chroma.Int())
	})
	t.Run("ChromaNum100", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 100}
		assert.Equal(t, 100, perception.Chroma.Int())
	})
	t.Run("ChromaNum127", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 127}
		assert.Equal(t, 100, perception.Chroma.Int())
	})
}

func TestChroma_Hex(t *testing.T) {
	lum := []Luminance{1, 16, 2, 4, 15, 16, 1, 0, 8}
	lMap := LightMap(lum)

	t.Run("ChromaFifteen", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 15}
		assert.Equal(t, "F", perception.Chroma.Hex())
	})
	t.Run("ChromaNum127", func(t *testing.T) {
		perception := ColorPerception{Colors: Colors{Orange, Lime, Cyan}, MainColor: Cyan, Luminance: lMap, Chroma: 127}
		assert.Equal(t, "64", perception.Chroma.Hex())
	})
}
