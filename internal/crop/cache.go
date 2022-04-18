package crop

import (
	"fmt"
	"path/filepath"

	"path"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FromCache returns the crop file name if cached.
func FromCache(hash, area string, size Size, thumbPath string) (fileName string, err error) {
	fileName, err = FileName(hash, area, size.Width, size.Height, thumbPath)

	if err != nil {
		return fileName, err
	}

	if fs.FileExists(fileName) {
		return fileName, nil
	}

	return fileName, fmt.Errorf("%s not found", filepath.Base(fileName))
}

// FileName returns the crop file name based on cache path, size, and area.
func FileName(hash, area string, width, height int, thumbPath string) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("crop: invalid file hash %s", clean.Log(hash))
	}

	if len(thumbPath) < 1 {
		return "", fmt.Errorf("crop: cache path missing")
	}

	if width < 1 || height < 1 || width > 2048 || height > 2048 {
		return "", fmt.Errorf("crop: invalid size %dx%d", width, height)
	}

	fileName = path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3], fmt.Sprintf("%s_%dx%d_crop_%s%s", hash, width, height, area, fs.ExtJPEG))

	return fileName, nil
}
