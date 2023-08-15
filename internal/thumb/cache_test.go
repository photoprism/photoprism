package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachePublic(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, false, CachePublic)
	})
}
