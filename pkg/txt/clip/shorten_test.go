package clip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {
	t.Run("ShortEnough", func(t *testing.T) {
		assert.Equal(t, "fox!", Shorten("fox!", 6, "..."))
	})
	t.Run("CustomSuffix", func(t *testing.T) {
		assert.Equal(t, "I'm ä...", Shorten("I'm ä lazy BRoWN fox!", 8, "..."))
	})
	t.Run("DefaultSuffix", func(t *testing.T) {
		assert.Equal(t, "I'm…", Shorten("I'm ä lazy BRoWN fox!", 7, ""))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Shorten("", -1, ""))
	})
}
