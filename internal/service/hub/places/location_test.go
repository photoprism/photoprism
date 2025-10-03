package places

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocationGetters(t *testing.T) {
	location := Location{
		ID:          "1e95998417cc",
		LocLat:      52.51961810676184,
		LocLng:      13.40806264572578,
		LocName:     "TestLocation",
		LocStreet:   "",
		LocPostcode: "",
		LocCategory: "test",
		Place:       Place{PlaceID: "1", LocLabel: "testLabel", LocDistrict: "Berlin", LocCity: "", LocState: "Berlin", LocCountry: "de", LocKeywords: "foobar"},
		Cached:      true,
	}
	t.Run("WrongId", func(t *testing.T) {
		assert.Equal(t, "1e95998417cc", location.CellID())
		assert.Equal(t, "TestLocation", location.Name())
		assert.Equal(t, "test", location.Category())
		assert.Equal(t, "testLabel", location.Label())
		assert.Equal(t, "Berlin", location.State())
		assert.Equal(t, "de", location.CountryCode())
		assert.Equal(t, "Berlin", location.District())
		assert.Equal(t, "", location.City())
		assert.Equal(t, 52.51961810676184, location.Latitude())
		assert.Equal(t, 13.40806264572578, location.Longitude())
		assert.Equal(t, "places", location.Source())
		assert.Equal(t, []string{"foobar"}, location.Keywords())
	})
}

func TestLocation_State(t *testing.T) {
	location := Location{
		ID:          "54903ee07f74",
		LocLat:      47.6129432,
		LocLng:      -122.4821475,
		LocName:     "TestLocation",
		LocStreet:   "",
		LocPostcode: "",
		LocCategory: "test",
		Place:       Place{PlaceID: "549ed22c0434", LocLabel: "Seattle, WA", LocDistrict: "Berlin", LocCity: "Seattle", LocState: "WA", LocCountry: "us", LocKeywords: "foobar"},
		Cached:      true,
	}
	t.Run("Washington", func(t *testing.T) {
		assert.Equal(t, "54903ee07f74", location.CellID())
		assert.Equal(t, "Seattle, WA", location.Label())
		assert.Equal(t, "Washington", location.State())
		assert.Equal(t, "us", location.CountryCode())
		assert.Equal(t, "Seattle", location.City())
		assert.Equal(t, "places", location.Source())
	})
}

func TestLocation_District(t *testing.T) {
	location := Location{
		ID:          "54903ee07f74",
		LocLat:      47.6129432,
		LocLng:      -122.4821475,
		LocName:     "TestLocation",
		LocStreet:   "",
		LocPostcode: "",
		LocCategory: "test",
		Place:       Place{PlaceID: "549ed22c0434", LocLabel: "Seattle, WA", LocDistrict: "Foo", LocCity: "Seattle", LocState: "WA", LocCountry: "us", LocKeywords: "foobar"},
		Cached:      true,
	}
	t.Run("Washington", func(t *testing.T) {
		assert.Equal(t, "54903ee07f74", location.CellID())
		assert.Equal(t, "Seattle, WA", location.Label())
		assert.Equal(t, "Foo", location.District())
		assert.Equal(t, "Washington", location.State())
		assert.Equal(t, "us", location.CountryCode())
		assert.Equal(t, "Seattle", location.City())
		assert.Equal(t, "places", location.Source())
	})
}
