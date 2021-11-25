package config

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// AppName returns the app name when installed on a device.
func (c *Config) AppName() string {
	name := strings.TrimSpace(c.options.AppName)

	if name == "" {
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

// AppIcon returns the app icon when installed on a device.
func (c *Config) AppIcon() string {
	defaultIcon := "default"

	if c.options.AppIcon == "" || c.options.AppIcon == defaultIcon {
		// Default.
	} else if fs.FileExists(c.AppIconsPath(c.options.AppIcon, "512.png")) {
		return c.options.AppIcon
	}

	return defaultIcon
}

// AppIconsPath returns the path to the app icons.
func (c *Config) AppIconsPath(name ...string) string {
	if len(name) > 0 {
		folder := []string{c.StaticPath(), "icons"}
		folder = append(folder, name...)
		return filepath.Join(folder...)
	}

	return filepath.Join(c.StaticPath(), "icons")
}
