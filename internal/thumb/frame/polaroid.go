package frame

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
)

//go:embed polaroid.png
var polaroidPng []byte

// Polaroid embeds the specified image file into a Polaroid frame and returns the resulting image.
func polaroid(img image.Image, rotate float64) (image.Image, error) {
	// Create image frame.
	frm, err := imaging.Decode(bytes.NewReader(polaroidPng))

	if err != nil {
		return nil, err
	}

	// Paste image into frame.
	out := imaging.Paste(frm, img, image.Pt(200, 152))

	// Rotate image before returning it?
	if rotate != 0.0 {
		out = imaging.Rotate(out, rotate, color.NRGBA{255, 255, 255, 0})
	}

	return out, nil
}
