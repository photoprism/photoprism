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

// vipsConvert loads a source image with libvips, applies the explicit orientation, and exports it.
// Unlike thumbnail generation, format conversion preserves source metadata where libvips can carry it over.
//
// Orientation handling for HEIF/AVIF (see strukturag/libheif#227):
// libheif always applies ISOBMFF irot/imir transforms during decode, and the HEIF spec treats
// EXIF orientation as informational only.  There is no reliable metadata signal from libheif
// that distinguishes "irot was applied" from "no irot was present," so we follow the spec:
// when the image was loaded through libheif (detected via the vips-loader field), we never
// apply the caller's EXIF orientation.  This is correct for conformant files and avoids
// double-rotation for all transform types including square rotations and pure flips.
//
// Trade-off: some older Apple HEIC files (e.g. iPhone 7) carry EXIF orientation without an
// irot box, which is non-conformant per the HEIF spec.  These files will not be auto-rotated
// by this path.  The libheif maintainer recommends treating this as the spec-correct behavior
// rather than introducing heuristics that risk corrupting conformant files.
func vipsConvert(srcFile, dstFile string, orientation int) (_ image.Image, err error) {
	VipsInit()

	img, err := vips.LoadImageFromFile(srcFile, vipsConvertImportParams())
	if err != nil {
		return nil, err
	}
	defer img.Close()

	// Apply orientation — but not for images loaded through libheif.
	//
	// libheif applies all ISOBMFF irot/imir pixel transforms during decode and the HEIF
	// spec says EXIF orientation is informational only (strukturag/libheif#227).  There is
	// no post-load signal that reliably distinguishes "irot applied" from "no irot present"
	// for every transform type (dimension-preserving rotations, flips), so we follow the
	// spec and skip explicit rotation entirely for heifload images.
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
