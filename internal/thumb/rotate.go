package thumb

import (
	"image"

	"github.com/disintegration/imaging"
)

const (
	OrientationUnspecified int = 0
	OrientationNormal          = 1
	OrientationFlipH           = 2
	OrientationRotate180       = 3
	OrientationFlipV           = 4
	OrientationTranspose       = 5
	OrientationRotate270       = 6
	OrientationTransverse      = 7
	OrientationRotate90        = 8
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
