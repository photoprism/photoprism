package frame

import (
	"fmt"
	"image"
	"image/color"

	"github.com/photoprism/photoprism/pkg/clean"
)

// CollageBackground is the default background for image collages.
var CollageBackground = color.NRGBA{32, 33, 36, 255}

// Collage embeds images into a collage and returns the resulting image.
func Collage(t Type, images []image.Image) (collage image.Image, err error) {
	width := 1600
	height := 900

	collage = newCanvas(width, height, CollageBackground)

	if len(images) == 0 {
		return collage, nil
	}

	switch t {
	case Polaroid:
		collage, err = polaroidCollage(collage, images)
	default:
		return collage, fmt.Errorf("unknown collage type %s", clean.Log(string(t)))
	}

	return collage, err
}

// polaroidCollage embeds images into a Polaroid collage and returns the resulting image. The layout
// is chosen by the number of images, so each count has a placement of its own and the background is
// returned unchanged when there is none.
func polaroidCollage(collage image.Image, images []image.Image) (image.Image, error) {
	switch len(images) {
	case 0:
		return collage, nil
	case 1:
		if framed, err := polaroid(images[0], RandomAngle(15)); err != nil {
			return collage, err
		} else {
			collage = overlayImage(collage, framed, image.Pt(275, -50))
		}
	case 2:
		if framed, err := polaroid(images[0], RandomAngle(20)); err != nil {
			return collage, err
		} else {
			collage = overlayImage(collage, framed, image.Pt(50, -80))
		}

		if framed, err := polaroid(images[1], RandomAngle(20)); err != nil { //nolint:gosec // the case holds two images
			return collage, err
		} else {
			collage = overlayImage(collage, framed, image.Pt(500, -30))
		}
	default:
		n := len(images) - 1
		dl := 1500 / n
		dr := 1350 / n

		for i := range n {
			img := images[i+1]

			framed, err := polaroid(img, RandomAngle(30))

			if err != nil {
				return collage, err
			}

			collage = overlayImage(collage, framed, RandomPoint(850-i*dl, -150-((i%2)*50), 950-i*dr, 125-((i%2)*125)))
		}

		if framed, err := polaroid(images[0], RandomAngle(15)); err != nil {
			return collage, err
		} else {
			collage = overlayImage(collage, framed, image.Pt(275, -50))
		}
	}

	return collage, nil
}
