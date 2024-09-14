package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumeric(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Numeric("")
		assert.Equal(t, "", result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Numeric("   Screenshot  ")
		assert.Equal(t, "", result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Numeric("0")
		assert.Equal(t, "0", result)
	})

	t.Run("0.5", func(t *testing.T) {
		result := Numeric("0.5")
		assert.Equal(t, "0.5", result)
	})

	t.Run("01:00", func(t *testing.T) {
		result := Numeric("01:00")
		assert.Equal(t, "0100", result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Numeric(" 000123")
		assert.Equal(t, "000123", result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Numeric("   123,556\t  ")
		assert.Equal(t, "123.556", result)
	})

	t.Run("PositiveFloat", func(t *testing.T) {
		result := Numeric("123,000.45245 ")
		assert.Equal(t, "123000.45245", result)
	})

	t.Run("NegativeFloat", func(t *testing.T) {
		result := Numeric(" - 123,000.45245 ")
		assert.Equal(t, "-123000.45245", result)
	})

	t.Run("MultipleDots", func(t *testing.T) {
		result := Numeric("123.000.45245.44 m")
		assert.Equal(t, "1230004524544", result)
	})
}
