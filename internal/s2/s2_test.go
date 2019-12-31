package s2

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	t.Run("Wildgehege", func(t *testing.T) {
		token := Token(48.56344833333333, 8.996878333333333)
		expected := "4799e370"

		assert.True(t, strings.HasPrefix(token, expected))
	})

	t.Run("LatOverflow", func(t *testing.T) {
		token := Token(548.56344833333333, 8.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		token := Token(48.56344833333333, 258.996878333333333)
		expected := ""

		assert.Equal(t, expected, token)
	})
}

func TestTokenLevel(t *testing.T) {
	t.Run("Wildgehege30", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 30)
		expected := "4799e370ca54c8b9"

		assert.Equal(t, expected, token)
	})

	t.Run("NearWildgehege30", func(t *testing.T) {
		plusCode := TokenLevel(48.56344839999999, 8.996878339999999, 30)
		expected := "4799e370ca54c8b7"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege18", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, token)
	})

	t.Run("NearWildgehege18", func(t *testing.T) {
		token := TokenLevel(48.56344839999999, 8.996878339999999, 18)
		expected := "4799e370cb"

		assert.Equal(t, expected, token)
	})

	t.Run("NearWildgehege15", func(t *testing.T) {
		plusCode := TokenLevel(48.56344833333333, 8.996878333333333, 15)
		expected := "4799e370c"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("Wildgehege10", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 8.996878333333333, 10)
		expected := "4799e3"

		assert.Equal(t, expected, token)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		token := TokenLevel(548.56344833333333, 8.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, token)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		token := TokenLevel(48.56344833333333, 258.996878333333333, 30)
		expected := ""

		assert.Equal(t, expected, token)
	})
}
