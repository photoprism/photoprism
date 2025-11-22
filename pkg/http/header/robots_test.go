package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRobots(t *testing.T) {
	t.Run("Header", func(t *testing.T) {
		assert.Equal(t, "X-Robots-Tag", RobotsTag)
	})
	t.Run("Values", func(t *testing.T) {
		assert.Equal(t, "all", RobotsAll)
		assert.Equal(t, "noindex, nofollow", RobotsNone)
		assert.Equal(t, "noindex", RobotsNoIndex)
		assert.Equal(t, "nofollow", RobotsNoFollow)
		assert.Equal(t, "noimageindex", RobotsNoImages)
	})
}
