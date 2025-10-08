package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountryName(t *testing.T) {
	t.Run("Gb", func(t *testing.T) {
		result := CountryName("gb")
		assert.Equal(t, "United Kingdom", result)
	})
	t.Run("Us", func(t *testing.T) {
		result := CountryName("us")
		assert.Equal(t, "United States", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := CountryName("")
		assert.Equal(t, "Unknown", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := CountryName("xyz")
		assert.Equal(t, "Unknown", result)
	})
}
