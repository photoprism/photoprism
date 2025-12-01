package thumb

import (
	"image"

	"github.com/disintegration/imaging"
)

// Resample downscales an image and returns it.
func Resample(img image.Image, width, height int, opts ...ResampleOption) image.Image {
	var resImg image.Image

	method, filter, _ := ResampleOptions(opts...)

	switch method {
	case ResampleFit:
		resImg = imaging.Fit(img, width, height, filter.Imaging())
	case ResampleFillCenter:
		resImg = imaging.Fill(img, width, height, imaging.Center, filter.Imaging())
	case ResampleFillTopLeft:
		resImg = imaging.Fill(img, width, height, imaging.TopLeft, filter.Imaging())
	case ResampleFillBottomRight:
		resImg = imaging.Fill(img, width, height, imaging.BottomRight, filter.Imaging())
	case ResampleResize:
		resImg = imaging.Resize(img, width, height, filter.Imaging())
	}

	return resImg
}
