package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountryCode(t *testing.T) {
	t.Run("London", func(t *testing.T) {
		result := CountryCode("London")
		assert.Equal(t, "gb", result)
	})

	t.Run("reunion island", func(t *testing.T) {
		result := CountryCode("Reunion-Island-2019")
		assert.Equal(t, "zz", result)
	})

	t.Run("reunion island france", func(t *testing.T) {
		result := CountryCode("Reunion-Island-france-2019")
		assert.Equal(t, "fr", result)
	})

	t.Run("réunion", func(t *testing.T) {
		result := CountryCode("My-RéunioN-2019")
		assert.Equal(t, "fr", result)
	})

	t.Run("NYC", func(t *testing.T) {
		result := CountryCode("NYC 2019")
		assert.Equal(t, "us", result)
	})

	t.Run("Scuba", func(t *testing.T) {
		result := CountryCode("Scuba 2019")
		assert.Equal(t, "zz", result)
	})

	t.Run("Cuba", func(t *testing.T) {
		result := CountryCode("Cuba 2019")
		assert.Equal(t, "cu", result)
	})

	t.Run("San Francisco", func(t *testing.T) {
		result := CountryCode("San Francisco 2019")
		assert.Equal(t, "us", result)
	})

	t.Run("Los Angeles", func(t *testing.T) {
		result := CountryCode("I was in Los Angeles")
		assert.Equal(t, "us", result)
	})

	t.Run("St Gallen", func(t *testing.T) {
		result := CountryCode("St.----Gallen")
		assert.Equal(t, "ch", result)
	})

	t.Run("Congo Brazzaville", func(t *testing.T) {
		result := CountryCode("Congo Brazzaville")
		assert.Equal(t, "cg", result)
	})

	t.Run("Congo", func(t *testing.T) {
		result := CountryCode("Congo")
		assert.Equal(t, "cd", result)
	})

	t.Run("U.S.A.", func(t *testing.T) {
		result := CountryCode("Born in the U.S.A. is a song written and performed by Bruce Springsteen...")
		assert.Equal(t, "zz", result)
	})

	t.Run("US", func(t *testing.T) {
		result := CountryCode("Somebody help us please!")
		assert.Equal(t, "zz", result)
	})

	t.Run("Never mind Nirvana", func(t *testing.T) {
		result := CountryCode("Never mind Nirvana.")
		assert.Equal(t, "zz", result)
	})

	t.Run("empty string", func(t *testing.T) {
		result := CountryCode("")
		assert.Equal(t, "zz", result)
	})

	t.Run("zz", func(t *testing.T) {
		result := CountryCode("zz")
		assert.Equal(t, "zz", result)
	})

	t.Run("full path", func(t *testing.T) {
		result := CountryCode("2018/Oktober 2018/1.-7. Oktober 2018 Berlin/_MG_9831-112.jpg")
		assert.Equal(t, "de", result)
	})

	t.Run("little italy montreal", func(t *testing.T) {
		result := CountryCode("Little Italy Montreal")
		assert.Equal(t, "ca", result)
	})

	t.Run("little montreal italy", func(t *testing.T) {
		result := CountryCode("Little Montreal Italy")
		assert.Equal(t, "it", result)
	})
}
