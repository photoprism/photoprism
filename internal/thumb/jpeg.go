package thumb

import (
	"fmt"
	"image"
	"path/filepath"
	"runtime/debug"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Jpeg converts an image to JPEG, saves it, and returns it.
func Jpeg(srcFile, jpgFile string, orientation int) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("jpeg: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Resolve symlinks.
	if srcFile, err = fs.Resolve(srcFile); err != nil {
		log.Debugf("jpeg: %s in %s (resolve filename)", err, clean.Log(srcFile))
		return img, err
	}

	// Open source image.
	img, err = imaging.Open(srcFile)

	// Failed?
	if err != nil {
		log.Debugf("jpeg: failed to open %s", clean.Log(filepath.Base(srcFile)))
		return img, err
	}

	// Adjust orientation.
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	// Get JPEG quality setting.
	quality := JpegQuality.EncodeOption()

	// Save JPEG file.
	if err = imaging.Save(img, jpgFile, quality); err != nil {
		log.Errorf("jpeg: failed to save %s", clean.Log(filepath.Base(jpgFile)))
		return img, err
	}

	// Return JPEG image.
	return img, nil
}
