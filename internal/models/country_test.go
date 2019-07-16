package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCountry(t *testing.T) {
	t.Run("name Fantasy code fy", func(t *testing.T) {
		country := NewCountry("fy", "Fantasy")
		assert.Equal(t, "fy", country.ID)
		assert.Equal(t, "Fantasy", country.CountryName)
		assert.Equal(t, "fantasy", country.CountrySlug)
	})
	t.Run("name Unknown code Unknown", func(t *testing.T) {
		country := NewCountry("", "")
		assert.Equal(t, "zz", country.ID)
		assert.Equal(t, "Unknown", country.CountryName)
		assert.Equal(t, "unknown", country.CountrySlug)
	})
}
