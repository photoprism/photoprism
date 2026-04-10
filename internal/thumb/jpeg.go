package thumb

import (
	"fmt"
	"image"
	"path/filepath"
	"runtime/debug"

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

	img, err = vipsConvert(srcFile, jpgFile, orientation)
	if err != nil {
		log.Debugf("jpeg: failed to open %s", clean.Log(filepath.Base(srcFile)))
		return img, err
	}

	// Return JPEG image.
	return img, nil
}
