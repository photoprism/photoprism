package thumb

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/disintegration/imaging"
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

// Filter specifies the default downscaling filter.
var Filter = ResampleLanczos

// String returns the downscaling filter name as string.
func (a ResampleFilter) String() string {
	return string(a)
}

// Imaging returns the downscaling filter for use with the "imaging" library.
func (a ResampleFilter) Imaging() imaging.ResampleFilter {
	switch a {
	case ResampleBlackman:
		return imaging.Blackman
	case ResampleLanczos, ResampleAuto:
		return imaging.Lanczos
	case ResampleCubic:
		return imaging.CatmullRom
	case ResampleLinear:
		return imaging.Linear
	case ResampleNearest:
		return imaging.NearestNeighbor
	default:
		return imaging.Lanczos
	}
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
		return ResampleAuto
	}
}
