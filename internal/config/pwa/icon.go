package pwa

import (
	"fmt"
	"strings"
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
func NewIcons(staticUri, appIcon string) Icons {
	if appIcon == "" {
		appIcon = "logo"
	} else if strings.Contains(appIcon, "/") {
		return Icons{{
			Src:  appIcon,
			Type: "image/png",
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
