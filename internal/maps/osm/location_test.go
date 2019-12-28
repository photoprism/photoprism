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
		assert.Equal(t, "Fernsehturm Berlin", l.LocName)
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
		assert.Equal(t, l.LocName, cached.LocName)
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
		assert.Equal(t, "Menschen Museum", l.LocName)
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

		assert.Equal(t, "osm: skipping lat 0.000000, lng 0.000000", err.Error())
		assert.False(t, l.Cached)
	})
}

func TestOSM_State(t *testing.T) {
	t.Run("Berlin", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Berlin", l.State())
	})
}

/*
func TestOSM_City(t *testing.T) {
	t.Run("Berlin", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln", Town: "Hamburg", Village: "Köln", County: "Wiesbaden"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Berlin", l.City())
	})
	t.Run("Hamburg", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln", Town: "Hamburg", Village: "Köln", County: "Wiesbaden"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Hamburg", l.City())
	})
	t.Run("Köln", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln", Town: "", Village: "Köln", County: "Wiesbaden"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Köln", l.City())
	})
	t.Run("Wiesbaden", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln", Town: "", Village: "", County: "Wiesbaden"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Wiesbaden", l.City())
	})
	t.Run("Frankfurt", func(t *testing.T) {
		a := Address{CountryCode: "de", City: "Frankfurt", State: "", HouseNumber: "63", Suburb: "Neukölln", Town: "", Village: "", County: ""}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Frankfurt", l.City())
	})
}
*/
func TestOSM_Suburb(t *testing.T) {
	t.Run("Neukölln", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "Neukölln", l.Suburb())
	})
}

func TestOSM_CountryCode(t *testing.T) {
	t.Run("de", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, "de", l.CountryCode())
	})
}

func TestOSM_Latitude(t *testing.T) {
	t.Run("52.5208", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, 52.5208, l.Latitude())
	})
}

func TestOSM_Longitude(t *testing.T) {
	t.Run("13.40953", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "dipslay name", Address: a}
		assert.Equal(t, 13.40953, l.Longitude())
	})
}

func TestOSM_Keywords(t *testing.T) {
	t.Run("cat", func(t *testing.T) {

		a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
		l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "cat", Address: a}
		assert.Equal(t, []string{"cat"}, l.Keywords())
	})
}

func TestOSM_Source(t *testing.T) {

	a := Address{CountryCode: "de", City: "Berlin", State: "Berlin", HouseNumber: "63", Suburb: "Neukölln"}
	l := &Location{LocCategory: "natural", LocLat: "52.5208", LocLng: "13.40953", LocName: "Nice title", LocType: "hill", LocDisplayName: "cat", Address: a}
	assert.Equal(t, "osm", l.Source())
}
