package config

import (
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// DefaultLocale returns the default user interface language locale name.
func (c *Config) DefaultLocale() string {
	if c.options.DefaultLocale == "" {
		return i18n.Default.Locale()
	}

	return c.options.DefaultLocale
}

// DefaultTimezone returns the default time zone, e.g. for scheduling backups
func (c *Config) DefaultTimezone() *time.Location {
	if c.options.DefaultTimezone == "" {
		return time.UTC
	}

	// Returns time zone if a valid identifier name was provided and UTC otherwise.
	if timeZone, err := time.LoadLocation(c.options.DefaultTimezone); err != nil {
		return time.UTC
	} else {
		return timeZone
	}
}

// DefaultTheme returns the default user interface theme name.
func (c *Config) DefaultTheme() string {
	if c.options.DefaultTheme == "" {
		return "default"
	}

	return c.options.DefaultTheme
}

// WallpaperUri returns the login screen background image URI.
func (c *Config) WallpaperUri() string {
	if cacheData, ok := Cache.Get(CacheKeyWallpaperUri); ok {
		// Return cached wallpaper URI.
		log.Tracef("config: cache hit for %s", CacheKeyWallpaperUri)

		return cacheData.(string)
	} else if strings.Contains(c.options.WallpaperUri, "/") {
		return c.options.WallpaperUri
	}

	wallpaperUri := c.options.WallpaperUri
	wallpaperPath := "img/wallpaper"

	// Default to "welcome.jpg" if value is empty and file exists.
	if wallpaperUri == "" {
		if !fs.PathExists(filepath.Join(c.StaticPath(), wallpaperPath)) {
			return ""
		}

		wallpaperUri = "welcome.jpg"
	} else if !strings.Contains(wallpaperUri, ".") {
		wallpaperUri += fs.ExtJPEG
	}

	// Complete URI as needed if file path is valid.
	if fileName := clean.Path(wallpaperUri); fileName == "" {
		return ""
	} else if fs.FileExists(path.Join(c.StaticPath(), wallpaperPath, fileName)) {
		wallpaperUri = c.StaticAssetUri(path.Join(wallpaperPath, fileName))
	} else if fs.FileExists(c.CustomStaticFile(path.Join(wallpaperPath, fileName))) {
		wallpaperUri = c.CustomStaticAssetUri(path.Join(wallpaperPath, fileName))
	} else {
		return ""
	}

	// Cache wallpaper URI if not empty.
	if wallpaperUri != "" {
		Cache.SetDefault(CacheKeyWallpaperUri, wallpaperUri)
	}

	return wallpaperUri
}

// SetWallpaperUri changes the login screen background image URI.
func (c *Config) SetWallpaperUri(uri string) *Config {
	c.options.WallpaperUri = uri
	FlushCache()
	return c
}
