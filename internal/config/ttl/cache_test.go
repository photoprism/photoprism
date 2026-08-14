package ttl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		assert.Equal(t, Duration(365*24*3600), CacheMaxAge)
		assert.Greater(t, CacheMaxAge, CacheDefault)
		assert.Greater(t, CacheDefault, CacheVideo)
		assert.Greater(t, CacheVideo, CacheCover)
		assert.Greater(t, CacheCover, CacheCollection)
		assert.Greater(t, CacheCollection, CacheCollectionCleanup)
	})
	t.Run("Duration", func(t *testing.T) {
		assert.Equal(t, 15*time.Minute, CacheCollection.Duration())
		assert.Equal(t, 5*time.Minute, CacheCollectionCleanup.Duration())
	})
}
