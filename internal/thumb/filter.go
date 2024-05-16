package thumb

import "github.com/disintegration/imaging"

// Supported downscaling filter types.
const (
	ResampleBlackman ResampleFilter = "blackman"
	ResampleLanczos  ResampleFilter = "lanczos"
	ResampleCubic    ResampleFilter = "cubic"
	ResampleLinear   ResampleFilter = "linear"
)

// Filter specifies the default downscaling filter.
var Filter = ResampleLanczos

// ResampleFilter represents a downscaling filter.
type ResampleFilter string

// Imaging returns the downscaling filter for use with the "disintegration/imaging" library.
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
	default:
		return imaging.Lanczos
	}
}
