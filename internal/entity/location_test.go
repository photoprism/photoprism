package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocation(t *testing.T) {
	t.Run("new label", func(t *testing.T) {
		l := NewLocation(1, 1)
		l.LocCategory = "restaurant"
		l.LocName = "LocationName"
		l.Place = &PlaceFixtureZinkwazi
		l.LocSource = "places"

		assert.Equal(t, "restaurant", l.Category())
		assert.Equal(t, false, l.NoCategory())
		assert.Equal(t, false, l.Unknown())
		assert.Equal(t, "LocationName", l.Name())
		assert.Equal(t, false, l.NoName())
		assert.Equal(t, "KwaDukuza, KwaZulu-Natal, South Africa", l.Label())
		assert.Equal(t, "KwaDukuza", l.City())
		assert.Equal(t, false, l.LongCity())
		assert.Equal(t, false, l.CityContains("xxx"))
		assert.Equal(t, false, l.NoCity())
		assert.Equal(t, "KwaZulu-Natal", l.State())
		assert.Equal(t, false, l.NoState())
		assert.Equal(t, "za", l.CountryCode())
		assert.Equal(t, "South Africa", l.CountryName())
		assert.Equal(t, "places", l.Source())
		assert.Equal(t, "africa", l.Notes())
	})
}

func TestLocation_Keywords(t *testing.T) {
	t.Run("location with place", func(t *testing.T) {
		r := LocationFixtureMexico.Keywords()
		assert.Equal(t, []string{"adosada", "ancient", "mexico", "platform", "pyramid", "teotihuacán", "tourism"}, r)
	})
	t.Run("location without place", func(t *testing.T) {
		r := LocationFixtureCaravanPark.Keywords()
		assert.Nil(t, r)
	})
}

func TestLocation_Find(t *testing.T) {
	t.Run("place in db", func(t *testing.T) {
		r := LocationFixtureMexico.Find("")
		assert.Nil(t, r)
	})
	t.Run("invalid api", func(t *testing.T) {
		l := NewLocation(2, 1)
		err := l.Find("")
		assert.Equal(t, "maps: reverse lookup disabled", err.Error())
	})
}
