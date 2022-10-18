package get

import (
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
)

var onceThumbCache sync.Once

func initThumbCache() {
	services.ThumbCache = gc.New(time.Hour*24, 10*time.Minute)
}

func ThumbCache() *gc.Cache {
	onceThumbCache.Do(initThumbCache)

	return services.ThumbCache
}
