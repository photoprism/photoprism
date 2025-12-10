package thumb

import (
	"image"

	"github.com/disintegration/imaging"
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
	switch o {
	case OrientationUnspecified:
		// Do nothing.
	case OrientationNormal:
		// Do nothing.
	case OrientationFlipH:
		img = imaging.FlipH(img)
	case OrientationFlipV:
		img = imaging.FlipV(img)
	case OrientationRotate90:
		img = imaging.Rotate90(img)
	case OrientationRotate180:
		img = imaging.Rotate180(img)
	case OrientationRotate270:
		img = imaging.Rotate270(img)
	case OrientationTranspose:
		img = imaging.Transpose(img)
	case OrientationTransverse:
		img = imaging.Transverse(img)
	default:
		log.Debugf("thumb: invalid orientation %d (rotate)", o)
	}

	return img
}
