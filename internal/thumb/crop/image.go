package crop

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Filenames of usable thumb sizes.
var thumbFileNames = []string{
	"%s_720x720_fit.jpg",
	"%s_1280x1024_fit.jpg",
	"%s_1920x1200_fit.jpg",
	"%s_2560x1600_fit.jpg",
	"%s_4096x4096_fit.jpg",
	"%s_7680x4320_fit.jpg",
}

// Suitable thumb file sizes.
var thumbFileSizes = []thumb.Size{
	thumb.Sizes[thumb.Fit720],
	thumb.Sizes[thumb.Fit1280],
	thumb.Sizes[thumb.Fit1920],
	thumb.Sizes[thumb.Fit2560],
	thumb.Sizes[thumb.Fit4096],
	thumb.Sizes[thumb.Fit7680],
}

// ImageFromThumb returns a cropped area from an existing thumbnail image.
func ImageFromThumb(thumbName string, area Area, size Size, cache bool) (img image.Image, err error) {
	// Use same folder for caching if "cache" is true.
	filePath := filepath.Dir(thumbName)

	// Extract hash from file name.
	hash := thumbHash(thumbName)

	// Resolve symlinks.
	if thumbName, err = fs.Resolve(thumbName); err != nil {
		return nil, err
	}

	// Compose cached crop image file name.
	cropBase := fmt.Sprintf("%s_%dx%d_crop_%s%s", hash, size.Width, size.Height, area.String(), fs.ExtJPEG)
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
	img, err = openIdealThumbFile(thumbName, hash, area, size)

	if err != nil {
		return img, err
	}

	// Get absolute crop coordinates and dimension.
	min, max, dim := area.Bounds(img)

	if dim < size.Width {
		log.Debugf("crop: %s is too small, upscaling %dpx to %dpx", filepath.Base(thumbName), dim, size.Width)
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

// ThumbFileName returns the ideal thumb file name.
func ThumbFileName(hash string, area Area, size Size, thumbPath string) (string, error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("invalid file hash %s", clean.Log(hash))
	}

	if len(thumbPath) < 1 {
		return "", fmt.Errorf("cache path missing")
	}

	if area.W <= 0 {
		return "", fmt.Errorf("invalid area width %f", area.W)
	}

	if size.Width <= 0 {
		return "", fmt.Errorf("invalid crop size %d", size.Width)
	}

	filePath := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])
	fileName := findIdealThumbFileName(hash, area.FileWidth(size), filePath)

	if fileName == "" {
		return "", fmt.Errorf("not found")
	}

	// Resolve symlinks.
	return fs.Resolve(fileName)
}

// FileWidth returns the minimal thumbnail width based on crop area and size.
func FileWidth(area Area, size Size) int {
	return int(float32(size.Width) / area.W)
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

// findIdealThumbFileName finds the filename of the ideal thumb size for the given width.
func findIdealThumbFileName(hash string, width int, filePath string) (fileName string) {
	if hash == "" || filePath == "" {
		return ""
	}

	for i, s := range thumbFileSizes {
		// Resolve symlinks.
		name, err := fs.Resolve(filepath.Join(filePath, fmt.Sprintf(thumbFileNames[i], hash)))

		if err != nil || !fs.FileExists(name) {
			continue
		} else if s.Width < width {
			fileName = name
			continue
		} else {
			return name
		}
	}

	return fileName
}

// openIdealThumbFile opens the thumbnail file and returns an image.
func openIdealThumbFile(fileName, hash string, area Area, size Size) (result image.Image, err error) {
	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return nil, err
	}

	if len(hash) != 40 || area.W <= 0 || size.Width <= 0 {
		// Not a standard thumb name with sha1 hash prefix.
		if imageBuffer, err := os.ReadFile(fileName); err != nil {
			return nil, err
		} else {
			return imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))
		}
	}

	if name := findIdealThumbFileName(hash, area.FileWidth(size), filepath.Dir(fileName)); name != "" {
		fileName = name
	}

	if imageBuffer, err := os.ReadFile(fileName); err != nil {
		return nil, err
	} else {
		return imaging.Decode(bytes.NewReader(imageBuffer))
	}
}
