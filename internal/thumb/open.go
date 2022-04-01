package thumb

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/fs"
)

// StandardRGB configures whether colors in the Apple Display P3 color space should be converted to standard RGB.
var StandardRGB = true

// Open loads an image from disk, rotates it, and converts the color profile if necessary.
func Open(fileName string, orientation int) (result image.Image, err error) {
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	// Open JPEG?
	if StandardRGB && fs.GetFileFormat(fileName) == fs.FormatJpeg {
		return OpenJpeg(fileName, orientation)
	}

	// Open file with imaging function.
	img, err := imaging.Open(fileName)

	if err != nil {
		return result, err
	}

	// Rotate?
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}
