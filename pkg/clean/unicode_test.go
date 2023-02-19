package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnicode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		assert.Equal(t, "Naïve bonds and futures surge as inflation eases 🚀🚀🚀", Unicode("Naïve bonds and futures surge as inflation eases 🚀🚀🚀"))
	})
	t.Run("Null", func(t *testing.T) {
		assert.Equal(t, "\x00", Unicode("\x00"))
	})
	t.Run("FFFD", func(t *testing.T) {
		assert.Equal(t, "", Unicode("\uFFFD"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Unicode(""))
	})
}
