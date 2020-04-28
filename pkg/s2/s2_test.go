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

func TestLatLng(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca54c8b9")
		assert.Equal(t, 48.56344835921243, lat)
		assert.Equal(t, 8.996878323369781, lng)
	})

	t.Run("invalid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca5q")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lng)
	})
	t.Run("empty", func(t *testing.T) {
		lat, lng := LatLng("")
		assert.Equal(t, 0.0, lat)
		assert.Equal(t, 0.0, lng)
	})
}

func TestIsZero(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca54c8b9")
		assert.False(t, IsZero(lat, lng))
	})
	t.Run("invalid", func(t *testing.T) {
		lat, lng := LatLng("4799e370ca5q")
		assert.True(t, IsZero(lat, lng))
	})
}

func TestRange(t *testing.T) {
	t.Run("valid_1", func(t *testing.T) {
		min, max := Range("4799e370ca54c8b9", 1)
		assert.Equal(t, "4799e370ca54c8b1", min)
		assert.Equal(t, "4799e370ca54c8c1", max)
	})
	t.Run("valid_2", func(t *testing.T) {
		min, max := Range("4799e370ca54c8b9", 2)
		assert.Equal(t, "4799e370ca54c881", min)
		assert.Equal(t, "4799e370ca54c8c1", max)
	})
	t.Run("valid_3", func(t *testing.T) {
		min, max := Range("4799e370ca54c8b9", 3)
		assert.Equal(t, "4799e370ca54c801", min)
		assert.Equal(t, "4799e370ca54c901", max)
	})
	t.Run("valid_4", func(t *testing.T) {
		min, max := Range("4799e370ca54c8b9", 4)
		assert.Equal(t, "4799e370ca54c601", min)
		assert.Equal(t, "4799e370ca54ca01", max)
	})
	t.Run("valid_5", func(t *testing.T) {
		min, max := Range("4799e370ca54c8b9", 5)
		assert.Equal(t, "4799e370ca54c001", min)
		assert.Equal(t, "4799e370ca54d001", max)
	})
	t.Run("invalid", func(t *testing.T) {
		min, max := Range("4799e370ca5q", 1)
		assert.Equal(t, "", min)
		assert.Equal(t, "", max)
	})
}
