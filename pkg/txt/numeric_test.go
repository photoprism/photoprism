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

func TestIsFloat(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsFloat(""))
	})

	t.Run("Zero", func(t *testing.T) {
		assert.True(t, IsFloat("0"))
	})

	t.Run("0.5", func(t *testing.T) {
		assert.True(t, IsFloat("0.5"))
	})

	t.Run("0,5", func(t *testing.T) {
		assert.True(t, IsFloat("0,5"))
	})

	t.Run("123000.45245", func(t *testing.T) {
		assert.True(t, IsFloat("123000.45245 "))
	})

	t.Run("123000.", func(t *testing.T) {
		assert.True(t, IsFloat("123000. "))
	})

	t.Run("01:00", func(t *testing.T) {
		assert.False(t, IsFloat("01:00"))
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		assert.True(t, IsFloat(" 000123"))
	})

	t.Run("Comma", func(t *testing.T) {
		assert.True(t, IsFloat("   123,556\t  "))
	})
}

func TestFloat(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Float("")
		assert.Equal(t, 0.0, result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Float("   Screenshot  ")
		assert.Equal(t, 0.0, result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Float("0")
		assert.Equal(t, 0.0, result)
	})

	t.Run("0.5", func(t *testing.T) {
		result := Float("0.5")
		assert.Equal(t, 0.5, result)
	})

	t.Run("01:00", func(t *testing.T) {
		result := Float("01:00")
		assert.Equal(t, 100.0, result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Float(" 000123")
		assert.Equal(t, 123.0, result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Float("   123,556\t  ")
		assert.Equal(t, 123.556, result)
	})

	t.Run("PositiveFloat", func(t *testing.T) {
		result := Float("123,000.45245 ")
		assert.Equal(t, 123000.45245, result)
	})

	t.Run("NegativeFloat", func(t *testing.T) {
		result := Float(" - 123,000.45245 ")
		assert.Equal(t, -123000.45245, result)
	})

	t.Run("MultipleDots", func(t *testing.T) {
		result := Float("123.000.45245.44 m")
		assert.Equal(t, 1230004524544.0, result)
	})
}

func TestFloat32(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Float32("")
		assert.Equal(t, float32(0), result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Float32(" 000123")
		assert.Equal(t, float32(123), result)
	})

	t.Run("LongFloat", func(t *testing.T) {
		result := Float32("123.87945632786543786547")
		assert.Equal(t, float32(123.87945632786543786547), result)
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
