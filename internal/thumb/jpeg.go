package thumb

import (
	"image"
	"path/filepath"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Jpeg converts an image to jpeg, saves and returns it.
func Jpeg(srcFilename, jpgFilename string, orientation int) (img image.Image, err error) {
	// Resolve symlinks.
	if srcFilename, err = fs.Resolve(srcFilename); err != nil {
		log.Debugf("jpeg: %s in %s (resolve source image filename)", err, clean.Log(srcFilename))
		return img, err
	}

	// Open source image.
	img, err = imaging.Open(srcFilename)

	// Failed?
	if err != nil {
		log.Errorf("jpeg: cannot open source image %s", clean.Log(filepath.Base(srcFilename)))
		return img, err
	}

	// Adjust orientation.
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	// Get JPEG quality setting.
	quality := JpegQuality.EncodeOption()

	// Save JPEG file.
	if err = imaging.Save(img, jpgFilename, quality); err != nil {
		log.Errorf("jpeg: failed to save %s", clean.Log(filepath.Base(jpgFilename)))
		return img, err
	}

	// Return JPEG image.
	return img, nil
}
