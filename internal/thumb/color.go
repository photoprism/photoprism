package thumb

import "github.com/photoprism/photoprism/pkg/clean"

// ColorSpace represents a thumbnail color profile preference.
type ColorSpace = string

// Supported thumbnail color profile settings.
const (
	ColorNone     ColorSpace = "none"
	ColorAuto     ColorSpace = "auto"
	ColorSRGB     ColorSpace = "srgb"
	ColorPreserve ColorSpace = "preserve"
)

// ParseColor returns a ColorSpace based on the config value string and image library.
func ParseColor(name string, lib Lib) ColorSpace {
	if lib == LibVips {
		return ColorPreserve
	}

	switch clean.TypeLowerUnderscore(name) {
	case ColorNone, "":
		return ColorNone
	default:
		return ColorSRGB
	}
}
