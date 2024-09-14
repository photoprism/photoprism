package crop

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FromRequest returns the crop file name for an image hash, and creates it if needed.
func FromRequest(hash, area string, size Size, thumbPath string) (fileName string, err error) {
	if fileName, err = FromCache(hash, area, size, thumbPath); err == nil {
		return fileName, err
	}

	a := AreaFromString(area)

	thumbName, err := ThumbFileName(hash, a, size, thumbPath)

	if err != nil {
		return "", err
	}

	// Compose cached crop image file name.
	cropBase := fmt.Sprintf("%s_%dx%d_crop_%s%s", hash, size.Width, size.Height, area, fs.ExtJPEG)
	cropName := filepath.Join(filepath.Dir(thumbName), cropBase)

	imageBuffer, err := os.ReadFile(thumbName)

	if err != nil {
		return "", err
	}

	img, err := imaging.Decode(bytes.NewReader(imageBuffer))

	if err != nil {
		return "", err
	}

	// Get absolute crop coordinates and dimension.
	min, max, dim := a.Bounds(img)

	if dim < size.Width {
		log.Debugf("crop: %s is too small, upscaling %dpx to %dpx", filepath.Base(thumbName), dim, size.Width)
	}

	// Crop area from image.
	img = imaging.Crop(img, image.Rect(min.X, min.Y, max.X, max.Y))

	// Resample crop area.
	img = thumb.Resample(img, size.Width, size.Height, size.Options...)

	// Save crop image.
	if err := imaging.Save(img, cropName); err != nil {
		log.Errorf("failed saving %s - no permission or disk full?", filepath.Base(cropName))
		log.Debug(err.Error())
	} else {
		log.Debugf("saved %s", filepath.Base(cropName))
	}

	return cropName, nil
}
