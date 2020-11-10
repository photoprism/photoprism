package places

import (
	"time"

	"github.com/allegro/bigcache"
)

var cache *bigcache.BigCache

func init() {
	var err error

	cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour))

	if err != nil {
		log.Errorf("")
	}
}
