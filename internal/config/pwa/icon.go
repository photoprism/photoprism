package pwa

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// Icons represents a list of app icons.
type Icons []Icon

// Icon represents an app icon.
type Icon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes,omitempty"`
	Type  string `json:"type,omitempty"`
}

// IconSizes represents standard app icon sizes.
var IconSizes = []int{16, 32, 76, 114, 128, 144, 152, 160, 167, 180, 192, 196, 256, 400, 512}

// NewIcons creates new app icons in the default sizes based on the parameters provided.
func NewIcons(c Config) Icons {
	staticUri := c.StaticUri
	appIcon := c.Icon

	if appIcon == "" {
		appIcon = "logo"
	} else if c.ThemePath != "" && strings.HasPrefix(appIcon, c.ThemeUri) {
		var appIconSize string
		var appIconType string

		if fileName := strings.Replace(appIcon, c.ThemeUri, c.ThemePath, 1); !fs.FileExistsNotEmpty(fileName) {
			appIconSize = "32x32"
			appIconType = "image/png"
		} else {
			if info, err := thumb.FileInfo(fileName); err == nil {
				appIconSize = fmt.Sprintf("%dx%d", info.Width, info.Height)
			}

			if mimeType := fs.MimeType(fileName); mimeType != "" {
				appIconType = mimeType
			}
		}

		return Icons{{
			Src:   appIcon,
			Sizes: appIconSize,
			Type:  appIconType,
		}}
	} else if strings.Contains(appIcon, "/") {
		var appIconType string

		switch fs.FileType(filepath.Base(appIcon)) {
		case fs.ImageJpeg:
			appIconType = header.ContentTypeJpeg
		case fs.ImageWebp:
			appIconType = header.ContentTypeWebp
		case fs.ImageAvif:
			appIconType = header.ContentTypeAvif
		default:
			appIconType = "image/png"
		}

		return Icons{{
			Src:  appIcon,
			Type: appIconType,
		}}
	}

	icons := make(Icons, len(IconSizes))

	for i, d := range IconSizes {
		icons[i] = Icon{
			Src:   fmt.Sprintf("%s/icons/%s/%d.png", staticUri, appIcon, d),
			Sizes: fmt.Sprintf("%dx%d", d, d),
			Type:  "image/png",
		}
	}

	return icons
}
