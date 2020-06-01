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
		m := Photo{PhotoName: "flughafen", PhotoPath: "2020/Berlin", OriginalName: "Flughafen BER.jpg"}
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

	//OriginalName: "demo/Toronto/port-lands--gardiner-expressway--toronto_7999515645_o.jpg"
}
