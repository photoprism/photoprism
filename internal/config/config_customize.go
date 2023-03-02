package config

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// DefaultTheme returns the default user interface theme name.
func (c *Config) DefaultTheme() string {
	if c.options.DefaultTheme == "" || c.NoSponsor() {
		return "default"
	}

	return c.options.DefaultTheme
}

// DefaultLocale returns the default user interface language locale name.
func (c *Config) DefaultLocale() string {
	if c.options.DefaultLocale == "" {
		return i18n.Default.Locale()
	}

	return c.options.DefaultLocale
}

// WallpaperUri returns the login screen background image `URI`.
func (c *Config) WallpaperUri() string {
	if strings.Contains(c.options.WallpaperUri, "/") {
		return c.options.WallpaperUri
	}

	assetPath := "img/wallpaper"

	// Empty URI?
	if c.options.WallpaperUri == "" {
		if !fs.PathExists(filepath.Join(c.StaticPath(), assetPath)) {
			return ""
		}

		c.options.WallpaperUri = "welcome.jpg"
	} else if !strings.Contains(c.options.WallpaperUri, ".") {
		c.options.WallpaperUri += fs.ExtJPEG
	}

	// Valid URI? Local file?
	if p := clean.Path(c.options.WallpaperUri); p == "" {
		c.options.WallpaperUri = ""
	} else if fs.FileExists(path.Join(c.StaticPath(), assetPath, p)) {
		c.options.WallpaperUri = path.Join(c.StaticUri(), assetPath, p)
	} else if fs.FileExists(c.CustomStaticFile(path.Join(assetPath, p))) {
		c.options.WallpaperUri = path.Join(c.CustomStaticUri(), assetPath, p)
	} else {
		c.options.WallpaperUri = ""
	}

	return c.options.WallpaperUri
}
