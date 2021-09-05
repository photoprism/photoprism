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

// Filenames of usable thumb sizes.
var thumbFileNames = []string{
	"%s_720x720_fit.jpg",
	"%s_1280x1024_fit.jpg",
	"%s_1920x1200_fit.jpg",
	"%s_2048x2048_fit.jpg",
	"%s_4096x4096_fit.jpg",
	"%s_7680x4320_fit.jpg",
}

// Usable thumb file sizes.
var thumbFileSizes = []thumb.Size{
	thumb.Sizes[thumb.Fit720],
	thumb.Sizes[thumb.Fit1280],
	thumb.Sizes[thumb.Fit1920],
	thumb.Sizes[thumb.Fit2048],
	thumb.Sizes[thumb.Fit4096],
	thumb.Sizes[thumb.Fit7680],
}

// FromThumb returns a cropped area from an existing thumbnail image.
func FromThumb(fileName string, area Area, size Size, cache bool) (img image.Image, err error) {
	// Use same folder for caching if "cache" is true.
	filePath := filepath.Dir(fileName)

	// Extract hash from file name.
	hash := thumbHash(fileName)

	// Compose cached crop image file name.
	cropBase := fmt.Sprintf("%s_%dx%d_crop_%s%s", hash, size.Width, size.Height, area.String(), fs.JpegExt)
	cropName := filepath.Join(filePath, cropBase)

	// Cached?
	if !fs.FileExists(cropName) {
		// Do nothing.
	} else if img, err := imaging.Open(cropName); err != nil {
		log.Errorf("crop: failed loading %s", filepath.Base(cropName))
	} else {
		return img, nil
	}

	// Open thumb image file.
	img, err = openIdealThumbFile(fileName, hash, area, size)

	if err != nil {
		return img, err
	}

	// Get absolute crop coordinates and dimension.
	min, max, dim := area.Bounds(img)

	if dim < size.Width {
		log.Debugf("crop: %s too small, crop size %dpx, actual size %dpx", filepath.Base(fileName), size.Width, dim)
	}

	// Crop area from image.
	img = imaging.Crop(img, image.Rect(min.X, min.Y, max.X, max.Y))

	// Resample crop area.
	img = thumb.Resample(img, size.Width, size.Height, size.Options...)

	// Cache crop image?
	if cache {
		if err := imaging.Save(img, cropName); err != nil {
			log.Errorf("crop: failed caching %s", filepath.Base(cropName))
		} else {
			log.Debugf("crop: saved %s", filepath.Base(cropName))
		}
	}

	return img, nil
}

// thumbHash returns the thumb filename base without extension and size.
func thumbHash(fileName string) (base string) {
	base = filepath.Base(fileName)

	// Example: 01244519acf35c62a5fea7a5a7dcefdbec4fb2f5_1280x1024_fit.jpg
	i := strings.Index(base, "_")

	if i <= 0 {
		return fs.StripExt(base)
	}

	return base[:i]
}

// idealThumbFileName returns the filename of the ideal thumb size for the given width.
func idealThumbFileName(fileName, hash string, width int) string {
	filePath := filepath.Dir(fileName)

	for i, s := range thumbFileSizes {
		if s.Width < width {
			continue
		}

		name := filepath.Join(filePath, fmt.Sprintf(thumbFileNames[i], hash))

		if fs.FileExists(name) {
			return name
		}
	}

	return fileName
}

// openIdealThumbFile opens the thumbnail file and returns an image.
func openIdealThumbFile(fileName, hash string, area Area, size Size) (image.Image, error) {
	if len(hash) != 40 || area.W <= 0 || size.Width <= 0 {
		// Not a standard thumb name with sha1 hash prefix.
		if imageBuffer, err := ioutil.ReadFile(fileName); err != nil {
			return nil, err
		} else {
			return imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))
		}
	}

	minWidth := int(float32(size.Width) / area.W)

	if imageBuffer, err := ioutil.ReadFile(idealThumbFileName(fileName, hash, minWidth)); err != nil {
		return nil, err
	} else {
		return imaging.Decode(bytes.NewReader(imageBuffer))
	}
}
