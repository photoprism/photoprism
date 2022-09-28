package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClip(t *testing.T) {
	t.Run("ASCII", func(t *testing.T) {
		assert.Equal(t, "ASCI", Clip("ASCII", 4))
	})
	t.Run("ShortEnough", func(t *testing.T) {
		assert.Equal(t, "I'm ä lazy BRoWN fox!", Clip("I'm ä lazy BRoWN fox!", 128))
	})
	t.Run("Clip", func(t *testing.T) {
		assert.Equal(t, "I'm ä", Clip("I'm ä lazy BRoWN fox!", 6))
		assert.Equal(t, "I'm ä l", Clip("I'm ä lazy BRoWN fox!", 7))
	})
	t.Run("TrimSpace", func(t *testing.T) {
		assert.Equal(t, "abc", Clip(" abc ty3q5y4y46uy", 4))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Clip("", -1))
	})
}

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
