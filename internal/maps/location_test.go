package maps

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/hub/places"
	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/stretchr/testify/assert"
)

func TestLocation_QueryPlaces(t *testing.T) {
	t.Run("U Berliner Rathaus", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", []string{})

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Berlin, Germany", l.LocLabel)
	})
}

func TestLocation_Assign(t *testing.T) {
	t.Run("Italy", func(t *testing.T) {
		id := "47786b2bed37"

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Comici I", o.Name())
		assert.Equal(t, "Trentino-Alto Adige", o.State())
		assert.Equal(t, "it", o.CountryCode())

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Comici I", l.LocName)
		assert.Equal(t, "Santa Cristina Gherdëina, Trentino-Alto Adige, Italy", l.LocLabel)
		assert.IsType(t, []string{}, l.Keywords())
		assert.Equal(t, "christina, cristina, gröden, santa, südtirol, valgardena", l.KeywordString())
	})

	t.Run("BerlinFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953
		id := s2.Token(lat, lng)

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Berliner Fernsehturm", o.Name())
		assert.Equal(t, "Berlin", o.City())
		assert.Equal(t, "Berlin", o.State())
		assert.Equal(t, "de", o.CountryCode())

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Berliner Fernsehturm", l.LocName)
		assert.Equal(t, "Berlin, Germany", l.LocLabel)
		assert.IsType(t, []string{}, l.Keywords())
		assert.Equal(t, "", l.KeywordString())
	})

	t.Run("SantaMonica", func(t *testing.T) {
		lat := 34.00909444444444
		lng := -118.49700833333334
		id := s2.Token(lat, lng)

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)
		assert.Equal(t, "California", o.State())
		assert.Equal(t, "us", o.CountryCode())

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Santa Monica, California, USA", l.LocLabel)
	})

	t.Run("AirportZurich", func(t *testing.T) {
		lat := 47.45401666666667
		lng := 8.557494444444446
		id := s2.Token(lat, lng)

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocName)
		assert.Equal(t, "Kloten, Zürich, Switzerland", l.LocLabel)
	})

	t.Run("AirportTegel", func(t *testing.T) {
		lat := 52.559864397033024
		lng := 13.28895092010498
		id := s2.Token(lat, lng)

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Airport", l.LocName)
		assert.Equal(t, "Berlin, Germany", l.LocLabel)
	})

	t.Run("SouthAfrica", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666
		id := s2.Token(lat, lng)

		o, err := places.FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, o.Cached)

		assert.Equal(t, "", o.Name())
		assert.Equal(t, "Eastern Cape", o.State())
		assert.Equal(t, "za", o.CountryCode())

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.True(t, strings.HasPrefix(l.ID, "1e5e4205"))
		assert.Equal(t, "", l.LocName)
		assert.Equal(t, "Eastern Cape, South Africa", l.LocLabel)
	})

	t.Run("ocean", func(t *testing.T) {
		lat := -21.976301666666668
		lng := 49.148046666666666
		id := s2.Token(lat, lng)
		// log.Printf("ID: %s", id)
		o, err := places.FindLocation(id)

		// log.Printf("Output: %+v", o)

		if err != nil {
			t.Fatal(err)
		}

		var l Location

		if err := l.Assign(o); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Indian Ocean", l.LocName)
		assert.Equal(t, "", l.LocCategory)
		assert.Equal(t, "Unknown", l.LocCity)
		assert.Equal(t, "zz", l.LocCountry)
	})
}

func TestLocation_Unknown(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		lat := 0.0
		lng := 0.0
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", []string{})

		assert.Equal(t, true, l.Unknown())
	})
	t.Run("false", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", []string{})

		assert.Equal(t, false, l.Unknown())
	})
}

func TestLocation_place(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		lat := 0.0
		lng := 0.0
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", []string{})

		assert.Equal(t, "Unknown", l.label())
	})
	t.Run("Nürnberg, Bayern, Germany", func(t *testing.T) {
		l := NewLocation("", "", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "Unknown", l.label())
	})
	t.Run("Freiburg im Breisgau, BW, Germany", func(t *testing.T) {
		l := NewLocation("47911b1a4f84", "", "", "Freiburg im Breisgau, BW, Germany", "Freiburg im Breisgau", "BW", "de", "", []string{})

		assert.Equal(t, "Freiburg im Breisgau, Baden-Württemberg, Germany", l.label())
	})
	t.Run("Sevilla, ES, Spain", func(t *testing.T) {
		l := NewLocation("0d126c12219c", "", "", "Sevilla, ES, Spain", "Sevilla", "ES", "es", "", []string{})

		assert.Equal(t, "Sevilla, Spain", l.label())
	})
	t.Run("Guarapari, ES, Brazil", func(t *testing.T) {
		l := NewLocation("00b85797fdbc", "", "", "Guarapari, ES, Brazil", "Guarapari", "ES", "br", "", []string{})

		assert.Equal(t, "Guarapari, Espírito Santo, Brazil", l.label())
	})
	t.Run("Porto Novo, PT, Portugal", func(t *testing.T) {
		l := NewLocation("0d1f30bb5564", "", "", "", "Porto Novo", "PT", "pt", "", []string{})

		assert.Equal(t, "Porto Novo, Portugal", l.label())
	})
}

func TestLocation_S2Token(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := NewLocation("123", "Indian ocean", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "123", l.S2Token())
	})
}

func TestLocation_PrefixedToken(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := NewLocation("123", "Indian ocean", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, s2.TokenPrefix+"123", l.PrefixedToken())
	})
}

func TestLocation_Name(t *testing.T) {
	t.Run("Christkindlesmarkt", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "Christkindlesmarkt", l.Name())
	})
}

func TestLocation_City(t *testing.T) {
	t.Run("Nürnberg", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "Nürnberg", l.City())
	})
}

func TestLocation_State(t *testing.T) {
	t.Run("Bayern", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "Bayern", l.State())
	})
}

func TestLocation_Category(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "test", l.Category())
	})
}

func TestLocation_Source(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "Bayern", "de", "source", []string{})

		assert.Equal(t, "source", l.Source())
	})
}

func TestLocation_Place(t *testing.T) {
	t.Run("test-label", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "test-label", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "test-label", l.Label())
	})
}

func TestLocation_CountryCode(t *testing.T) {
	t.Run("de", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "de", l.CountryCode())
	})
}

func TestLocation_CountryName(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "Bayern", "de", "", []string{})

		assert.Equal(t, "Germany", l.CountryName())
	})
}

func TestLocation_QueryApi(t *testing.T) {
	l := NewLocation("3", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "Bayern", "de", "", []string{})
	t.Run("xxx", func(t *testing.T) {
		api := l.QueryApi("xxx")
		assert.Error(t, api, "maps: reverse lookup disabled")
	})
	t.Run("osm", func(t *testing.T) {
		api := l.QueryApi("osm")
		assert.Error(t, api)
	})
	t.Run("places", func(t *testing.T) {
		api := l.QueryApi("places")
		assert.Error(t, api)
	})

}
