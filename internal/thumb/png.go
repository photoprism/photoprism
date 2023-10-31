package thumb

import (
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"runtime/debug"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Png converts an image to PNG, saves it, and returns it.
func Png(srcFile, pngFile string, orientation int) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("png: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	// Resolve symlinks.
	if srcFile, err = fs.Resolve(srcFile); err != nil {
		log.Debugf("png: %s in %s (resolve filename)", err, clean.Log(srcFile))
		return img, err
	}

	// Open source image.
	img, err = imaging.Open(srcFile)

	// Failed?
	if err != nil {
		log.Debugf("png: failed to open %s", clean.Log(filepath.Base(srcFile)))
		return img, err
	}

	// Adjust orientation.
	if orientation > 1 {
		img = Rotate(img, orientation)
	}

	// Save PNG file.
	if err = imaging.Save(img, pngFile, imaging.PNGCompressionLevel(png.BestCompression)); err != nil {
		log.Errorf("png: failed to save %s", clean.Log(filepath.Base(pngFile)))
		return img, err
	}

	// Return PNG image.
	return img, nil
}
