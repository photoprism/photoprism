package thumb

import (
	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/clean"
)

// ResampleFilter represents a downscaling filter.
type ResampleFilter string

// Supported downscaling filter types.
const (
	ResampleAuto     ResampleFilter = "auto"
	ResampleBlackman ResampleFilter = "blackman"
	ResampleLanczos  ResampleFilter = "lanczos"
	ResampleCubic    ResampleFilter = "cubic"
	ResampleLinear   ResampleFilter = "linear"
	ResampleNearest  ResampleFilter = "nearest"
)

// String returns the downscaling filter name as string.
func (a ResampleFilter) String() string {
	return string(a)
}

// Vips returns the downscaling filter for use with the "govips" library.
func (a ResampleFilter) Vips() vips.Kernel {
	switch a {
	case ResampleBlackman:
		return vips.KernelLanczos3
	case ResampleLanczos, ResampleAuto:
		return vips.KernelLanczos3
	case ResampleCubic:
		return vips.KernelCubic
	case ResampleLinear:
		return vips.KernelLinear
	case ResampleNearest:
		return vips.KernelNearest
	default:
		return vips.KernelLanczos3
	}
}

// ParseFilter returns a ResampleFilter based on the config value string and image library.
func ParseFilter(name string, lib Lib) ResampleFilter {
	if lib == LibVips {
		return ResampleAuto
	}

	filter := ResampleFilter(clean.TypeLowerUnderscore(name))

	switch filter {
	case ResampleBlackman, ResampleLanczos, ResampleCubic, ResampleLinear, ResampleNearest:
		return filter
	default:
		return ResampleLanczos
	}
}
