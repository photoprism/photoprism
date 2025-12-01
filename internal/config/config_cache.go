package config

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

// Cache stores shared config values for quick reuse across requests.
var Cache = gc.New(time.Hour, 15*time.Minute)

const (
	// CacheKeyAppManifest is the cache key for the PWA manifest.
	CacheKeyAppManifest = "app-manifest"
	// CacheKeyWallpaperUri is the cache key for the current wallpaper URI.
	CacheKeyWallpaperUri = "wallpaper-uri"
)

// FlushCache clears the config cache.
func FlushCache() {
	Cache.Flush()
}
