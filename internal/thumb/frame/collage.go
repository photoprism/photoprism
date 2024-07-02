package frame

import (
	"fmt"
	"image"
	"image/color"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/clean"
)

// CollageBackground is the default background for image collages.
var CollageBackground = color.NRGBA{32, 33, 36, 255}

// Collage embeds images into a collage and returns the resulting image.
func Collage(t Type, images []image.Image) (collage image.Image, err error) {
	width := 1600
	height := 900

	collage = imaging.New(width, height, CollageBackground)

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

// polaroidCollage embeds images into a Polaroid collage and returns the resulting image.
func polaroidCollage(collage image.Image, images []image.Image) (image.Image, error) {
	n := len(images) - 1

	if n == 1 {
		if framed, err := polaroid(images[0], RandomAngle(20)); err != nil {
			return collage, err
		} else {
			collage = imaging.Overlay(collage, framed, image.Pt(50, -80), 1)
		}

		if framed, err := polaroid(images[1], RandomAngle(20)); err != nil {
			return collage, err
		} else {
			collage = imaging.Overlay(collage, framed, image.Pt(500, -30), 1)
		}
	} else {
		dl := 1500 / n
		dr := 1350 / n

		for i := 0; i < n; i++ {
			img := images[i+1]

			framed, err := polaroid(img, RandomAngle(30))

			if err != nil {
				return collage, err
			}

			collage = imaging.Overlay(collage, framed, RandomPoint(850-i*dl, -150-((i%2)*50), 950-i*dr, 125-((i%2)*125)), 1)
		}

		if framed, err := polaroid(images[0], RandomAngle(15)); err != nil {
			return collage, err
		} else {
			collage = imaging.Overlay(collage, framed, image.Pt(275, -50), 1)
		}
	}

	return collage, nil
}
