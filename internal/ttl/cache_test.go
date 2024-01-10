package ttl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, Duration(365*24*3600), CacheMaxAge)
		assert.Greater(t, CacheMaxAge, CacheDefault)
		assert.Greater(t, CacheDefault, CacheVideo)
		assert.Greater(t, CacheVideo, CacheCover)
	})
}
