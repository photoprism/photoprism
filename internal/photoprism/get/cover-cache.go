package get

import (
	"sync"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/config/ttl"
)

var onceCoverCache sync.Once

func initCoverCache() {
	services.CoverCache = gc.New(ttl.CacheCollection.Duration(), ttl.CacheCollectionCleanup.Duration())
}

// CoverCache returns the shared album cover cache instance.
func CoverCache() *gc.Cache {
	onceCoverCache.Do(initCoverCache)

	return services.CoverCache
}
