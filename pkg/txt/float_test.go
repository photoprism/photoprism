package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestFloatRange(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		start, end, err := FloatRange("", 1, 31)
		assert.Equal(t, 0.0, start)
		assert.Equal(t, 0.0, end)
		assert.Error(t, err)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		start, end, err := FloatRange("Screenshot", 1, 31)
		assert.Equal(t, 0.0, start)
		assert.Equal(t, 0.0, end)
		assert.Error(t, err)
	})

	t.Run("Day", func(t *testing.T) {
		start, end, err := FloatRange("5.11-24.64", 1, 31)
		assert.Equal(t, 5.11, start)
		assert.Equal(t, 24.64, end)
		assert.NoError(t, err)
	})

	t.Run("Zero", func(t *testing.T) {
		start, end, err := FloatRange("0", 5, 10)
		assert.Equal(t, 5.0, start)
		assert.Equal(t, 5.0, end)
		assert.NoError(t, err)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		start, end, err := FloatRange("000123", 1, 1000)
		assert.Equal(t, 123.0, start)
		assert.Equal(t, 123.0, end)
		assert.NoError(t, err)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		start, end, err := FloatRange("   123\t  ", 1, 1000)
		assert.Equal(t, 123.0, start)
		assert.Equal(t, 123.0, end)
		assert.NoError(t, err)
	})

	t.Run("PositiveInt", func(t *testing.T) {
		start, end, err := FloatRange("123", 1, 1000)
		assert.Equal(t, 123.0, start)
		assert.Equal(t, 123.0, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeInt", func(t *testing.T) {
		start, end, err := FloatRange("-123", -1000, 1000)
		assert.Equal(t, -123.0, start)
		assert.Equal(t, -123.0, end)
		assert.NoError(t, err)
	})

	t.Run("ZeroOne", func(t *testing.T) {
		start, end, err := FloatRange("0-1", -10, 10)
		assert.Equal(t, 0.0, start)
		assert.Equal(t, 1.0, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeRange", func(t *testing.T) {
		start, end, err := FloatRange("-99.9--50.005", -100, 1000)
		assert.Equal(t, -99.9, start)
		assert.Equal(t, -50.005, end)
		assert.NoError(t, err)
	})

	t.Run("PositiveRange", func(t *testing.T) {
		start, end, err := FloatRange("100 - 201", 1, 1000)
		assert.Equal(t, 100.0, start)
		assert.Equal(t, 201.0, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeToPositive", func(t *testing.T) {
		start, end, err := FloatRange("-99999-123456563", -1000000, 1000000)
		assert.Equal(t, -99999.0, start)
		assert.Equal(t, 1000000.0, end)
		assert.NoError(t, err)
	})
}
