package maps

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/s2"
	"github.com/stretchr/testify/assert"
)

func TestLocation_QueryPlaces(t *testing.T) {
	t.Run("U Berliner Rathaus", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", "", []string{})

		if err := l.QueryPlaces(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Mitte, Berlin, Germany", l.LocLabel)
	})
}

func TestLocation_Unknown(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		lat := 0.0
		lng := 0.0
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", "", []string{})

		assert.Equal(t, true, l.Unknown())
	})
	t.Run("false", func(t *testing.T) {
		lat := -31.976301666666668
		lng := 29.148046666666666
		id := s2.Token(lat, lng)

		l := NewLocation(id, "", "", "", "", "", "", "", "", []string{})

		assert.Equal(t, false, l.Unknown())
	})
}

func TestLocation_S2Token(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := NewLocation("123", "Indian ocean", "", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "123", l.S2Token())
	})
}

func TestLocation_PrefixedToken(t *testing.T) {
	t.Run("123", func(t *testing.T) {
		l := NewLocation("123", "Indian ocean", "", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, s2.TokenPrefix+"123", l.PrefixedToken())
	})
}

func TestLocation_Name(t *testing.T) {
	t.Run("Christkindlesmarkt", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "Christkindlesmarkt", l.Name())
	})
}

func TestLocation_City(t *testing.T) {
	t.Run("Nürnberg", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "Nürnberg", l.City())
	})
}

func TestLocation_State(t *testing.T) {
	t.Run("Bayern", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "Bayern", l.State())
	})
}

func TestLocation_Category(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "test", l.Category())
	})
}

func TestLocation_Source(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "", "Nürnberg", "", "Bayern", "de", "source", []string{})

		assert.Equal(t, "source", l.Source())
	})
}

func TestLocation_Place(t *testing.T) {
	t.Run("test-label", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "", "test-label", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "test-label", l.Label())
	})
}

func TestLocation_CountryCode(t *testing.T) {
	t.Run("de", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "de", l.CountryCode())
	})
}

func TestLocation_CountryName(t *testing.T) {
	t.Run("Germany", func(t *testing.T) {
		l := NewLocation("", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "", "Bayern", "de", "", []string{})

		assert.Equal(t, "Germany", l.CountryName())
	})
}

func TestLocation_QueryApi(t *testing.T) {
	l := NewLocation("3", "Christkindlesmarkt", "test", "test-label", "Nürnberg", "", "Bayern", "de", "", []string{})
	t.Run("xxx", func(t *testing.T) {
		api := l.QueryApi("xxx")
		assert.Error(t, api, "maps: reverse lookup disabled")
	})
	t.Run("places", func(t *testing.T) {
		api := l.QueryApi("places")
		assert.Error(t, api)
	})

}
