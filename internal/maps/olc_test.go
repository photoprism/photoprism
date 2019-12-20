package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOlcEncode(t *testing.T) {
	t.Run("Wildgehege", func(t *testing.T) {
		plusCode := OlcEncode(48.56344833333333, 8.996878333333333)
		expected := "8FWCHX7W+"

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LatOverflow", func(t *testing.T) {
		plusCode := OlcEncode(548.56344833333333, 8.996878333333333)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})

	t.Run("LongOverflow", func(t *testing.T) {
		plusCode := OlcEncode(48.56344833333333, 258.996878333333333)
		expected := ""

		assert.Equal(t, expected, plusCode)
	})
}
