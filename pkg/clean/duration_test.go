package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuration(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Duration("")
		assert.Equal(t, "", result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Duration("   Screenshot  ")
		assert.Equal(t, "", result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Duration("0")
		assert.Equal(t, "0", result)
	})

	t.Run("Float", func(t *testing.T) {
		result := Duration("0.5")
		assert.Equal(t, "0.5", result)
	})

	t.Run("Seconds", func(t *testing.T) {
		result := Duration("0.5 s")
		assert.Equal(t, "0.5s", result)
	})

	t.Run("MinutesSeconds", func(t *testing.T) {
		result := Duration("1.0 m0.01 s ")
		assert.Equal(t, "1.0m0.01s", result)
	})

	t.Run("01:00", func(t *testing.T) {
		result := Duration("01:00")
		assert.Equal(t, "01:00", result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Duration(" 000123")
		assert.Equal(t, "000123", result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Duration("   123,556\t  ")
		assert.Equal(t, "123.556", result)
	})

	t.Run("PositiveFloat", func(t *testing.T) {
		result := Duration("123,000.45245 ")
		assert.Equal(t, "123.00045245", result)
	})

	t.Run("NegativeFloat", func(t *testing.T) {
		result := Duration(" - 123,000.45245 ")
		assert.Equal(t, "-123.00045245", result)
	})

	t.Run("MultipleDots", func(t *testing.T) {
		result := Duration("123.000.45245.44 m")
		assert.Equal(t, "123.0004524544m", result)
	})
}
