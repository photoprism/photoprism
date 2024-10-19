package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCountry(t *testing.T) {
	t.Run("unknown country", func(t *testing.T) {
		country := NewCountry("", "")

		assert.Equal(t, &UnknownCountry, country)
	})
	t.Run("United States", func(t *testing.T) {
		country := NewCountry("us", "United States")

		assert.Equal(t, "United States", country.CountryName)
		assert.Equal(t, "united-states", country.CountrySlug)
	})
	t.Run("Germany", func(t *testing.T) {
		country := NewCountry("de", "Germany")

		assert.Equal(t, "Germany", country.CountryName)
		assert.Equal(t, "germany", country.CountrySlug)
	})
}

func TestFirstOrCreateCountry(t *testing.T) {
	t.Run("es", func(t *testing.T) {
		country := NewCountry("es", "spain")
		country = FirstOrCreateCountry(country)
		if country == nil {
			t.Fatal("country should not be nil")
		}
	})
	t.Run("de", func(t *testing.T) {
		country := &Country{ID: "de"}
		r := FirstOrCreateCountry(country)
		if r == nil {
			t.Fatal("country should not be nil")
		}
	})
}

func TestCountry_Name(t *testing.T) {
	country := NewCountry("xy", "Neverland")
	assert.Equal(t, "Neverland", country.Name())
}

func TestCountry_Code(t *testing.T) {
	country := NewCountry("xy", "Neverland")
	assert.Equal(t, "xy", country.Code())
}
