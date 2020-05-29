package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsNumber(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, ContainsNumber("f3abcde"))
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, ContainsNumber("abcd"))
	})
}

func TestIsSeparator(t *testing.T) {
	t.Run("rune A", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('A'))
	})
	t.Run("rune 99", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('9'))
	})
	t.Run("rune /", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('/'))
	})
	t.Run("rune \\", func(t *testing.T) {
		assert.Equal(t, true, isSeparator('\\'))
	})
	t.Run("rune ♥ ", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('♥'))
	})
	t.Run("rune  space", func(t *testing.T) {
		assert.Equal(t, true, isSeparator(' '))
	})
	t.Run("rune '", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('\''))
	})
	t.Run("rune ý", func(t *testing.T) {
		assert.Equal(t, false, isSeparator('ý'))
	})
}

func TestUcFirst(t *testing.T) {
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Photo-lover", UcFirst("photo-lover"))
	})
	t.Run("cat", func(t *testing.T) {
		assert.Equal(t, "Cat", UcFirst("Cat"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", UcFirst(""))
	})
}

func TestTitle(t *testing.T) {
	t.Run("Browse your life in pictures", func(t *testing.T) {
		assert.Equal(t, "Browse Your Life In Pictures", Title("Browse your life in pictures"))
	})
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Photo-Lover", Title("photo-lover"))
	})
}

func TestTitleFromFileName(t *testing.T) {
	t.Run("Browse your life in pictures", func(t *testing.T) {
		assert.Equal(t, "Browse Your Life In Pictures", TitleFromFileName("Browse your life in pictures"))
	})
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Photo Lover", TitleFromFileName("photo-lover"))
	})
	t.Run("BRIDGE in nyc", func(t *testing.T) {
		assert.Equal(t, "Bridge In NYC", TitleFromFileName("BRIDGE in nyc"))
	})
	t.Run("phil unveils iphone, ipad, imac or macbook 11 pro and max", func(t *testing.T) {
		assert.Equal(t, "Phil Unveils iPhone iPad iMac or MacBook Pro and Max", TitleFromFileName("phil unveils iphone, ipad, imac or macbook 11 pro and max"))
	})
	t.Run("IMG_4568", func(t *testing.T) {
		assert.Equal(t, "", TitleFromFileName("IMG_4568"))
	})
	t.Run("queen-city-yacht-club--toronto-island_7999432607_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Queen City Yacht Club / Toronto Island", TitleFromFileName("queen-city-yacht-club--toronto-island_7999432607_o.jpg"))
	})
	t.Run("tim-robbins--tiff-2012_7999233420_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Tim Robbins / TIFF", TitleFromFileName("tim-robbins--tiff-2012_7999233420_o.jpg"))
	})
	t.Run("20200102-204030-Berlin-Germany-2020-3h4.jpg", func(t *testing.T) {
		assert.Equal(t, "Berlin Germany", TitleFromFileName("20200102-204030-Berlin-Germany-2020-3h4.jpg"))
	})
	t.Run("changing-of-the-guard--buckingham-palace_7925318070_o.jpg", func(t *testing.T) {
		assert.Equal(t, "Changing Of The Guard / Buckingham Palace", TitleFromFileName("changing-of-the-guard--buckingham-palace_7925318070_o.jpg"))
	})
}

func TestBool(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, true, Bool("Browse your life in pictures"))
	})
	t.Run("no", func(t *testing.T) {
		assert.Equal(t, false, Bool("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, false, Bool("false"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, false, Bool(""))
	})
}
