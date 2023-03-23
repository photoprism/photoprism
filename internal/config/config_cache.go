package config

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var Cache = gc.New(time.Hour, 15*time.Minute)

const (
	CacheKeyAppManifest  = "app-manifest"
	CacheKeyWallpaperUri = "wallpaper-uri"
)

// FlushCache clears the config cache.
func FlushCache() {
	Cache.Flush()
}
