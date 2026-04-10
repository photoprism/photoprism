package thumb

import (
	"image"
	"image/color"
)

// EXIF orientation values.
const (
	OrientationUnspecified int = 0
	OrientationNormal      int = 1
	OrientationFlipH       int = 2
	OrientationRotate180   int = 3
	OrientationFlipV       int = 4
	OrientationTranspose   int = 5
	OrientationRotate270   int = 6
	OrientationTransverse  int = 7
	OrientationRotate90    int = 8
)

// Rotate rotates an image based on the Exif orientation.
func Rotate(img image.Image, o int) image.Image {
	src := cloneImage(img)

	switch o {
	case OrientationUnspecified:
		return src
	case OrientationNormal:
		return src
	case OrientationFlipH:
		return flipImageHorizontal(src)
	case OrientationFlipV:
		return flipImageVertical(src)
	case OrientationRotate90:
		return rotateImage90(src)
	case OrientationRotate180:
		return rotateImage180(src)
	case OrientationRotate270:
		return rotateImage270(src)
	case OrientationTranspose:
		return transposeImage(src)
	case OrientationTransverse:
		return transverseImage(src)
	default:
		log.Debugf("thumb: invalid orientation %d (rotate)", o)
		return src
	}
}

// flipImageHorizontal mirrors an image along the vertical axis.
func flipImageHorizontal(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, width-1-x, y, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// flipImageVertical mirrors an image along the horizontal axis.
func flipImageVertical(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, x, height-1-y, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// rotateImage90 rotates an image 90 degrees counterclockwise.
func rotateImage90(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, y, width-1-x, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// rotateImage180 rotates an image 180 degrees.
func rotateImage180(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, width-1-x, height-1-y, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// rotateImage270 rotates an image 270 degrees counterclockwise.
func rotateImage270(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, height-1-y, x, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// transposeImage mirrors an image along the top-left to bottom-right diagonal.
func transposeImage(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, y, x, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// transverseImage mirrors an image along the top-right to bottom-left diagonal.
func transverseImage(src *image.NRGBA) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, height, width))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			setPixel(dst, height-1-y, width-1-x, src.NRGBAAt(x, y))
		}
	}

	return dst
}

// setPixel stores a pixel in an NRGBA image.
func setPixel(dst *image.NRGBA, x, y int, c color.NRGBA) {
	dst.SetNRGBA(x, y, c)
}
