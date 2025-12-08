package thumb

import (
	"github.com/davidbyttow/govips/v2/vips"
)

// VipsRotate rotates a vips image based on the Exif orientation.
func VipsRotate(img *vips.ImageRef, orientation int) error {
	var err error

	switch orientation {
	case OrientationUnspecified:
		// Do nothing.
	case OrientationNormal:
		// Do nothing.
	case OrientationFlipH:
		err = img.Flip(vips.DirectionHorizontal)
	case OrientationFlipV:
		err = img.Flip(vips.DirectionVertical)
	case OrientationRotate90:
		// Rotate the image 90 degrees counter-clockwise.
		err = img.Rotate(vips.Angle270)
	case OrientationRotate180:
		err = img.Rotate(vips.Angle180)
	case OrientationRotate270:
		// Rotate the image 270 degrees counter-clockwise.
		err = img.Rotate(vips.Angle90)
	case OrientationTranspose:
		err = img.Flip(vips.DirectionHorizontal)
		if err == nil {
			// Rotate the image 90 degrees counter-clockwise.
			err = img.Rotate(vips.Angle270)
		}
	case OrientationTransverse:
		err = img.Flip(vips.DirectionVertical)
		if err == nil {
			// Rotate the image 90 degrees counter-clockwise.
			err = img.Rotate(vips.Angle270)
		}
	default:
		log.Debugf("vips: invalid orientation %d (rotate image)", orientation)
	}

	return err
}
