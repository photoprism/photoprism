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
		assert.Equal(t, "Browse Your Life in Pictures", Title("Browse your life in pictures"))
	})
	t.Run("PhotoLover", func(t *testing.T) {
		assert.Equal(t, "Photo-Lover", Title("photo-lover"))
	})
	t.Run("NaomiWatts", func(t *testing.T) {
		assert.Equal(t, "Naomi Watts / Ewan McGregor / The Impossible / TIFF", Title(" /Naomi watts / Ewan Mcgregor / the   Impossible /   TIFF  "))
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
	t.Run("photoprism", func(t *testing.T) {
		assert.Equal(t, "PhotoPrism", Title("photoprism"))
	})
	t.Run("youtube", func(t *testing.T) {
		assert.Equal(t, "YouTube", Title("youtube"))
	})
	t.Run("interpunctio-1", func(t *testing.T) {
		assert.Equal(t, "This,,, Is !a ! a Very Strange Title....", Title("this,,, is !a ! a very strange title...."))
	})
	t.Run("interpunctio-2", func(t *testing.T) {
		assert.Equal(t, "This Is a Not So Strange Title!", Title("This is a not so strange title!"))
	})
	t.Run("horse", func(t *testing.T) {
		assert.Equal(t, "A Horse Is Not a Cow :-)", Title("a horse is not a cow :-)"))
	})
}

func TestTitleFromFileName(t *testing.T) {
	t.Run("photoprism", func(t *testing.T) {
		assert.Equal(t, "PhotoPrism: Browse Your Life in Pictures", TitleFromFileName("photoprism: Browse your life in pictures"))
	})
	t.Run("dash", func(t *testing.T) {
		assert.Equal(t, "Photo Lover", TitleFromFileName("photo-lover"))
	})
	t.Run("nyc", func(t *testing.T) {
		assert.Equal(t, "Bridge in, or by, NYC", TitleFromFileName("BRIDGE in, or by, nyc"))
	})
	t.Run("apple", func(t *testing.T) {
		assert.Equal(t, "Phil Unveils iPhone, iPad, iPod, 'airpods', Airpod, AirPlay, iMac or MacBook", TitleFromFileName("phil unveils iphone, ipad, ipod, 'airpods', airpod, airplay, imac or macbook 11 pro and max"))
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
