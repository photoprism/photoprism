package config

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/pwa"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// AppName returns the app name when installed on a device.
func (c *Config) AppName() string {
	name := strings.TrimSpace(c.options.AppName)

	if name == "" {
		name = c.SiteTitle()
	}

	name = strings.Map(func(r rune) rune {
		switch r {
		case '\'', '"':
			return -1
		}

		return r
	}, name)

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

// AppIcon returns the app icon when installed on a device.
func (c *Config) AppIcon() string {
	defaultIcon := "logo"

	if c.options.AppIcon == "" || c.options.AppIcon == defaultIcon {
		// Default.
	} else if strings.Contains(c.options.AppIcon, "/") {
		return c.options.AppIcon
	} else if fs.FileExists(c.AppIconsPath(c.options.AppIcon, "16.png")) {
		return c.options.AppIcon
	}

	return defaultIcon
}

// AppColor returns the app splash screen color when installed on a device.
func (c *Config) AppColor() string {
	if appColor := clean.Color(c.options.AppColor); appColor == "" {
		return "#000000"
	} else {
		return appColor
	}
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

// AppConfig returns the progressive web app config.
func (c *Config) AppConfig() pwa.Config {
	return pwa.Config{
		Icon:        c.AppIcon(),
		Color:       c.AppColor(),
		Name:        c.AppName(),
		Description: c.SiteDescription(),
		Mode:        c.AppMode(),
		BaseUri:     c.BaseUri("/"),
		StaticUri:   c.StaticUri(),
	}
}

// AppManifest returns the progressive web app manifest.
func (c *Config) AppManifest() *pwa.Manifest {
	if cacheData, ok := Cache.Get(CacheKeyAppManifest); ok {
		log.Tracef("config: cache hit for %s", CacheKeyAppManifest)

		return cacheData.(*pwa.Manifest)
	}
	result := pwa.NewManifest(c.AppConfig())
	if result != nil {
		Cache.SetDefault(CacheKeyAppManifest, result)
	} else {
		log.Warnf("config: web app manifest is nil - you may have found a bug")
	}
	return result
}
