package crop

import (
	"fmt"
	"image"
	"image/draw"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Filenames of usable thumb sizes, ordered so a rendition is never narrower than its predecessor.
var thumbFileNames = []string{
	"%s_720x720_fit.jpg",
	"%s_1280x1024_fit.jpg",
	"%s_1920x1200_fit.jpg",
	"%s_2560x1600_fit.jpg",
	"%s_4096x4096_fit.jpg",
	"%s_5120x5120_fit.jpg",
	"%s_7680x4320_fit.jpg",
	"%s_15360x8640_fit.jpg",
}

// Suitable thumb file sizes, in the same order as thumbFileNames.
var thumbFileSizes = []thumb.Size{
	thumb.Sizes[thumb.Fit720],
	thumb.Sizes[thumb.Fit1280],
	thumb.Sizes[thumb.Fit1920],
	thumb.Sizes[thumb.Fit2560],
	thumb.Sizes[thumb.Fit4096],
	thumb.Sizes[thumb.Fit5120],
	thumb.Sizes[thumb.Fit7680],
	thumb.Sizes[thumb.Fit15360],
}

// WidestCachedSize returns the widest rendition a crop can be taken from, which is the widest
// usable size that is still pre-generated at the configured thumbnail limit.
//
// The selection below stats cached files, so a size above the limit is rendered on demand, never
// written to disk, and therefore unreachable here however much detail the original holds. The
// zero size is returned when the limit excludes even the smallest.
func WidestCachedSize() (widest thumb.Size) {
	for _, s := range thumbFileSizes {
		if s.Uncached() {
			break
		}

		widest = s
	}

	return widest
}

// UsableSizes returns the renditions a crop can be taken from, in ascending order and whether or
// not the configured limit pre-generates them. A caller asking what a larger limit would deliver
// needs the ones above it, which is why this is not filtered.
func UsableSizes() []thumb.Size {
	return slices.Clone(thumbFileSizes)
}

// CachedSizeExists reports whether the rendition of the specified size is on disk for a file hash,
// stating it from the same names the selection walks.
//
// Not through thumb.Size.FileName, which refuses a size above the configured limit: the selection
// stats whatever exists, so a rendition written while the limit was higher is still read from, and
// a caller asking what the cache holds has to be able to see it. An empty file is not one a crop
// can be taken from - a write interrupted by a signal or a full volume leaves one behind, and the
// selection would hand it to a decoder that cannot read it.
func CachedSizeExists(size thumb.Size, hash, thumbPath string) bool {
	if len(hash) < 4 || thumbPath == "" {
		return false
	}

	for i, s := range thumbFileSizes {
		if s.Name != size.Name {
			continue
		}

		filePath := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])
		name, err := fs.Resolve(filepath.Join(filePath, fmt.Sprintf(thumbFileNames[i], hash)))

		return err == nil && fs.FileExistsNotEmpty(name)
	}

	return false
}

// ImageFromThumb returns a cropped area from an existing thumbnail image, reusing a cached crop
// when one exists. srcWidth is then 0, because a reused crop cannot say what it was drawn from.
func ImageFromThumb(thumbName string, area Area, size Size, cache bool) (img image.Image, cropName string, srcWidth int, err error) {
	return cropFromThumb(thumbName, area, size, cache, true)
}

// ImageFromSource returns a cropped area taken from the source rendition, ignoring any cached crop,
// and reports the width it sampled.
//
// A caller that records what an embedding was drawn from has to use this: the crop cache is keyed
// on hash, area and size alone, so the UI's own face thumbnails satisfy it and would otherwise
// leave every such embedding unmeasured.
func ImageFromSource(thumbName string, area Area, size Size, cache bool) (img image.Image, cropName string, srcWidth int, err error) {
	return cropFromThumb(thumbName, area, size, cache, false)
}

func cropFromThumb(thumbName string, area Area, size Size, cache, reuse bool) (img image.Image, cropName string, srcWidth int, err error) {
	// Use same folder for caching if "cache" is true.
	filePath := filepath.Dir(thumbName)

	// Extract hash from file name.
	hash := thumbHash(thumbName)

	// Resolve symlinks.
	if thumbName, err = fs.Resolve(thumbName); err != nil {
		return nil, "", 0, err
	}

	// Compose cached crop image file name.
	cropBase := fmt.Sprintf("%s_%dx%d_crop_%s%s", hash, size.Width, size.Height, area.String(), fs.ExtJpeg)
	cropName = filepath.Join(filePath, cropBase)

	// Cached?
	if !reuse {
		// Do nothing.
	} else if !fs.FileExists(cropName) {
		// Do nothing.
	} else if cropImg, _, cropErr := fs.DecodeImageFile(cropName); cropErr != nil {
		log.Errorf("crop: failed loading %s", filepath.Base(cropName))
	} else {
		// Zero rather than resolved: the crop's name records its area and dimensions but not its
		// source, and a rendition cached since would predict a wider one than this crop came from.
		return cropImg, cropName, 0, nil
	}

	// Open thumb image file.
	img, srcName, err := openIdealThumbFile(thumbName, hash, area, size)

	if err != nil {
		return img, "", srcWidth, err
	}

	// Exact rather than resolved: this is the image the crop is taken from, whatever selection
	// produced it, including a path that is not a standard thumbnail name.
	srcWidth = img.Bounds().Dx()

	// Get absolute crop coordinates and dimension.
	posMin, posMax, dim := area.Bounds(img)

	// The rendition that was opened, which the selection may have swapped for a wider one: naming
	// the requested file instead reports an upscale from a source that was never read.
	if dim < size.Width {
		log.Debugf("crop: %s is too small, upscaling %dpx to %dpx", filepath.Base(srcName), dim, size.Width)
	}

	// Crop area from image.
	img = imageCrop(img, image.Rect(posMin.X, posMin.Y, posMax.X, posMax.Y))

	// Resample crop area.
	img = thumb.Resample(img, size.Width, size.Height, size.Options...)

	// Cache crop image?
	if cache {
		if err = thumb.Save(img, cropName); err != nil {
			log.Errorf("crop: failed caching %s", filepath.Base(cropName))
		} else {
			log.Debugf("crop: saved %s", filepath.Base(cropName))
		}
	}

	return img, cropName, srcWidth, nil
}

// ImageFromIdealThumb decodes the smallest cached thumbnail that can still supply the
// requested crop size for an area, falling back to the given thumbnail when there is
// none. Callers that need the whole image rather than the crop use this, so that a face
// warped onto a template is not upscaled from a rendition it outgrew.
func ImageFromIdealThumb(thumbName string, area Area, size Size) (img image.Image, err error) {
	img, _, err = openIdealThumbFile(thumbName, thumbHash(thumbName), area, size)

	return img, err
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

// findIdealThumbFileName returns the smallest cached thumbnail that covers the given width, or the
// widest one cached when none does. A name bounds a rendition from above but not below - the file
// called 1920x1200 holds a 4:3 picture 1600 px wide - so a candidate the name allows is measured.
func findIdealThumbFileName(hash string, width int, filePath string) (fileName string) {
	if hash == "" || filePath == "" {
		return ""
	}

	for i, s := range thumbFileSizes {
		// Resolve symlinks.
		name, err := fs.Resolve(filepath.Join(filePath, fmt.Sprintf(thumbFileNames[i], hash)))

		if err != nil || !fs.FileExists(name) {
			continue
		}

		// Renditions grow with the list, so the last one seen is the widest cached.
		fileName = name

		if s.Width < width {
			continue
		} else if cfg, _, cfgErr := fs.DecodeImageConfigFile(name); cfgErr == nil && cfg.Width < width {
			continue
		}

		return name
	}

	return fileName
}

// openIdealThumbFile opens the thumbnail file and returns the image with the name it was read
// from, which is not the name that was asked for whenever the selection found a wider rendition.
func openIdealThumbFile(fileName, hash string, area Area, size Size) (result image.Image, opened string, err error) {
	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return nil, "", err
	}

	if len(hash) != 40 || area.W <= 0 || size.Width <= 0 {
		// Not a standard thumb name with sha1 hash prefix.
		result, _, err = fs.DecodeImageFile(fileName)
		return result, fileName, err
	}

	if name := findIdealThumbFileName(hash, area.FileWidth(size), filepath.Dir(fileName)); name != "" {
		fileName = name
	}

	result, _, err = fs.DecodeImageFile(fileName)

	return result, fileName, err
}

// imageCrop returns a copy of the requested crop rectangle.
func imageCrop(img image.Image, rect image.Rectangle) image.Image {
	rect = rect.Intersect(img.Bounds())

	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	cropped := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), img, rect.Min, draw.Src)

	return cropped
}
