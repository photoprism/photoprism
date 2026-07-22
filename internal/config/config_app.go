package config

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/config/pwa"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// DefaultAppColor specifies the default app background and splash screen color.
var DefaultAppColor = "#19191a"

// AppName returns the app name shown when installed as a PWA, preferring an explicit
// AppName, then the distinctive SiteName (SITE_NAME), then the SiteTitle.
func (c *Config) AppName() string {
	name := strings.TrimSpace(c.options.AppName)

	if name == "" {
		name = c.SiteName()
	}

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

	if c.options.AppIcon != "" && c.options.AppIcon != defaultIcon {
		if themeIcon := filepath.Join(c.ThemePath(), c.options.AppIcon); fs.FileExistsNotEmpty(themeIcon) {
			return path.Join(ThemeUri, c.options.AppIcon)
		} else if strings.Contains(c.options.AppIcon, "/") {
			return c.options.AppIcon
		} else if fs.FileExistsNotEmpty(c.AppIconsPath(c.options.AppIcon, "16.png")) {
			return c.options.AppIcon
		}
	}

	return defaultIcon
}

// AppTouchIcon returns the icon set used for the iOS home-screen apple-touch-icon. Every built-in
// set ships a full-bleed "touch" variant whose background fills the whole square opaquely: iOS
// masks the icon into a squircle and, during the app-open zoom, composites it over the light
// launch background, so a source with transparent rounded corners would flash white around it.
// Custom theme/URL icons have no variant and fall back to the built-in "app" touch squircle.
func (c *Config) AppTouchIcon() string {
	icon := c.AppIcon()

	if fs.FileExistsNotEmpty(c.AppIconsPath(icon, "touch", "180.png")) {
		return path.Join(icon, "touch")
	}

	// Every built-in icon set ships a full-bleed touch variant, so this fallback is reached only
	// by custom theme/URL icons (which contain a slash): they have no touch variant and cannot
	// drive the sized ladder, so they use the built-in "app" touch squircle.
	return path.Join("app", "touch")
}

// AppColor returns the app background and splash screen color.
func (c *Config) AppColor() string {
	if appColor := clean.Color(c.options.AppColor); appColor == "" {
		return DefaultAppColor
	} else {
		return appColor
	}
}

// AppIconsPath returns the path to the app icons.
func (c *Config) AppIconsPath(name ...string) string {
	if len(name) > 0 {
		filePath := []string{c.StaticPath(), fs.IconsDir}
		filePath = append(filePath, name...)
		return filepath.Join(filePath...)
	}

	return filepath.Join(c.StaticPath(), fs.IconsDir)
}

// AppConfig returns the progressive web app config.
func (c *Config) AppConfig() pwa.Config {
	return pwa.Config{
		Icon:          c.AppIcon(),
		Color:         c.AppColor(),
		Name:          c.AppName(),
		Description:   c.SiteDescription(),
		DefaultLocale: c.DefaultLocale(),
		Mode:          c.AppMode(),
		BaseUri:       c.BaseUri("/"),
		FrontendUri:   c.FrontendUri(""),
		StaticUri:     c.StaticUri(),
		StaticPath:    c.StaticPath(),
		SiteUrl:       c.SiteUrl(),
		CdnUrl:        c.CdnUrl("/"),
		ThemeUri:      ThemeUri,
		ThemePath:     c.ThemePath(),
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
		log.Warnf("config: no web app manifest returned - you may have found a bug")
	}

	return result
}
