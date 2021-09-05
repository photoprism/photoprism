package crop

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FromThumb returns a cropped area from an existing thumbnail image.
func FromThumb(thumbName string, area Area, size Size, cache bool) (img image.Image, err error) {
	// Use same folder for caching if "cache" is true.
	cacheFolder := filepath.Dir(thumbName)

	// Use existing thumbnail name as cached crop filename prefix.
	thumbBase := filepath.Base(thumbName)
	if i := strings.Index(thumbBase, "_"); i > 0 {
		thumbBase = thumbBase[:i]
	}

	// Compose cached crop image file name.
	cacheBase := fmt.Sprintf("%s_%dx%d_crop_%s", thumbBase, size.Width, size.Height, area.String())
	cropFile := filepath.Join(cacheFolder, cacheBase+fs.JpegExt)

	// Cached?
	if !fs.FileExists(cropFile) {
		// Do nothing.
	} else if img, err := imaging.Open(cropFile); err != nil {
		log.Errorf("crop: failed loading %s", filepath.Base(cropFile))
	} else {
		return img, nil
	}

	// Open image.
	imageBuffer, err := ioutil.ReadFile(thumbName)
	img, err = imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))

	if err != nil {
		return img, err
	}

	// Get absolute crop coordinates and dimension.
	min, max, dim := area.Bounds(img)

	if dim < size.Width {
		log.Debugf("crop: %s too small, crop size %dpx, actual size %dpx", filepath.Base(thumbName), size.Width, dim)
	}

	// Crop area from image.
	img = imaging.Crop(img, image.Rect(min.X, min.Y, max.X, max.Y))

	// Resample crop area.
	img = thumb.Resample(img, size.Width, size.Height, size.Options...)

	// Cache crop image?
	if cache {
		if err := imaging.Save(img, cropFile); err != nil {
			log.Errorf("crop: failed caching %s", filepath.Base(cropFile))
		} else {
			log.Debugf("crop: saved %s", filepath.Base(cropFile))
		}
	}

	return img, nil
}
