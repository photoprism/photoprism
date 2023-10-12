package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := Int("")
		assert.Equal(t, 0, result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := Int("Screenshot")
		assert.Equal(t, 0, result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := Int("0")
		assert.Equal(t, 0, result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := Int("000123")
		assert.Equal(t, 123, result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := Int("   123\t  ")
		assert.Equal(t, 123, result)
	})

	t.Run("PositiveInt", func(t *testing.T) {
		result := Int("123")
		assert.Equal(t, 123, result)
	})

	t.Run("NegativeInt", func(t *testing.T) {
		result := Int("-123")
		assert.Equal(t, -123, result)
	})
}

func TestIntVal(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := IntVal("", 1, 31, 1)
		assert.Equal(t, 1, result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := IntVal("Screenshot", 1, 31, -1)
		assert.Equal(t, -1, result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := IntVal("0", -10, 10, -1)
		assert.Equal(t, 0, result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := IntVal("000123", 1, 1000, 1)
		assert.Equal(t, 123, result)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		result := IntVal("   123\t  ", 1, 1000, 1)
		assert.Equal(t, 123, result)
	})

	t.Run("PositiveInt", func(t *testing.T) {
		result := IntVal("123", 1, 1000, 1)
		assert.Equal(t, 123, result)
	})

	t.Run("NegativeInt", func(t *testing.T) {
		result := IntVal("-123", -1000, 1000, 1)
		assert.Equal(t, -123, result)
	})
}

func TestIntRange(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		start, end, err := IntRange("", 1, 31)
		assert.Equal(t, 0, start)
		assert.Equal(t, 0, end)
		assert.Error(t, err)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		start, end, err := IntRange("Screenshot", 1, 31)
		assert.Equal(t, 0, start)
		assert.Equal(t, 0, end)
		assert.Error(t, err)
	})

	t.Run("Day", func(t *testing.T) {
		start, end, err := IntRange("5-24", 1, 31)
		assert.Equal(t, 5, start)
		assert.Equal(t, 24, end)
		assert.NoError(t, err)
	})

	t.Run("Zero", func(t *testing.T) {
		start, end, err := IntRange("0", 5, 10)
		assert.Equal(t, 5, start)
		assert.Equal(t, 5, end)
		assert.NoError(t, err)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		start, end, err := IntRange("000123", 1, 1000)
		assert.Equal(t, 123, start)
		assert.Equal(t, 123, end)
		assert.NoError(t, err)
	})

	t.Run("WhitespacePadding", func(t *testing.T) {
		start, end, err := IntRange("   123\t  ", 1, 1000)
		assert.Equal(t, 123, start)
		assert.Equal(t, 123, end)
		assert.NoError(t, err)
	})

	t.Run("PositiveInt", func(t *testing.T) {
		start, end, err := IntRange("123", 1, 1000)
		assert.Equal(t, 123, start)
		assert.Equal(t, 123, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeInt", func(t *testing.T) {
		start, end, err := IntRange("-123", -1000, 1000)
		assert.Equal(t, -123, start)
		assert.Equal(t, -123, end)
		assert.NoError(t, err)
	})

	t.Run("ZeroOne", func(t *testing.T) {
		start, end, err := IntRange("0-1", -10, 10)
		assert.Equal(t, 0, start)
		assert.Equal(t, 1, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeRange", func(t *testing.T) {
		start, end, err := IntRange("-100--50", -100, 1000)
		assert.Equal(t, -100, start)
		assert.Equal(t, -50, end)
		assert.NoError(t, err)
	})

	t.Run("PositiveRange", func(t *testing.T) {
		start, end, err := IntRange("100 - 201", 1, 1000)
		assert.Equal(t, 100, start)
		assert.Equal(t, 201, end)
		assert.NoError(t, err)
	})

	t.Run("NegativeToPositive", func(t *testing.T) {
		start, end, err := IntRange("-99999-123456563", -1000000, 1000000)
		assert.Equal(t, -99999, start)
		assert.Equal(t, 1000000, end)
		assert.NoError(t, err)
	})
}

func TestUInt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := UInt("")
		assert.Equal(t, uint(0), result)
	})

	t.Run("NonNumeric", func(t *testing.T) {
		result := UInt("Screenshot")
		assert.Equal(t, uint(0), result)
	})

	t.Run("Zero", func(t *testing.T) {
		result := UInt("0")
		assert.Equal(t, uint(0), result)
	})

	t.Run("LeadingZeros", func(t *testing.T) {
		result := UInt("000123")
		assert.Equal(t, uint(0x7b), result)
	})

	t.Run("PositiveInt", func(t *testing.T) {
		result := UInt("123")
		assert.Equal(t, uint(0x7b), result)
	})

	t.Run("NegativeInt", func(t *testing.T) {
		result := UInt("-123")
		assert.Equal(t, uint(0), result)
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

func TestIsUInt(t *testing.T) {
	assert.False(t, IsUInt(""))
	assert.False(t, IsUInt("12 3"))
	assert.True(t, IsUInt("123"))
}

func TestIsPosInt(t *testing.T) {
	assert.False(t, IsPosInt(""))
	assert.False(t, IsPosInt("12 3"))
	assert.True(t, IsPosInt("123"))
	assert.False(t, IsPosInt(" "))
	assert.False(t, IsPosInt("-1"))
	assert.False(t, IsPosInt("0"))
	assert.False(t, IsPosInt("0.1"))
	assert.False(t, IsPosInt("0,1"))
	assert.True(t, IsPosInt("1"))
	assert.True(t, IsPosInt("99943546356"))
}
