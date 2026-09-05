package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
)

// TestNsfwPrivateFlag verifies that unavailable decisions preserve the private flag.
func TestNsfwPrivateFlag(t *testing.T) {
	t.Run("UnsafeFlagsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.NewResult(0.99, nsfw.DefaultThreshold))
		assert.True(t, flag)
		assert.True(t, write)
	})
	t.Run("UnsafeKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.NewResult(0.99, nsfw.DefaultThreshold))
		assert.True(t, flag)
		assert.False(t, write, "nothing to write when the flag already matches")
	})
	t.Run("SafeClearsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.NewResult(0.01, nsfw.DefaultThreshold))
		assert.False(t, flag)
		assert.True(t, write, "a decided safe result must still be able to clear the flag")
	})
	t.Run("SafeKeepsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.NewResult(0.01, nsfw.DefaultThreshold))
		assert.False(t, flag)
		assert.False(t, write)
	})
	t.Run("UnavailableKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.Unavailable("model is missing"))
		assert.True(t, flag, "an undecided result must not un-private a photo")
		assert.False(t, write)
	})
	t.Run("UnavailableKeepsPublicPhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(false, nsfw.Unavailable("thumbnail is missing"))
		assert.False(t, flag)
		assert.False(t, write)
	})
	// The zero value reaches this function whenever a caller forgets to decide, so it has to
	// behave like any other undecided result rather than like a clearance.
	t.Run("ZeroResultKeepsPrivatePhoto", func(t *testing.T) {
		flag, write := nsfwPrivateFlag(true, nsfw.Result{})
		assert.True(t, flag)
		assert.False(t, write)
	})
}
