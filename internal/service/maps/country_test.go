package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountryName(t *testing.T) {
	t.Run("gb", func(t *testing.T) {
		result := CountryName("gb")
		assert.Equal(t, "United Kingdom", result)
	})

	t.Run("us", func(t *testing.T) {
		result := CountryName("us")
		assert.Equal(t, "United States", result)
	})

	t.Run("empty", func(t *testing.T) {
		result := CountryName("")
		assert.Equal(t, "Unknown", result)
	})
	t.Run("invalid", func(t *testing.T) {
		result := CountryName("xyz")
		assert.Equal(t, "Unknown", result)
	})
}
