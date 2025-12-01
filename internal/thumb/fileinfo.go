package thumb

import (
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder for config reads
	_ "image/jpeg" // register JPEG decoder for config reads
	_ "image/png"  // register PNG decoder for config reads
	"os"
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

	file, err := os.Open(fileName) //nolint:gosec // fileName is resolved path provided by caller; reading images is expected

	if err != nil || file == nil {
		return info, err
	}

	defer file.Close()

	// Reset file offset.
	// see https://github.com/golang/go/issues/45902#issuecomment-1007953723
	_, err = file.Seek(0, 0)

	if err != nil {
		return info, fmt.Errorf("%s on seek", err)
	}

	// Decode image config (dimensions).
	info, _, err = image.DecodeConfig(file)

	if err != nil {
		return info, fmt.Errorf("%s while decoding file info", err)
	}

	return info, err
}
