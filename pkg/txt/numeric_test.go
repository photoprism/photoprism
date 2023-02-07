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

func TestInt64(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Int64("")
		assert.Equal(t, int64(0), result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Int64("   Screenshot  ")
		assert.Equal(t, int64(0), result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Int64("0")
		assert.Equal(t, int64(0), result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Int64(" 000123")
		assert.Equal(t, int64(123), result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Int64("   123,556\t  ")
		assert.Equal(t, int64(123), result)
	})

	t.Run("PositiveFloat", func(t *testing.T) {
		result := Int64("123,000.45245 ")
		assert.Equal(t, int64(123000), result)
	})

	t.Run("NegativeFloat", func(t *testing.T) {
		result := Int64(" - 123,000.45245 ")
		assert.Equal(t, int64(-123000), result)
	})

	t.Run("MultipleDots", func(t *testing.T) {
		result := Int64("123.000.45245.44 m")
		assert.Equal(t, int64(1230004524544), result)
	})
}

func TestFloat64(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Float64("")
		assert.Equal(t, 0.0, result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Float64("   Screenshot  ")
		assert.Equal(t, 0.0, result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Float64("0")
		assert.Equal(t, 0.0, result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Float64(" 000123")
		assert.Equal(t, 123.0, result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Float64("   123,556\t  ")
		assert.Equal(t, 123.556, result)
	})

	t.Run("PositiveFloat", func(t *testing.T) {
		result := Float64("123,000.45245 ")
		assert.Equal(t, 123000.45245, result)
	})

	t.Run("NegativeFloat", func(t *testing.T) {
		result := Float64(" - 123,000.45245 ")
		assert.Equal(t, -123000.45245, result)
	})

	t.Run("MultipleDots", func(t *testing.T) {
		result := Float64("123.000.45245.44 m")
		assert.Equal(t, 1230004524544.0, result)
	})
}
