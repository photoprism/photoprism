package thumb

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// vipsConvertImportParams provides libvips import parameters for image conversion helpers.
func vipsConvertImportParams() *vips.ImportParams {
	params := &vips.ImportParams{}
	params.FailOnError.Set(false)
	return params
}

// vipsConvert loads a source image with libvips, applies the explicit
// orientation, and exports it. EXIF orientation is skipped for HEIF/AVIF
// (libheif already applied any irot/imir during decode — the HEIF spec treats
// EXIF orientation as informational only). Non-conformant HEIC files that
// carry EXIF orientation without irot will not be auto-rotated by this path.
func vipsConvert(srcFile, dstFile string, orientation int) (_ image.Image, err error) {
	VipsInit()

	img, err := vips.LoadImageFromFile(srcFile, vipsConvertImportParams())
	if err != nil {
		return nil, err
	}
	defer img.Close()

	// Skip EXIF orientation for HEIF/AVIF — libheif already applied irot/imir during decode.
	if orientation > OrientationNormal && !vipsLoadedViaHeif(img) {
		if err = VipsRotate(img, orientation); err != nil {
			return nil, err
		}
	}

	if err = img.RemoveOrientation(); err != nil {
		return nil, err
	}

	width, height := img.Width(), img.Height()

	var imageBytes []byte
	switch fs.FileType(dstFile) {
	case fs.ImagePng:
		params := VipsPngExportParams(width, height)
		imageBytes, _, err = img.ExportPng(params)
	default:
		params := VipsJpegExportParams(width, height)
		imageBytes, _, err = img.ExportJpeg(params)
	}

	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(dstFile, imageBytes, fs.ModeFile); err != nil {
		return nil, err
	}

	decoded, _, err := fs.DecodeImageData(imageBytes)
	if err != nil {
		return nil, fmt.Errorf("vips: %s in %s (decode exported image)", err, clean.Log(filepath.Base(dstFile)))
	}

	return decoded, nil
}

// vipsLoadedViaHeif reports whether the image was decoded by the libheif loader.
func vipsLoadedViaHeif(img *vips.ImageRef) bool {
	loader := img.GetString("vips-loader")
	return len(loader) >= 8 && loader[:8] == "heifload"
}
