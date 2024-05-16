package thumb

import (
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/disintegration/imaging"
)

// Supported downscaling filter types.
const (
	ResampleBlackman ResampleFilter = "blackman"
	ResampleLanczos  ResampleFilter = "lanczos"
	ResampleCubic    ResampleFilter = "cubic"
	ResampleLinear   ResampleFilter = "linear"
	ResampleNearest  ResampleFilter = "nearest"
)

// Filter specifies the default downscaling filter.
var Filter = ResampleLanczos

// ResampleFilter represents a downscaling filter.
type ResampleFilter string

// Imaging returns the downscaling filter for use with the "imaging" library.
func (a ResampleFilter) Imaging() imaging.ResampleFilter {
	switch a {
	case ResampleBlackman:
		return imaging.Blackman
	case ResampleLanczos:
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
	case ResampleLanczos:
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
