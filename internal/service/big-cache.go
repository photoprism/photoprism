package service

import (
	"sync"
	"time"

	"github.com/allegro/bigcache"
)

var onceBigCache sync.Once

func initBigCache() {
	var err error

	services.BigCache, err = bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour))

	if err != nil {
		log.Error(err)
	}
}

func BigCache() *bigcache.BigCache {
	onceBigCache.Do(initBigCache)

	return services.BigCache
}
