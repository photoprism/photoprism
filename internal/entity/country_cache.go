package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var countryCache = gc.New(time.Hour, 15*time.Minute)

// FlushCountryCache clears the cached countries, so the next lookup reads the database.
func FlushCountryCache() {
	countryCache.Flush()
}
