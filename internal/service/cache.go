package service

import (
	"sync"
	"time"

	"github.com/allegro/bigcache"
)

var onceCache sync.Once

func initCache() {
	var err error

	services.Cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour))

	if err != nil {
		log.Errorf("")
	}
}

func Cache() *bigcache.BigCache {
	onceCache.Do(initCache)

	return services.Cache
}
