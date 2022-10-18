package get

import (
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
)

var onceCoverCache sync.Once

func initCoverCache() {
	services.CoverCache = gc.New(time.Hour, 10*time.Minute)
}

func CoverCache() *gc.Cache {
	onceCoverCache.Do(initCoverCache)

	return services.CoverCache
}
