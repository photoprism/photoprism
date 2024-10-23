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

	t.Run("ReunionIsland", func(t *testing.T) {
		result := CountryCode("Reunion-Island-2019")
		assert.Equal(t, "zz", result)
	})

	t.Run("ReunionIslandFrance", func(t *testing.T) {
		result := CountryCode("Reunion-Island-france-2019")
		assert.Equal(t, "fr", result)
	})

	t.Run("Reunion", func(t *testing.T) {
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

	t.Run("SanFrancisco", func(t *testing.T) {
		result := CountryCode("San Francisco 2019")
		assert.Equal(t, "us", result)
	})

	t.Run("LosAngeles", func(t *testing.T) {
		result := CountryCode("I was in Los Angeles")
		assert.Equal(t, "us", result)
	})

	t.Run("Melbourne", func(t *testing.T) {
		result := CountryCode("The name Narrm is commonly used by the broader Aboriginal community\n\rto refer to the city, \t stemming from the traditional name recorded for the area on which the Melbourne city centre is built.")
		assert.Equal(t, "au", result)
	})

	t.Run("ZugspitzeMelbourne", func(t *testing.T) {
		result := CountryCode("The name Narrm is commonly used by the broader Zugspitze community\n\rto refer to the city, \t stemming from the traditional name recorded for the area on which the Melbourne city centre is built.")
		assert.Equal(t, "au", result)
	})

	t.Run("MelbourneZugspitze", func(t *testing.T) {
		result := CountryCode("The name Narrm is commonly used by the broader Melbourne community\n\rto refer to the city, \t stemming from the traditional name recorded for the area on which the Zugspitze city centre is built.")
		assert.Equal(t, "de", result)
	})

	t.Run("StGallen", func(t *testing.T) {
		result := CountryCode("St.----Gallen")
		assert.Equal(t, "ch", result)
	})

	t.Run("CongoBrazzaville", func(t *testing.T) {
		result := CountryCode("Congo Brazzaville")
		assert.Equal(t, "cg", result)
	})

	t.Run("Congo", func(t *testing.T) {
		result := CountryCode("Congo")
		assert.Equal(t, "cd", result)
	})

	t.Run("BornInTheUSA", func(t *testing.T) {
		result := CountryCode("Born in the U.S.A. is a song written and performed by Bruce Springsteen...")
		assert.Equal(t, "zz", result)
	})

	t.Run("SomebodyHelpUs", func(t *testing.T) {
		result := CountryCode("Somebody help us please!")
		assert.Equal(t, "zz", result)
	})

	t.Run("NeverMindNirvana", func(t *testing.T) {
		result := CountryCode("Never mind Nirvana.")
		assert.Equal(t, "zz", result)
	})

	t.Run("EmptyString", func(t *testing.T) {
		result := CountryCode("")
		assert.Equal(t, "zz", result)
	})

	t.Run("Unknown", func(t *testing.T) {
		result := CountryCode("zz")
		assert.Equal(t, "zz", result)
	})

	t.Run("DirectoryName", func(t *testing.T) {
		result := CountryCode("2018/Oktober 2018/1.-7. Oktober 2018 Berlin/_MG_9831-112.jpg")
		assert.Equal(t, "de", result)
	})

	t.Run("LittleItaly", func(t *testing.T) {
		result := CountryCode("Little Italy Montreal")
		assert.Equal(t, "ca", result)
	})

	t.Run("LittleMontreal", func(t *testing.T) {
		result := CountryCode("Little Montreal Italy")
		assert.Equal(t, "it", result)
	})

	t.Run("Sharjah", func(t *testing.T) {
		result := CountryCode("Sharjah")
		assert.Equal(t, "ae", result)
	})

	t.Run("Arabic", func(t *testing.T) {
		result := CountryCode("الشارقة")
		assert.Equal(t, "ae", result)
	})

	t.Run("Hebrew", func(t *testing.T) {
		result := CountryCode("באר שבע")
		assert.Equal(t, "il", result)
	})
}
