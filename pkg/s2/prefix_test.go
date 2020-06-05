package s2

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeToken(t *testing.T) {
	t.Run(TokenPrefix+"1242342bac", func(t *testing.T) {
		input := TokenPrefix + "1242342bac"

		output := NormalizeToken(input)

		assert.Equal(t, "1242342bac", output)

	})

	t.Run("abc", func(t *testing.T) {
		input := "abc"

		output := NormalizeToken(input)

		assert.Equal(t, "abc", output)

	})
}

func TestPrefix(t *testing.T) {
	t.Run(TokenPrefix+"1242342bac", func(t *testing.T) {
		input := TokenPrefix + "1242342bac"

		output := Prefix(input)

		assert.Equal(t, input, output)

	})

	t.Run("abc", func(t *testing.T) {
		input := "1242342bac"

		output := Prefix(input)

		assert.Equal(t, TokenPrefix+input, output)

	})

	t.Run("empty string", func(t *testing.T) {
		output := Prefix("")

		assert.Equal(t, "", output)

	})
}

func TestPrefixedToken(t *testing.T) {
	t.Run("germany", func(t *testing.T) {
		token := PrefixedToken(48.56344833333333, 8.996878333333333)
		expected := TokenPrefix + "4799e370"

		assert.True(t, strings.HasPrefix(token, expected))
	})

	t.Run("lat_overflow", func(t *testing.T) {
		token := PrefixedToken(548.56344833333333, 8.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("lng_overflow", func(t *testing.T) {
		token := PrefixedToken(48.56344833333333, 258.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})
}

func TestPrefixedRange(t *testing.T) {
	t.Run("valid_1", func(t *testing.T) {
		min, max := PrefixedRange("4799e370ca54c8b9", 1)
		assert.Equal(t, TokenPrefix+"4799e370ca54c8b1", min)
		assert.Equal(t, TokenPrefix+"4799e370ca54c8c1", max)
	})
	t.Run("valid_2", func(t *testing.T) {
		min, max := PrefixedRange(TokenPrefix+"4799e370ca54c8b9", 2)
		assert.Equal(t, TokenPrefix+"4799e370ca54c881", min)
		assert.Equal(t, TokenPrefix+"4799e370ca54c8c1", max)
	})
	t.Run("valid_3", func(t *testing.T) {
		min, max := PrefixedRange("4799e370ca54c8b9", 3)
		assert.Equal(t, TokenPrefix+"4799e370ca54c801", min)
		assert.Equal(t, TokenPrefix+"4799e370ca54c901", max)
	})
	t.Run("valid_4", func(t *testing.T) {
		min, max := PrefixedRange(TokenPrefix+"4799e370ca54c8b9", 4)
		assert.Equal(t, TokenPrefix+"4799e370ca54c601", min)
		assert.Equal(t, TokenPrefix+"4799e370ca54ca01", max)
	})
	t.Run("valid_5", func(t *testing.T) {
		min, max := PrefixedRange("4799e370ca54c8b9", 5)
		assert.Equal(t, TokenPrefix+"4799e370ca54c001", min)
		assert.Equal(t, TokenPrefix+"4799e370ca54d001", max)
	})
	t.Run("invalid", func(t *testing.T) {
		min, max := PrefixedRange("4799e370ca5q", 1)
		assert.Equal(t, "", min)
		assert.Equal(t, "", max)
	})
}
