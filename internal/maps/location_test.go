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

		l := NewLocation(lat, lng)

		if err := l.Query(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Fernsehturm Berlin", l.LocName)
		assert.Equal(t, "Berlin, Germany", l.LocPlace)
	})
}

func TestLocation_Assign(t *testing.T) {
	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 189675302, o.PlaceID)
		assert.Equal(t, "Fernsehturm Berlin", o.LocName)
		assert.Equal(t, "10178", o.Address.Postcode)
		assert.Equal(t, "Berlin", o.Address.State)
		assert.Equal(t, "de", o.Address.CountryCode)
		assert.Equal(t, "Germany", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Fernsehturm Berlin", l.LocName)
		assert.Equal(t, "Berlin, Germany", l.LocPlace)
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
		assert.Equal(t, "Santa Monica Pier", o.LocName)
		assert.Equal(t, "90401", o.Address.Postcode)
		assert.Equal(t, "California", o.Address.State)
		assert.Equal(t, "us", o.Address.CountryCode)
		assert.Equal(t, "United States of America", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Santa Monica Pier", l.LocName)
		assert.Equal(t, "Santa Monica, California, USA", l.LocPlace)
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
		assert.Equal(t, "Dock A", o.LocName)
		assert.Equal(t, "8302", o.Address.Postcode)
		assert.Equal(t, "Zurich", o.Address.State)
		assert.Equal(t, "ch", o.Address.CountryCode)
		assert.Equal(t, "Switzerland", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocName)
		assert.Equal(t, "Kloten, Zurich, Switzerland", l.LocPlace)
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
		assert.Equal(t, "TGL", o.LocName)
		assert.Equal(t, "13405", o.Address.Postcode)
		assert.Equal(t, "Berlin", o.Address.State)
		assert.Equal(t, "de", o.Address.CountryCode)
		assert.Equal(t, "Germany", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocName)
		assert.Equal(t, "Berlin, Germany", l.LocPlace)
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
		assert.Equal(t, "Pink Beach", o.LocName)
		assert.Equal(t, "", o.Address.Postcode)
		assert.Equal(t, "Crete", o.Address.State)
		assert.Equal(t, "gr", o.Address.CountryCode)
		assert.Equal(t, "Greece", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint64(0x149ce78540000000), l.ID)
		assert.Equal(t, "Pink Beach", l.LocName)
		assert.Equal(t, "Chrisoskalitissa, Crete, Greece", l.LocPlace)
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
		assert.Equal(t, "", o.LocName)
		assert.Equal(t, "07307", o.Address.Postcode)
		assert.Equal(t, "New Jersey", o.Address.State)
		assert.Equal(t, "us", o.Address.CountryCode)
		assert.Equal(t, "United States", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint64(0x9c25741c0000000), l.ID)
		assert.Equal(t, "", l.LocName)
		assert.Equal(t, "Jersey City, New Jersey, USA", l.LocPlace)
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
		assert.Equal(t, "R411", o.LocName)
		assert.Equal(t, "", o.Address.Postcode)
		assert.Equal(t, "Eastern Cape", o.Address.State)
		assert.Equal(t, "za", o.Address.CountryCode)
		assert.Equal(t, "South Africa", o.Address.Country)

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint64(0x1e5e4205c0000000), l.ID)
		assert.Equal(t, "R411", l.LocName)
		assert.Equal(t, "Eastern Cape, South Africa", l.LocPlace)
	})

	t.Run("Unknown", func(t *testing.T) {
		lat := -21.976301666666668
		lng := 49.148046666666666

		o, err := osm.FindLocation(lat, lng)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		var l Location

		assert.Error(t, l.Assign(o))
		assert.Equal(t, "unknown", l.LocCategory)
	})
}

func TestLocation_Unknown(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		lat := 0.0
		lng := 0.0

		l := NewLocation(lat, lng)

		assert.Equal(t, true, l.Unknown())
	})
	t.Run("false", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := NewLocation(lat, lng)

		assert.Equal(t, false, l.Unknown())
	})
}

func TestLocation_place(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		lat := 0.0
		lng := 0.0

		l := NewLocation(lat, lng)

		assert.Equal(t, "Unknown", l.place())
	})
	t.Run("Nürnberg, Bayern, Germany", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern"}

		assert.Equal(t, "Nürnberg, Bayern, Germany", l.place())
	})
}

func TestLocation_Latitude(t *testing.T) {
	t.Run("-31.976301666666668", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern"}

		assert.Equal(t, -31.976301666666668, l.Latitude())
	})
}

func TestLocation_Longitude(t *testing.T) {
	t.Run("29.148046666666666", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern"}

		assert.Equal(t, 29.148046666666666, l.Longitude())
	})
}

func TestLocation_Name(t *testing.T) {
	t.Run("Christkindlesmarkt", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt"}

		assert.Equal(t, "Christkindlesmarkt", l.Name())
	})
}

func TestLocation_City(t *testing.T) {
	t.Run("Nürnberg", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt"}

		assert.Equal(t, "Nürnberg", l.City())
	})
}

func TestLocation_Suburb(t *testing.T) {
	t.Run("Hauptmarkt", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "Hauptmarkt", l.Suburb())
	})
}

func TestLocation_State(t *testing.T) {
	t.Run("Bayern", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "Bayern", l.State())
	})
}

func TestLocation_Category(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocCategory: "test", LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "test", l.Category())
	})
}

func TestLocation_Source(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocCategory: "test", LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt", LocSource: "source"}

		assert.Equal(t, "source", l.Source())
	})
}

func TestLocation_Place(t *testing.T) {
	t.Run("test-place", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocCategory: "test", LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocPlace: "test-place", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "test-place", l.Place())
	})
}

func TestLocation_CountryCode(t *testing.T) {
	t.Run("de", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocCategory: "test", LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocPlace: "test-place", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "de", l.CountryCode())
	})
}

func TestLocation_CountryName(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666

		l := &Location{LocCategory: "test", LocLat: lat, LocLng: lng, LocCountry: "de", LocCity: "Nürnberg", LocPlace: "test-place", LocState: "Bayern", LocName: "Christkindlesmarkt", LocSuburb: "Hauptmarkt"}

		assert.Equal(t, "Germany", l.CountryName())
	})
}
