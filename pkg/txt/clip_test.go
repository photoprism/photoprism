package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClip(t *testing.T) {
	t.Run("clip", func(t *testing.T) {
		assert.Equal(t, "I'm ä", Clip("I'm ä lazy BRoWN fox!", 6))
	})
	t.Run("ok", func(t *testing.T) {
		assert.Equal(t, "I'm ä lazy BRoWN fox!", Clip("I'm ä lazy BRoWN fox!", 128))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", Clip("", -1))
	})
}
