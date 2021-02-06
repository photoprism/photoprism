package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var countryCache = gc.New(time.Hour, 15*time.Minute)

func FlushCountryCache() {
	countryCache.Flush()
}
