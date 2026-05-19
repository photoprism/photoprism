package thumb

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Save writes an image to disk using the file extension to choose the encoder.
func Save(img image.Image, fileName string, quality ...Quality) error {
	if img == nil {
		return fmt.Errorf("thumb: image is nil")
	}

	dirName := filepath.Dir(fileName)
	baseName := filepath.Base(fileName)
	fileQuality := JpegQualityDefault

	if len(quality) > 0 && quality[0] > 0 {
		fileQuality = quality[0]
	}

	tmpFile, err := os.CreateTemp(dirName, "."+baseName+".tmp-*")
	if err != nil {
		return err
	}

	tmpName := tmpFile.Name()
	renamed := false

	defer func() {
		_ = tmpFile.Close()

		if !renamed {
			_ = os.Remove(tmpName)
		}
	}()

	if err = tmpFile.Chmod(fs.ModeFile); err != nil {
		return err
	}

	if err = encodeImageFile(tmpFile, img, fileName, fileQuality); err != nil {
		return err
	}

	if err = tmpFile.Close(); err != nil {
		return err
	}

	if err = os.Rename(tmpName, fileName); err != nil {
		return err
	}

	renamed = true

	return nil
}

// encodeImageFile writes an image to an open file using the correct encoder for the destination type.
func encodeImageFile(file *os.File, img image.Image, fileName string, quality Quality) error {
	switch fs.FileType(fileName) {
	case fs.ImagePng:
		return encodePNG(file, img)
	default:
		return encodeJPEG(file, img, quality)
	}
}

// encodePNG writes an image as PNG.
func encodePNG(file *os.File, img image.Image) error {
	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
	return encoder.Encode(file, img)
}

// encodeJPEG writes an image as JPEG.
func encodeJPEG(file *os.File, img image.Image, quality Quality) error {
	return jpeg.Encode(file, img, &jpeg.Options{Quality: quality.Int()})
}

// cloneImage normalizes an image to an origin-based NRGBA image.
func cloneImage(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	clone := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(clone, clone.Bounds(), img, bounds.Min, draw.Src)
	return clone
}

// cropImage returns a cropped copy of the requested rectangle.
func cropImage(img image.Image, rect image.Rectangle) image.Image {
	rect = rect.Intersect(img.Bounds())

	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	cropped := image.NewNRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), img, rect.Min, draw.Src)

	return cropped
}

// resizeImage rescales an image to the requested size, preserving aspect ratio when one dimension is zero.
func resizeImage(img image.Image, width, height int, filter ResampleFilter) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	width, height = normalizeResizeTarget(srcWidth, srcHeight, width, height)

	if width <= 0 || height <= 0 {
		return cloneImage(img)
	}

	if width == srcWidth && height == srcHeight {
		return cloneImage(img)
	}

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	resampleScaler(filter).Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)

	return dst
}

// fitImage rescales an image to fit inside the requested bounds without cropping.
func fitImage(img image.Image, width, height int, filter ResampleFilter) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 || width <= 0 || height <= 0 {
		return cloneImage(img)
	}

	scale := math.Min(float64(width)/float64(srcWidth), float64(height)/float64(srcHeight))
	scale = math.Min(scale, 1)

	if scale <= 0 {
		return cloneImage(img)
	}

	targetWidth := maxInt(1, int(float64(srcWidth)*scale))
	targetHeight := maxInt(1, int(float64(srcHeight)*scale))

	return resizeImage(img, targetWidth, targetHeight, filter)
}

// fillImage rescales an image to cover the requested bounds and then crops it.
func fillImage(img image.Image, width, height int, filter ResampleFilter, anchor cropAnchor) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 || width <= 0 || height <= 0 {
		return cloneImage(img)
	}

	scale := math.Max(float64(width)/float64(srcWidth), float64(height)/float64(srcHeight))

	if scale <= 0 {
		return cloneImage(img)
	}

	targetWidth := maxInt(width, int(math.Ceil(float64(srcWidth)*scale)))
	targetHeight := maxInt(height, int(math.Ceil(float64(srcHeight)*scale)))
	resized := resizeImage(img, targetWidth, targetHeight, filter)
	left, top := anchorOffset(resized.Bounds().Dx(), resized.Bounds().Dy(), width, height, anchor)

	return cropImage(resized, image.Rect(left, top, left+width, top+height))
}

// normalizeResizeTarget derives the missing resize dimension when only one target edge is provided.
func normalizeResizeTarget(srcWidth, srcHeight, width, height int) (int, int) {
	switch {
	case width <= 0 && height <= 0:
		return srcWidth, srcHeight
	case width <= 0:
		width = maxInt(1, int(math.Round(float64(srcWidth)*float64(height)/float64(srcHeight))))
	case height <= 0:
		height = maxInt(1, int(math.Round(float64(srcHeight)*float64(width)/float64(srcWidth))))
	}

	return width, height
}

// resampleScaler maps configured filters to x/image scalers for in-memory fallback work.
func resampleScaler(filter ResampleFilter) xdraw.Scaler {
	switch filter {
	case ResampleLinear:
		return xdraw.ApproxBiLinear
	case ResampleNearest:
		return xdraw.NearestNeighbor
	default:
		return xdraw.CatmullRom
	}
}

// cropAnchor defines how Fill-style crops are aligned after resizing.
type cropAnchor int

const (
	// cropAnchorCenter keeps the crop centered on both axes.
	cropAnchorCenter cropAnchor = iota
	// cropAnchorTopLeft keeps the crop pinned to the top-left corner.
	cropAnchorTopLeft
	// cropAnchorBottomRight keeps the crop pinned to the bottom-right corner.
	cropAnchorBottomRight
)

// anchorOffset calculates the crop origin for the requested anchor.
func anchorOffset(srcWidth, srcHeight, width, height int, anchor cropAnchor) (left, top int) {
	switch anchor {
	case cropAnchorTopLeft:
		return 0, 0
	case cropAnchorBottomRight:
		return maxInt(0, srcWidth-width), maxInt(0, srcHeight-height)
	default:
		return maxInt(0, (srcWidth-width)/2), maxInt(0, (srcHeight-height)/2)
	}
}

// maxInt returns the larger of two ints.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
