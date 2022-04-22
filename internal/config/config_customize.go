package config

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/i18n"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
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

// AppIcon returns the app icon when installed on a device.
func (c *Config) AppIcon() string {
	defaultIcon := "logo"

	if c.NoSponsor() || c.options.AppIcon == "" || c.options.AppIcon == defaultIcon {
		// Default.
	} else if fs.FileExists(c.AppIconsPath(c.options.AppIcon, "512.png")) {
		return c.options.AppIcon
	}

	return defaultIcon
}

// AppIconsPath returns the path to the app icons.
func (c *Config) AppIconsPath(name ...string) string {
	if len(name) > 0 {
		filePath := []string{c.StaticPath(), "icons"}
		filePath = append(filePath, name...)
		return filepath.Join(filePath...)
	}

	return filepath.Join(c.StaticPath(), "icons")
}

// AppName returns the app name when installed on a device.
func (c *Config) AppName() string {
	name := strings.TrimSpace(c.options.AppName)

	if c.NoSponsor() || name == "" {
		name = c.SiteTitle()
	}

	clean := func(r rune) rune {
		switch r {
		case '\'', '"':
			return -1
		}

		return r
	}

	name = strings.Map(clean, name)

	return txt.Clip(name, 32)
}

// AppMode returns the app mode when installed on a device.
func (c *Config) AppMode() string {
	switch c.options.AppMode {
	case "fullscreen", "standalone", "minimal-ui", "browser":
		return c.options.AppMode
	default:
		return "standalone"
	}
}

// WallpaperUri returns the login screen background image `URI`.
func (c *Config) WallpaperUri() string {
	if c.NoSponsor() {
		return ""
	} else if strings.Contains(c.options.WallpaperUri, "/") {
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
		return ""
	} else if fs.FileExists(filepath.Join(c.StaticPath(), assetPath, p)) {
		c.options.WallpaperUri = path.Join(c.StaticUri(), assetPath, p)
	} else {
		c.options.WallpaperUri = ""
	}

	return c.options.WallpaperUri
}
