package thumb

import (
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder for config reads
	_ "image/jpeg" // register JPEG decoder for config reads
	_ "image/png"  // register PNG decoder for config reads
	"runtime/debug"

	_ "golang.org/x/image/bmp"  // register BMP decoder for config reads
	_ "golang.org/x/image/webp" // register WEBP decoder for config reads

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FileInfo returns the image header info containing width and height.
func FileInfo(fileName string) (info image.Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic %s while decoding %s file info\nstack: %s", r, clean.Log(fileName), debug.Stack())
		}
	}()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return info, err
	}

	// Decode image config (dimensions) with a reader bounded to the file size.
	info, _, err = fs.DecodeImageConfigFile(fileName)

	if err != nil {
		return info, fmt.Errorf("%s while decoding file info", err)
	}

	return info, err
}
