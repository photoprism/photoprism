package get

import (
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
)

var onceFolderCache sync.Once

func initFolderCache() {
	services.FolderCache = gc.New(time.Minute*15, 5*time.Minute)
}

func FolderCache() *gc.Cache {
	onceFolderCache.Do(initFolderCache)

	return services.FolderCache
}
