package s2

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Run("germany", func(t *testing.T) {
		token := Token(48.56344833333333, 8.996878333333333)
		expected := "4799e370"

		assert.True(t, strings.HasPrefix(token, expected))
	})

	t.Run("lat_overflow", func(t *testing.T) {
		token := Token(548.56344833333333, 8.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("lng_overflow", func(t *testing.T) {
		token := Token(48.56344833333333, 258.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})
}

func TestTokenLevel(t *testing.T) {
	t.Run("level_30", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 30)
		expected := "4799e370ca54c8b9"

		assert.Equal(t, expected, token)
	})

	t.Run("level_30_diff", func(t *testing.T) {
		plusCode := TokenLevel(48.56344839999999, 8.996878339999999, 30)
		expected := "4799e370ca54c8b7"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("level_21", func(t *testing.T) {
		plusCode := TokenLevel(48.56344839999999, 8.996878339999999, 21)
		expected := "4799e370ca54"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("level_18", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, token)
	})

	t.Run("level_18_diff", func(t *testing.T) {
		token := TokenLevel(48.56344839999999, 8.996878339999999, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, token)
	})

	t.Run("level_15", func(t *testing.T) {
		plusCode := TokenLevel(48.56344833333333, 8.996878333333333, 15)
		expected := "4799e370c"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("level_10", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 10)
		expected := "4799e3"

		assert.Equal(t, expected, token)
	})

	t.Run("lat_overflow", func(t *testing.T) {
		token := TokenLevel(548.56344833333333, 8.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("lng_overflow", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 258.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("lat & long 0.0", func(t *testing.T) {
		token := TokenLevel(0.0, 0.0, 30)
		expected := ""

		assert.Equal(t, expected, token)
	})
}

func TestLevel(t *testing.T) {
	t.Run("8000", func(t *testing.T) {
		assert.Equal(t, 0, Level(8000))
	})
	t.Run("150", func(t *testing.T) {
		assert.Equal(t, 6, Level(150))
	})
	t.Run("0.25", func(t *testing.T) {
		assert.Equal(t, 15, Level(0.25))
	})
	t.Run("0.1", func(t *testing.T) {
		assert.Equal(t, 17, Level(0.1))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, 21, Level(0))
	})
	t.Run("3999", func(t *testing.T) {
		assert.Equal(t, 1, Level(3999))
	})
	t.Run("1825", func(t *testing.T) {
		assert.Equal(t, 2, Level(1825))
	})
	t.Run("1500", func(t *testing.T) {
		assert.Equal(t, 3, Level(1500))
	})
	t.Run("600", func(t *testing.T) {
		assert.Equal(t, 4, Level(600))
	})
	t.Run("100", func(t *testing.T) {
		assert.Equal(t, 7, Level(100))
	})
	t.Run("40", func(t *testing.T) {
		assert.Equal(t, 8, Level(40))
	})
	t.Run("25", func(t *testing.T) {
		assert.Equal(t, 9, Level(25))
	})
	t.Run("10", func(t *testing.T) {
		assert.Equal(t, 10, Level(10))
	})
	t.Run("5", func(t *testing.T) {
		assert.Equal(t, 11, Level(5))
	})
	t.Run("3", func(t *testing.T) {
		assert.Equal(t, 12, Level(3))
	})
	t.Run("1.5", func(t *testing.T) {
		assert.Equal(t, 13, Level(1.5))
	})
	t.Run("0.5", func(t *testing.T) {
		assert.Equal(t, 14, Level(0.5))
	})
	t.Run("0.15", func(t *testing.T) {
		assert.Equal(t, 16, Level(0.15))
	})
	t.Run("0.03", func(t *testing.T) {
		assert.Equal(t, 18, Level(0.03))
	})
	t.Run("0.015", func(t *testing.T) {
		assert.Equal(t, 19, Level(0.015))
	})
	t.Run("0.008", func(t *testing.T) {
		assert.Equal(t, 20, Level(0.008))
	})
	t.Run("445", func(t *testing.T) {
		assert.Equal(t, 5, Level(445))
	})
	t.Run("Negative", func(t *testing.T) {
		assert.Equal(t, 21, Level(-1))
	})
}

func TestLatLng(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca54c8b9")
		assert.Equal(t, 48.56344835921243, lat)
		assert.Equal(t, 8.996878323369781, lng)
	})

	t.Run("Invalid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca5q")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lng)
	})
	t.Run("Empty", func(t *testing.T) {
		lat, lng := LatLng("")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lng)
	})
}

func TestIsZero(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca54c8b9")
		assert.False(t, IsZero(lat, lng))
	})
	t.Run("Invalid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca5q")
		assert.True(t, IsZero(lat, lng))
	})
}

func TestRange(t *testing.T) {
	t.Run("Level1", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 1)
		assert.Equal(t, "3800000000000001", start)
		assert.Equal(t, "4800000000000001", end)
	})
	t.Run("Level2", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 2)
		assert.Equal(t, "4400000000000001", start)
		assert.Equal(t, "4800000000000001", end)
	})
	t.Run("Level5", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 5)
		assert.Equal(t, "4790000000000001", start)
		assert.Equal(t, "47a0000000000001", end)
	})
	t.Run("Level7", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 7)
		assert.Equal(t, "4799000000000001", start)
		assert.Equal(t, "479a000000000001", end)
	})
	t.Run("Level10", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 10)
		assert.Equal(t, "4799e00000000001", start)
		assert.Equal(t, "4799e40000000001", end)
	})
	t.Run("Level14", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 14)
		assert.Equal(t, "4799e36e00000001", start)
		assert.Equal(t, "4799e37200000001", end)
	})
	t.Run("Level21", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 21)
		assert.Equal(t, "4799e370ca480001", start)
		assert.Equal(t, "4799e370ca580001", end)
	})
	t.Run("Level23", func(t *testing.T) {
		start, end := Range("4799e370ca54c8b9", 23)
		assert.Equal(t, "4799e370ca540001", start)
		assert.Equal(t, "4799e370ca550001", end)
	})
	t.Run("Invalid", func(t *testing.T) {
		start, end := Range("4799e370ca5q", 1)
		assert.Equal(t, "", start)
		assert.Equal(t, "", end)
	})
}
