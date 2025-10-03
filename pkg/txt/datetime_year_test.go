package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYear(t *testing.T) {
	t.Run("LondonNum2002", func(t *testing.T) {
		result := Year("/2002/London 81/")
		assert.Equal(t, 2002, result)
	})
	t.Run("SanFranciscoNum2019", func(t *testing.T) {
		result := Year("San Francisco 2019")
		assert.Equal(t, 2019, result)
	})
	t.Run("StringWithNoNumber", func(t *testing.T) {
		result := Year("Born in the U.S.A. is a song written and performed by Bruce Springsteen...")
		assert.Equal(t, 0, result)
	})
	t.Run("FileName", func(t *testing.T) {
		result := Year("/share/photos/243546/2003/01/myfile.jpg")
		assert.Equal(t, 2003, result)
	})
	t.Run("Num1981", func(t *testing.T) {
		result := Year("/root/1981/London 2005")
		assert.Equal(t, 1981, result)
	})
	t.Run("Num1970", func(t *testing.T) {
		result := Year("/root/1970/London 2005")
		assert.Equal(t, 2005, result)
	})
	t.Run("Num1969", func(t *testing.T) {
		result := Year("/root/1969/London 2005")
		assert.Equal(t, 2005, result)
	})
	t.Run("Num1950", func(t *testing.T) {
		result := Year("/root/1950/London 2005")
		assert.Equal(t, 2005, result)
	})
	t.Run("EmptyString", func(t *testing.T) {
		result := Year("")
		assert.Equal(t, 0, result)
	})
}

func TestExpandYear(t *testing.T) {
	t.Run("Num1977", func(t *testing.T) {
		result := ExpandYear("1977")
		assert.Equal(t, 1977, result)
	})
	t.Run("Num2002", func(t *testing.T) {
		result := ExpandYear("2002")
		assert.Equal(t, 2002, result)
	})
	t.Run("Num2019", func(t *testing.T) {
		result := ExpandYear("2019")
		assert.Equal(t, 2019, result)
	})
	t.Run("XXXX", func(t *testing.T) {
		result := ExpandYear("XXXX")
		assert.Equal(t, -1, result)
	})
	t.Run("Num88", func(t *testing.T) {
		result := ExpandYear("88")
		assert.Equal(t, -1, result)
	})
	t.Run("Num91", func(t *testing.T) {
		result := ExpandYear("91")
		assert.Equal(t, 1991, result)
	})
	t.Run("Num01", func(t *testing.T) {
		result := ExpandYear("01")
		assert.Equal(t, 2001, result)
	})
	t.Run("One", func(t *testing.T) {
		result := ExpandYear("1")
		assert.Equal(t, -1, result)
	})
	t.Run("Twelve", func(t *testing.T) {
		result := ExpandYear("12")
		assert.Equal(t, 2012, result)
	})
	t.Run("Num22", func(t *testing.T) {
		result := ExpandYear("22")
		assert.Equal(t, 2022, result)
	})
}
