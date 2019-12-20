package maps

import (
	"testing"

	"github.com/photoprism/photoprism/internal/maps/osm"
	"github.com/stretchr/testify/assert"
)

func TestLocation_Query(t *testing.T) {
	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953

		var l Location

		if err := l.Query(lat, lng); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Fernsehturm Berlin", l.LocTitle)
		assert.Equal(t, "Berlin, Germany", l.LocRegion)
	})
}

func TestLocation_QueryOpenStreetMap(t *testing.T) {
	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953

		var l Location

		if err := l.QueryOpenStreetMap(lat, lng); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Fernsehturm Berlin", l.LocTitle)
		assert.Equal(t, "Berlin, Germany", l.LocRegion)
	})
}

func TestLocation_OpenStreetMap(t *testing.T) {
	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 189675302, o.PlaceID)
		assert.Equal(t, "Fernsehturm Berlin", o.Name)
		assert.Equal(t, "10178", o.Address.Postcode)
		assert.Equal(t, "Berlin", o.Address.State)
		assert.Equal(t, "de", o.Address.CountryCode)
		assert.Equal(t, "Germany", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Fernsehturm Berlin", l.LocTitle)
		assert.Equal(t, "Berlin, Germany", l.LocRegion)
	})

	t.Run("SantaMonica", func(t *testing.T) {
		lat := 34.00909444444444
		lng := -118.49700833333334

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)
		assert.Equal(t, 79854991, o.PlaceID)
		assert.Equal(t, "Santa Monica Pier", o.Name)
		assert.Equal(t, "90401", o.Address.Postcode)
		assert.Equal(t, "California", o.Address.State)
		assert.Equal(t, "us", o.Address.CountryCode)
		assert.Equal(t, "United States of America", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Santa Monica Pier", l.LocTitle)
		assert.Equal(t, "Santa Monica, California, USA", l.LocRegion)
	})

	t.Run("AirportZurich", func(t *testing.T) {
		lat := 47.45401666666667
		lng := 8.557494444444446

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, 115198412, o.PlaceID)
		assert.Equal(t, "Dock A", o.Name)
		assert.Equal(t, "8302", o.Address.Postcode)
		assert.Equal(t, "Zurich", o.Address.State)
		assert.Equal(t, "ch", o.Address.CountryCode)
		assert.Equal(t, "Switzerland", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocTitle)
		assert.Equal(t, "Kloten, Zurich, Switzerland", l.LocRegion)
	})

	t.Run("AirportTegel", func(t *testing.T) {
		lat := 52.559864397033024
		lng := 13.28895092010498

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, 25410613, o.PlaceID)
		assert.Equal(t, "TGL", o.Name)
		assert.Equal(t, "13405", o.Address.Postcode)
		assert.Equal(t, "Berlin", o.Address.State)
		assert.Equal(t, "de", o.Address.CountryCode)
		assert.Equal(t, "Germany", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocTitle)
		assert.Equal(t, "Berlin, Germany", l.LocRegion)
	})

	t.Run("PinkBeach", func(t *testing.T) {
		lat := 35.26967222222222
		lng := 23.53711666666667

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, 119616937, o.PlaceID)
		assert.Equal(t, "Pink Beach", o.Name)
		assert.Equal(t, "", o.Address.Postcode)
		assert.Equal(t, "Crete", o.Address.State)
		assert.Equal(t, "gr", o.Address.CountryCode)
		assert.Equal(t, "Greece", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "8G757G9P+R5C", l.ID)
		assert.Equal(t, "Pink Beach", l.LocTitle)
		assert.Equal(t, "Crete, Greece", l.LocRegion)
	})

	t.Run("NewJersey", func(t *testing.T) {
		lat := 40.74290
		lng := -74.04862

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, 164551421, o.PlaceID)
		assert.Equal(t, "", o.Name)
		assert.Equal(t, "07307", o.Address.Postcode)
		assert.Equal(t, "New Jersey", o.Address.State)
		assert.Equal(t, "us", o.Address.CountryCode)
		assert.Equal(t, "United States", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "87G7PXV2+4G9", l.ID)
		assert.Equal(t, "", l.LocTitle)
		assert.Equal(t, "Jersey City, New Jersey, USA", l.LocRegion)
	})

	t.Run("SouthAfrica", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, 98820569, o.PlaceID)
		assert.Equal(t, "R411", o.Name)
		assert.Equal(t, "", o.Address.Postcode)
		assert.Equal(t, "Eastern Cape", o.Address.State)
		assert.Equal(t, "za", o.Address.CountryCode)
		assert.Equal(t, "South Africa", o.Address.Country)

		var l Location

		if err := l.OpenStreetMap(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "4GWF24FX+F5H", l.ID)
		assert.Equal(t, "Route R411", l.LocTitle)
		assert.Equal(t, "Eastern Cape, South Africa", l.LocRegion)
	})
}
