package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	t.Run("BrowseYourLife", func(t *testing.T) {
		assert.Equal(t, "Browse Your Life In Pictures", Title("Browse your life in pictures"))
	})
	t.Run("PhotoLover", func(t *testing.T) {
		assert.Equal(t, "Photo-Lover", Title("photo-lover"))
	})
	t.Run("NaomiWatts", func(t *testing.T) {
		assert.Equal(t, "Naomi Watts / Ewan Mcgregor / The Impossible / TIFF", Title(" /Naomi watts / Ewan Mcgregor / the   Impossible /   TIFF  "))
	})
	t.Run("Penguin", func(t *testing.T) {
		assert.Equal(t, "A Boulders Penguin Colony / Simon's Town / 2013", Title("A Boulders Penguin Colony /// Simon's Town / 2013 "))
	})
	t.Run("AirportBer", func(t *testing.T) {
		assert.Equal(t, "Around the Terminal / Airport BER", Title("Around  the Terminal  / Airport Ber"))
	})
	t.Run("KwaZulu-Natal", func(t *testing.T) {
		assert.Equal(t, "KwaZulu-Natal", Title("KwaZulu-Natal"))
	})
	t.Run("testAddLabel", func(t *testing.T) {
		assert.Equal(t, "TestAddLabel", Title("testAddLabel"))
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
		assert.Equal(t, "Changing of the Guard / Buckingham Palace", TitleFromFileName("changing-of-the-guard--buckingham-palace_7925318070_o.jpg"))
	})
}
