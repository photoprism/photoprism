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
	t.Run("Level1", func(t *testing.T) {
		start, end := PrefixedRange("4799e370ca54c8b9", 1)
		assert.Equal(t, TokenPrefix+"3800000000000001", start)
		assert.Equal(t, TokenPrefix+"4800000000000001", end)
	})
	t.Run("Level2", func(t *testing.T) {
		start, end := PrefixedRange(TokenPrefix+"4799e370ca54c8b9", 2)
		assert.Equal(t, TokenPrefix+"4400000000000001", start)
		assert.Equal(t, TokenPrefix+"4800000000000001", end)
	})
	t.Run("Level3", func(t *testing.T) {
		start, end := PrefixedRange("4799e370ca54c8b9", 3)
		assert.Equal(t, TokenPrefix+"4700000000000001", start)
		assert.Equal(t, TokenPrefix+"4800000000000001", end)
	})
	t.Run("Level4", func(t *testing.T) {
		start, end := PrefixedRange(TokenPrefix+"4799e370ca54c8b9", 4)
		assert.Equal(t, TokenPrefix+"4760000000000001", start)
		assert.Equal(t, TokenPrefix+"47a0000000000001", end)
	})
	t.Run("Level5", func(t *testing.T) {
		start, end := PrefixedRange("4799e370ca54c8b9", 5)
		assert.Equal(t, TokenPrefix+"4790000000000001", start)
		assert.Equal(t, TokenPrefix+"47a0000000000001", end)
	})
	t.Run("Invalid", func(t *testing.T) {
		start, end := PrefixedRange("4799e370ca5q", 1)
		assert.Equal(t, "", start)
		assert.Equal(t, "", end)
	})
}
