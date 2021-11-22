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
	defaultIcon := "logo"

	if c.options.AppIcon == "" || c.options.AppIcon == defaultIcon {
		// Default.
	} else if fs.FileExists(filepath.Join(c.ImgPath(), "icons", c.options.AppIcon+"-192.png")) {
		return c.options.AppIcon
	}

	return defaultIcon
}
