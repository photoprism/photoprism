package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_EstimateCountry(t *testing.T) {
	t.Run("uk", func(t *testing.T) {
		m := Photo{PhotoName: "20200102_194030_9EFA9E5E", PhotoPath: "2000/05", OriginalName: "flickr import/changing-of-the-guard--buckingham-palace_7925318070_o.jpg"}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "gb", m.CountryCode())
		assert.Equal(t, "United Kingdom", m.CountryName())
	})

	t.Run("zz", func(t *testing.T) {
		m := Photo{PhotoName: "20200102_194030_ADADADAD", PhotoPath: "2020/Berlin", OriginalName: "Zimmermannstrasse.jpg"}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
	})

	t.Run("de", func(t *testing.T) {
		m := Photo{PhotoName: "Brauhaus", PhotoPath: "2020/Bayern", OriginalName: "München.jpg"}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "de", m.CountryCode())
		assert.Equal(t, "Germany", m.CountryName())
	})

	t.Run("ca", func(t *testing.T) {
		m := Photo{PhotoTitle: "Port Lands / Gardiner Expressway / Toronto", PhotoPath: "2012/09", PhotoName: "20120910_231851_CA06E1AD", OriginalName: "demo/Toronto/port-lands--gardiner-expressway--toronto_7999515645_o.jpg"}
		assert.Equal(t, UnknownCountry.ID, m.CountryCode())
		assert.Equal(t, UnknownCountry.CountryName, m.CountryName())
		m.EstimateCountry()
		assert.Equal(t, "ca", m.CountryCode())
		assert.Equal(t, "Canada", m.CountryName())
	})
	t.Run("photo has latlng", func(t *testing.T) {
		m := Photo{PhotoTitle: "Port Lands / Gardiner Expressway / Toronto", PhotoLat: 13.333, PhotoLng: 40.998, PhotoName: "20120910_231851_CA06E1AD", OriginalName: "demo/Toronto/port-lands--gardiner-expressway--toronto_7999515645_o.jpg"}
		m.EstimateCountry()
		assert.Equal(t, "zz", m.CountryCode())
		assert.Equal(t, "Unknown", m.CountryName())
	})
}

func TestPhoto_Optimize(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")
		if updated, err := photo.Optimize(); err != nil {
			t.Fatal(err)
		} else if !updated {
			t.Error("photo should be updated")
		}

		if updated, err := photo.Optimize(); err != nil {
			t.Fatal(err)
		} else if updated {
			t.Error("photo should NOT be updated")
		}
	})
	t.Run("photo withouth id", func(t *testing.T) {
		photo := Photo{}
		bool, err := photo.Optimize()
		assert.Error(t, err)
		assert.False(t, bool)

	})
}
