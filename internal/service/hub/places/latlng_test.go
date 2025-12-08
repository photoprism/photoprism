package places

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatLng(t *testing.T) {
	lat := 52.5208
	lng := 13.40953

	t.Run("Local", func(t *testing.T) {
		l, err := LatLng(lat, lng, LocalLocale)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, LocalLocale, l.Locale)
		assert.Equal(t, "Berliner Fernsehturm", l.Name())
		assert.Equal(t, "Berlin", l.City())
		assert.Equal(t, "de", l.CountryCode())
	})
	t.Run("Englisb", func(t *testing.T) {
		l, err := LatLng(lat, lng, "en")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "en", l.Locale)
		assert.Equal(t, "Fernsehturm Berlin", l.Name())
		assert.Equal(t, "Berlin", l.City())
		assert.Equal(t, "de", l.CountryCode())
	})
	t.Run("MissingLng", func(t *testing.T) {
		l, err := LatLng(1, 0, LocalLocale)
		assert.Error(t, err, "places: skipping lat 0.000000, lng 0.000000")
		t.Log(l)
	})
	t.Run("MissingLat", func(t *testing.T) {
		l, err := LatLng(0, 1, LocalLocale)
		assert.Error(t, err, "places: skipping lat 0.000000, lng 0.000000")
		t.Log(l)
	})
	t.Run("Cached", func(t *testing.T) {
		location := Location{
			ID:          "1e95998417cc",
			LocLat:      52.51961810676184,
			LocLng:      13.40806264572578,
			LocName:     "TestLocation",
			LocStreet:   "",
			LocPostcode: "",
			LocCategory: "test",
			Place:       Place{PlaceID: "1"},
			Cached:      true,
		}

		_, err := LatLng(location.LocLat, location.LocLng, LocalLocale)

		if err != nil {
			t.Fatal(err)
		}

		cachedLoc, cacheErr := LatLng(location.LocLat, location.LocLng, LocalLocale)

		if cacheErr != nil {
			t.Fatal(cacheErr)
		}
		assert.Equal(t, true, cachedLoc.Cached)
	})
}
