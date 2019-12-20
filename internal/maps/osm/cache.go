package osm

import (
	"time"

	"github.com/melihmucuk/geocache"
)

var geoCache *geocache.Cache

func init() {
	c, err := geocache.NewCache(time.Hour, 5*time.Minute, geocache.WithIn1M)

	if err != nil {
		log.Panicf("osm: %s", err.Error())
	}

	geoCache = c
}
