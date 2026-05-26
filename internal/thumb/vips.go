package thumb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Vips generates a new thumbnail with the requested size and returns the file name and a buffer with the image bytes,
// or an error if thumbnail generation failed. For more information on libvips, see https://github.com/libvips/libvips.
func Vips(imageName string, imageBuffer []byte, hash, thumbPath string, width, height int, opts ...ResampleOption) (thumbName string, thumbBuffer []byte, err error) {
	if len(hash) < 4 {
		return "", nil, fmt.Errorf("thumb: invalid file hash %s", clean.Log(hash))
	}

	if len(imageName) < 4 {
		return "", nil, fmt.Errorf("thumb: invalid file name %s", clean.Log(imageName))
	}

	if InvalidSize(width) {
		return "", nil, fmt.Errorf("thumb: width has an invalid value (%d)", width)
	}

	if InvalidSize(height) {
		return "", nil, fmt.Errorf("thumb: height has an invalid value (%d)", height)
	}

	// Get thumb cache filename.
	thumbName, err = FileName(hash, thumbPath, width, height, opts...)

	logName := clean.Log(filepath.Base(imageName))

	if err != nil {
		log.Debugf("thumb: %s in %s (filename)", err, logName)
		return "", nil, err
	}

	// Initialize libvips before using it.
	VipsInit()

	// Load image from file or buffer.
	var img *vips.ImageRef

	if len(imageBuffer) == 0 {
		if img, err = vips.LoadImageFromFile(imageName, VipsImportParams()); err != nil {
			log.Debugf("vips: %s in %s (load image from file)", err, logName)
			return "", nil, err
		}
	} else if img, err = vips.LoadImageFromBuffer(imageBuffer, VipsImportParams()); err != nil {
		log.Debugf("vips: %s in %s (load image from buffer)", err, logName)
		return "", nil, err
	}
	defer img.Close()

	// Set resample options.
	var method ResampleOption
	var size vips.Size

	method, _, _ = ResampleOptions(opts...)

	// Choose thumbnail crop.
	var crop vips.Interesting
	switch method {
	case ResampleFillTopLeft:
		crop = vips.InterestingLow
		size = vips.SizeBoth
	case ResampleFillBottomRight:
		crop = vips.InterestingHigh
		size = vips.SizeBoth
	case ResampleFit:
		crop = vips.InterestingNone
		size = vips.SizeDown
	case ResampleFillCenter, ResampleResize:
		crop = vips.InterestingCentre
		size = vips.SizeBoth
	default:
		// Use defaults.
	}

	// Embed an ICC profile when a JPEG declares its color space via the EXIF InteroperabilityIndex tag.
	if err = vipsSetIccProfileForInteropIndex(img, logName); err != nil {
		log.Debugf("vips: %s in %s (set icc profile for interop index tag)", err, logName)
	}

	// Create thumbnail image.
	if err = img.ThumbnailWithSize(width, height, crop, size); err != nil {
		log.Debugf("vips: %s in %s (create thumbnail)", err, logName)
		return "", nil, err
	}

	// Guard against zero-dimension intermediates so callers see the real cause
	// instead of an opaque libvips "unable to write to target" further downstream.
	if w, h := img.Width(), img.Height(); w <= 0 || h <= 0 {
		err = fmt.Errorf("vips: produced empty %dx%d image for %s", w, h, logName)
		log.Debugf("%s (create thumbnail)", err)
		return "", nil, err
	}

	// Remove metadata from thumbnail.
	if err = img.RemoveMetadata(); err != nil {
		log.Debugf("vips: %s in %s (remove metadata)", err, logName)
		return "", nil, err
	}

	// Export to standard image format.
	var format string
	switch fs.FileType(thumbName) {
	case fs.ImagePng:
		format = "png"
		pngParams := VipsPngExportParams(width, height)
		// If ResampleStripICC is set, remove the ICC profile.
		if Options(opts).Contains(ResampleStripICC) && img.HasICCProfile() {
			if iccErr := img.RemoveICCProfile(); iccErr != nil {
				log.Debugf("vips: %s in %s (remove icc profile)", iccErr, logName)
			}
		}
		// Try to export PNG thumbnail image.
		thumbBuffer, _, err = img.ExportPng(pngParams)
		// If that fails, try again without the ICC profile, since libpng may reject an invalid ICCP chunk (e.g. malformed profile length).
		if err != nil && img.HasICCProfile() {
			log.Tracef("vips: %s in %s (export png with icc)", err, logName)
			if iccErr := img.RemoveICCProfile(); iccErr != nil {
				log.Debugf("vips: %s in %s (remove icc profile)", iccErr, logName)
			} else if thumbBuffer, _, err = img.ExportPng(pngParams); err != nil {
				log.Debugf("vips: %s in %s (export png without icc)", err, logName)
			}
		}
	default:
		format = "jpeg"
		thumbBuffer, _, err = img.ExportJpeg(VipsJpegExportParams(width, height))
	}

	// Check if export failed.
	if err != nil {
		err = wrapVipsExportErr(format, thumbName, width, height, err)
		log.Debugf("%s (export thumbnail)", err)
		return "", thumbBuffer, err
	}

	// Write thumbnail to file.
	if err = os.WriteFile(thumbName, thumbBuffer, fs.ModeFile); err != nil {
		err = wrapVipsWriteErr(thumbName, err)
		log.Debugf("%s (write thumbnail to file)", err)
		return "", thumbBuffer, err
	}

	return thumbName, thumbBuffer, nil
}

// wrapVipsExportErr annotates a libvips export error with the target format, filename, and dimensions.
func wrapVipsExportErr(format, thumbName string, width, height int, err error) error {
	return fmt.Errorf("vips: %s export failed for %s at %dx%d (%w)",
		format, clean.Log(filepath.Base(thumbName)), width, height, err)
}

// wrapVipsWriteErr annotates a thumbnail file-write error with the destination filename.
func wrapVipsWriteErr(thumbName string, err error) error {
	return fmt.Errorf("vips: failed to write thumbnail %s (%w)", clean.Log(filepath.Base(thumbName)), err)
}

// VipsImportParams provides parameters for opening files with libvips.
func VipsImportParams() *vips.ImportParams {
	params := &vips.ImportParams{}
	params.AutoRotate.Set(true)
	params.FailOnError.Set(false)
	return params
}

// VipsPngExportParams returns PNG image encoding parameters for libvips.
func VipsPngExportParams(width, height int) *vips.PngExportParams {
	params := vips.NewPngExportParams()
	params.Filter = vips.PngFilterNone
	params.Interlace = false
	params.Palette = false

	// Set compression depending on image size.
	if width > 20 || height > 20 {
		params.Compression = 6
	} else {
		params.Compression = 0
	}

	return params
}

// VipsJpegExportParams returns JPEG image encoding parameters for libvips.
func VipsJpegExportParams(width, height int) *vips.JpegExportParams {
	params := vips.NewJpegExportParams()
	params.Quality = JpegQuality(width, height).Int()
	params.Interlace = true
	params.SubsampleMode = vips.VipsForeignSubsampleAuto

	// Enable quality enhancements depending on image size.
	if width > 150 || height > 150 {
		params.OptimizeCoding = true
		// The following options can only be set if libvips
		// has been compiled with the "--with-mozjpeg" flag:
		// params.QuantTable = 3
		// params.TrellisQuant = true
		// params.OvershootDeringing = true
		// params.OptimizeScans = true
	}

	return params
}
