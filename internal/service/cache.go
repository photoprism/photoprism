package service

import (
	"sync"
	"time"

	gc "github.com/patrickmn/go-cache"
)

var onceCache sync.Once

func initCache() {
	services.Cache = gc.New(336*time.Hour, 30*time.Minute)
}

func Cache() *gc.Cache {
	onceCache.Do(initCache)

	return services.Cache
}
