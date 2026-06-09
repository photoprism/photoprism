package thumb

import (
	"fmt"
	"image"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Open loads a natively supported image file from disk, rotates it, and converts the color profile if necessary.
func Open(fileName string, orientation int) (result image.Image, err error) {
	// Filename missing?
	if fileName == "" {
		return result, fmt.Errorf("filename missing")
	}

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return result, err
	}

	// Open JPEG image with color processing?
	if Color != ColorNone && fs.FileType(fileName) == fs.ImageJpeg {
		return OpenJpeg(fileName, orientation)
	}

	// Open file with a reader bounded to the actual file size.
	img, _, err := fs.DecodeImageFile(fileName)

	if err != nil {
		return result, err
	}

	// Adjust orientation.
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	return img, nil
}
