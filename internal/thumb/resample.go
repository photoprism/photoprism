package thumb

import (
	"image"
)

// Resample downscales an image and returns it.
func Resample(img image.Image, width, height int, opts ...ResampleOption) image.Image {
	method, filter, _ := ResampleOptions(opts...)

	return ResampleWithFilter(img, width, height, method, filter)
}

// ResampleWithFilter resizes an image with an explicit method and interpolation filter.
func ResampleWithFilter(img image.Image, width, height int, method ResampleOption, filter ResampleFilter) image.Image {
	switch method {
	case ResampleFit:
		return fitImage(img, width, height, filter)
	case ResampleFillCenter:
		return fillImage(img, width, height, filter, cropAnchorCenter)
	case ResampleFillTopLeft:
		return fillImage(img, width, height, filter, cropAnchorTopLeft)
	case ResampleFillBottomRight:
		return fillImage(img, width, height, filter, cropAnchorBottomRight)
	case ResampleResize:
		return resizeImage(img, width, height, filter)
	default:
		return cloneImage(img)
	}
}
