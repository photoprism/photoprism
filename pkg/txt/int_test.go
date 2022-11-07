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
