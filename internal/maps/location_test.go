package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/s2"
)

func TestLocation_QueryPlaces(t *testing.T) {
	t.Run("BerlinerRathaus", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l := Location{ID: id}

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}

		t.Logf("BerlinerRathaus: %#v", l)

		assert.Equal(t, "Mitte, Berlin, Germany", l.LocLabel)
	})
	t.Run("BerlinerFernsehturm", func(t *testing.T) {
		lat := 52.5208
		lng := 13.40953
		id := s2.Token(lat, lng)

		l := Location{ID: id}

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}
		t.Logf("BerlinerFernsehturm: %#v", l)

		assert.Equal(t, "Berliner Fernsehturm", l.LocName)
		assert.Equal(t, "Berlin", l.LocState)
		assert.Equal(t, "Berlin", l.LocCity)
		assert.Equal(t, "Panoramastraße", l.LocStreet)
		assert.Equal(t, "10178", l.LocPostcode)
		assert.Equal(t, "Mitte", l.LocDistrict)
		assert.Equal(t, "Mitte, Berlin, Germany", l.LocLabel)
	})
	t.Run("NorthAtlanticOcean", func(t *testing.T) {
		id := "0a3c25fcffad"

		l := Location{ID: id}

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", l.LocName)
		assert.Equal(t, "", l.LocState)
		assert.Equal(t, "ocean", l.LocCategory)
		assert.Equal(t, "North Atlantic Ocean", l.LocDistrict)
		assert.Equal(t, "North Atlantic Ocean", l.LocLabel)
	})
	t.Run("SouthPacificOcean", func(t *testing.T) {
		id := "9aa986feefb4"

		l := Location{ID: id}

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Puerto Velasco Ibarra", l.LocName)
		assert.Equal(t, "", l.LocCategory)
		assert.Equal(t, "Galápagos", l.LocState)
		assert.Equal(t, "", l.LocDistrict)
		assert.Equal(t, "ec", l.LocCountry)
		assert.Equal(t, "Galápagos, Ecuador", l.LocLabel)
		assert.Equal(t, "places", l.LocSource)
	})
}

func TestLocation_Unknown(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		lat := 0.0
		lng := 0.0
		id := s2.Token(lat, lng)

		l := Location{ID: id}

		assert.Equal(t, true, l.Unknown())
	})
	t.Run("false", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666
		id := s2.Token(lat, lng)

		l := Location{ID: id}

		assert.Equal(t, false, l.Unknown())
	})
}

func TestLocation_S2Token(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := Location{ID: "123"}

		assert.Equal(t, "123", l.S2Token())
	})
}

func TestLocation_PrefixedToken(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := Location{ID: "123"}

		assert.Equal(t, s2.TokenPrefix+"123", l.PrefixedToken())
	})
}

func TestLocation_Name(t *testing.T) {
	t.Run("Christkindlesmarkt", func(t *testing.T) {
		l := Location{ID: "123", LocName: "Christkindlesmarkt"}

		assert.Equal(t, "Christkindlesmarkt", l.Name())
	})
}

func TestLocation_City(t *testing.T) {
	t.Run("Nürnberg", func(t *testing.T) {
		l := Location{ID: "123", LocCity: "Nürnberg"}

		assert.Equal(t, "Nürnberg", l.City())
	})
}

func TestLocation_State(t *testing.T) {
	t.Run("Bayern", func(t *testing.T) {
		l := Location{ID: "123", LocState: "Bayern"}

		assert.Equal(t, "Bayern", l.State())
	})
}

func TestLocation_Category(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		l := Location{ID: "123", LocCategory: "test"}

		assert.Equal(t, "test", l.Category())
	})
}

func TestLocation_Source(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		l := Location{ID: "123", LocSource: "mySource"}

		assert.Equal(t, "mySource", l.Source())
	})
}

func TestLocation_Place(t *testing.T) {
	t.Run("test-label", func(t *testing.T) {
		l := Location{ID: "123", LocLabel: "test-label"}

		assert.Equal(t, "test-label", l.Label())
	})
}

func TestLocation_CountryCode(t *testing.T) {
	t.Run("de", func(t *testing.T) {
		l := Location{ID: "123", LocCountry: "de"}

		assert.Equal(t, "de", l.CountryCode())
	})
}

func TestLocation_CountryName(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		l := Location{ID: "123", LocCountry: "de"}

		assert.Equal(t, "Germany", l.CountryName())
	})
}

func TestLocation_QueryApi(t *testing.T) {
	l := Location{ID: "3", LocCountry: "de"}

	t.Run("xxx", func(t *testing.T) {
		api := l.QueryApi("xxx")
		assert.Error(t, api, "maps: reverse lookup disabled")
	})
	t.Run("places", func(t *testing.T) {
		api := l.QueryApi("places")
		assert.Error(t, api)
	})

}
