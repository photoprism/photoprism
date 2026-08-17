package get

import (
	"sync"

	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/config/ttl"
)

var onceFolderCache sync.Once

func initFolderCache() {
	services.FolderCache = gc.New(ttl.CacheCollection.Duration(), ttl.CacheCollectionCleanup.Duration())
}

// FolderCache returns the shared folder listing cache instance.
func FolderCache() *gc.Cache {
	onceFolderCache.Do(initFolderCache)

	return services.FolderCache
}
