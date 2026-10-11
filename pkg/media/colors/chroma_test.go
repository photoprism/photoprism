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

func TestChroma_Value(t *testing.T) {
	t.Run("Monochrome", func(t *testing.T) {
		assert.Equal(t, int16(1), Chroma(0).Value())
		assert.Equal(t, int16(1), Chroma(-1).Value())
	})
	t.Run("NearlyGray", func(t *testing.T) {
		assert.Equal(t, int16(2), Chroma(1).Value())
	})
	t.Run("Fifteen", func(t *testing.T) {
		assert.Equal(t, int16(16), Chroma(15).Value())
	})
	t.Run("Maximum", func(t *testing.T) {
		assert.Equal(t, int16(99), Chroma(98).Value())
		assert.Equal(t, int16(100), Chroma(99).Value())
		assert.Equal(t, int16(100), Chroma(100).Value())
		assert.Equal(t, int16(100), Chroma(127).Value())
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
