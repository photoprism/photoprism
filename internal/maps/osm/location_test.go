package osm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindLocation(t *testing.T) {
	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953

		l, err := FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, l.Cached)
		assert.Equal(t, 189675302, l.PlaceID)
		assert.Equal(t, "Fernsehturm Berlin", l.Name)
		assert.Equal(t, "10178", l.Address.Postcode)
		assert.Equal(t, "Berlin", l.Address.State)
		assert.Equal(t, "de", l.Address.CountryCode)
		assert.Equal(t, "Germany", l.Address.Country)

		l.PlaceID = 123456

		assert.Equal(t, 123456, l.PlaceID)

		cached, err := FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, cached.Cached)
		assert.Equal(t, 189675302, cached.PlaceID)
		assert.Equal(t, l.Name, cached.Name)
		assert.Equal(t, l.Address.Postcode, cached.Address.Postcode)
		assert.Equal(t, l.Address.State, cached.Address.State)
		assert.Equal(t, l.Address.CountryCode, cached.Address.CountryCode)
		assert.Equal(t, l.Address.Country, cached.Address.Country)
	})

	t.Run("BerlinMuseum", func(t *testing.T) {
		lat := 52.52057
		lng := 13.40889

		l, err := FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, l.Cached)
		assert.Equal(t, 48287001, l.PlaceID)
		assert.Equal(t, "Menschen Museum", l.Name)
		assert.Equal(t, "10178", l.Address.Postcode)
		assert.Equal(t, "Berlin", l.Address.State)
		assert.Equal(t, "de", l.Address.CountryCode)
		assert.Equal(t, "Germany", l.Address.Country)
	})

	t.Run("No Location", func(t *testing.T) {
		lat := 0.0
		lng := 0.0

		l, err := FindLocation(lat, lng)

		if err == nil {
			t.Fatal("err should not be nil")
		}

		assert.Equal(t, "osm: skipping lat 0.000000 / lng 0.000000", err.Error())
		assert.False(t, l.Cached)
	})
}
