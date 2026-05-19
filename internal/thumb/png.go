package thumb

import (
	"fmt"
	"image"
	"path/filepath"
	"runtime/debug"

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

	img, err = vipsConvert(srcFile, pngFile, orientation)
	if err != nil {
		log.Debugf("png: failed to open %s", clean.Log(filepath.Base(srcFile)))
		return img, err
	}

	// Return PNG image.
	return img, nil
}
