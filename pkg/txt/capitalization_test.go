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
	t.Run("KwaZulu natal", func(t *testing.T) {
		assert.Equal(t, "KwaZulu Natal", Title("KwaZulu natal"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", UcFirst(""))
	})
}

func TestTitle(t *testing.T) {
	t.Run("Cour d'Honneur", func(t *testing.T) {
		assert.Equal(t, "Cour d'Honneur", Title("Cour d'Honneur"))
	})
	t.Run("Île-de-France", func(t *testing.T) {
		assert.Equal(t, "Île-de-France", Title("Île-de-France"))
	})
	t.Run("ile-de-France", func(t *testing.T) {
		assert.Equal(t, "Île-de-France", Title("ile-de-France"))
	})
	t.Run("BrowseYourLife", func(t *testing.T) {
		assert.Equal(t, "Browse Your Life in Pictures", Title("Browse your life in pictures"))
	})
	t.Run("German", func(t *testing.T) {
		assert.Equal(t, "Die Burg von oben gesehen.", Title("die burg von oben gesehen."))
		assert.Equal(t, "Die Katze ist auf dem Dach für viele nicht sichtbar!", Title("die katze ist auf dem dach für viele nicht sichtbar!"))
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
	t.Run("interpunction-1", func(t *testing.T) {
		assert.Equal(t, "This,,, Is !a ! a Very Strange Title....", Title("this,,, is !a ! a very strange title...."))
	})
	t.Run("interpunction-2", func(t *testing.T) {
		assert.Equal(t, "This Is a Not So Strange Title!", Title("This is a not so strange title!"))
	})
	t.Run("horse", func(t *testing.T) {
		assert.Equal(t, "A Horse Is Not a Cow :-)", Title("a horse is not a cow :-)"))
	})
	t.Run("NewYears", func(t *testing.T) {
		assert.Equal(t, "Boston New Year's", Title("boston new year's"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Empty(t, Title(""))
	})
	t.Run("NYC", func(t *testing.T) {
		assert.Equal(t, "NYC, NY - LonDon, UK - NYC, NY and London, UK.", Title("NYC, NY - LonDon, UK - Nyc, Ny and London, Uk."))
	})
}
