package places

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/geo/s2"
)

func TestCell(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l, err := Cell(id, DefaultLocale)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, l.Cached)
		assert.Equal(t, "Berlin", l.City())
		assert.Equal(t, "de", l.CountryCode())
	})
	t.Run("MissingId", func(t *testing.T) {
		l, err := Cell("", DefaultLocale)
		assert.Error(t, err, "places: invalid location id ")
		t.Log(l)
	})
	t.Run("WrongId", func(t *testing.T) {
		l, err := Cell("2", DefaultLocale)
		assert.Error(t, err, "places: skipping lat 0.000000, lng 0.000000")
		t.Log(l)
	})
	t.Run("ShortId", func(t *testing.T) {
		l, err := Cell("ab", DefaultLocale)
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

		l, err := Cell(location.ID, DefaultLocale)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, l.Cached)
		l2, err2 := Cell("1e95998417cc", DefaultLocale)

		if err2 != nil {
			t.Fatal(err2)
		}
		assert.Equal(t, true, l2.Cached)
	})
}
