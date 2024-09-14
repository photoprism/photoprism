package frame

import (
	"fmt"
	"image"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Image embeds the specified image file into a frame and returns the resulting image.
func Image(t Type, img image.Image, rotate float64) (image.Image, error) {
	switch t {
	case Polaroid:
		return polaroid(img, rotate)
	default:
		return img, fmt.Errorf("unknown collage type %s", clean.Log(string(t)))
	}
}
