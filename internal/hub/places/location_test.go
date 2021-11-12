package places

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/s2"
)

func TestFindLocation(t *testing.T) {
	t.Run("U Berliner Rathaus", func(t *testing.T) {
		lat := 52.51961810676184
		lng := 13.40806264572578
		id := s2.Token(lat, lng)

		l, err := FindLocation(id)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, l.Cached)
		assert.Equal(t, "Berlin", l.City())
		assert.Equal(t, "de", l.CountryCode())
	})
	t.Run("wrong id", func(t *testing.T) {
		l, err := FindLocation("2")
		assert.Error(t, err, "places: skipping lat 0.000000, lng 0.000000")
		t.Log(l)
	})
	t.Run("short id", func(t *testing.T) {
		l, err := FindLocation("ab")
		assert.Error(t, err, "places: skipping lat 0.000000, lng 0.000000")
		t.Log(l)
	})
	t.Run("invalid id", func(t *testing.T) {
		l, err := FindLocation("")
		assert.Error(t, err, "places: invalid location id ")
		t.Log(l)
	})
	t.Run("cached true", func(t *testing.T) {
		var p = NewPlace("1", "", "", "", "", "de", "")
		location := NewLocation("1e95998417cc", 52.51961810676184, 13.40806264572578, "TestLocation", "test", p, true)
		l, err := FindLocation(location.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, false, l.Cached)
		l2, err2 := FindLocation("1e95998417cc")

		if err2 != nil {
			t.Fatal(err2)
		}
		assert.Equal(t, true, l2.Cached)
	})
}

func TestLocationGetters(t *testing.T) {
	var p = NewPlace("1", "testLabel", "Berlin", "", "Berlin", "de", "foobar")
	location := NewLocation("1e95998417cc", 52.51961810676184, 13.40806264572578, "TestLocation", "test", p, true)
	t.Run("wrong id", func(t *testing.T) {
		assert.Equal(t, "1e95998417cc", location.CellID())
		assert.Equal(t, "TestLocation", location.Name())
		assert.Equal(t, "test", location.Category())
		assert.Equal(t, "testLabel", location.Label())
		assert.Equal(t, "Berlin", location.State())
		assert.Equal(t, "de", location.CountryCode())
		assert.Equal(t, "Berlin", location.City())
		assert.Equal(t, 52.51961810676184, location.Latitude())
		assert.Equal(t, 13.40806264572578, location.Longitude())
		assert.Equal(t, "places", location.Source())
		assert.Equal(t, []string{"foobar"}, location.Keywords())
	})
}

func TestLocation_State(t *testing.T) {
	t.Run("Washington", func(t *testing.T) {
		var p = NewPlace("549ed22c0434", "Seattle, WA", "Seattle", "", "WA", "us", "")
		location := NewLocation("54903ee07f74", 47.6129432, -122.4821475, "", "", p, true)
		assert.Equal(t, "54903ee07f74", location.CellID())
		assert.Equal(t, "Seattle, WA", location.Label())
		assert.Equal(t, "Washington", location.State())
		assert.Equal(t, "us", location.CountryCode())
		assert.Equal(t, "Seattle", location.City())
		assert.Equal(t, "places", location.Source())
	})
}

func TestLocation_District(t *testing.T) {
	t.Run("Washington", func(t *testing.T) {
		var p = NewPlace("549ed22c0434", "Seattle, WA", "Seattle", "Foo", "WA", "us", "")
		location := NewLocation("54903ee07f74", 47.6129432, -122.4821475, "", "", p, true)
		assert.Equal(t, "54903ee07f74", location.CellID())
		assert.Equal(t, "Seattle, WA", location.Label())
		assert.Equal(t, "Foo", location.District())
		assert.Equal(t, "Washington", location.State())
		assert.Equal(t, "us", location.CountryCode())
		assert.Equal(t, "Seattle", location.City())
		assert.Equal(t, "places", location.Source())
	})
}
