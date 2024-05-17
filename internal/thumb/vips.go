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

	if err != nil {
		log.Debugf("thumb: %s in %s (filename)", err, clean.Log(filepath.Base(imageName)))
		return "", nil, err
	}

	// Initialize libvips before using it.
	VipsInit()

	// Load image from file or buffer.
	var img *vips.ImageRef

	if len(imageBuffer) == 0 {
		if img, err = vips.LoadImageFromFile(imageName, VipsImportParams()); err != nil {
			log.Debugf("vips: %s in %s (load image from file)", err, clean.Log(filepath.Base(imageName)))
			return "", nil, err
		}
	} else if img, err = vips.LoadImageFromBuffer(imageBuffer, VipsImportParams()); err != nil {
		log.Debugf("vips: %s in %s (load image from buffer)", err, clean.Log(filepath.Base(imageName)))
		return "", nil, err
	}

	// Set resample options.
	var method ResampleOption
	var size vips.Size

	method, _, _ = ResampleOptions(opts...)

	// Choose thumbnail crop.
	var crop vips.Interesting
	if method == ResampleFillTopLeft {
		crop = vips.InterestingLow
		size = vips.SizeBoth
	} else if method == ResampleFillBottomRight {
		crop = vips.InterestingHigh
		size = vips.SizeBoth
	} else if method == ResampleFit {
		crop = vips.InterestingNone
		size = vips.SizeDown
	} else if method == ResampleFillCenter || method == ResampleResize {
		crop = vips.InterestingCentre
		size = vips.SizeBoth
	}

	// Create thumbnail image.
	if err = img.ThumbnailWithSize(width, height, crop, size); err != nil {
		log.Debugf("vips: %s in %s (create thumbnail)", err, clean.Log(filepath.Base(imageName)))
		return "", nil, err
	}

	// Remove metadata from thumbnail.
	if err = img.RemoveMetadata(); err != nil {
		log.Debugf("vips: %s in %s (remove metadata)", err, clean.Log(filepath.Base(imageName)))
		return "", nil, err
	}

	// Export to PNG or JPEG.
	if fs.FileType(thumbName) == fs.ImagePNG {
		params := vips.NewPngExportParams()
		thumbBuffer, _, err = img.ExportPng(params)
	} else {
		params := vips.NewJpegExportParams()

		if width <= 150 && height <= 150 {
			params.Quality = JpegQualitySmall.Int()
		} else {
			params.Quality = JpegQuality.Int()
		}

		thumbBuffer, _, err = img.ExportJpeg(params)
	}

	// Check if export failed.
	if err != nil {
		log.Debugf("vips: %s in %s (export thumbnail)", err, clean.Log(filepath.Base(imageName)))
		return "", thumbBuffer, err
	}

	// Write thumbnail to file.
	if err = os.WriteFile(thumbName, thumbBuffer, fs.ModeFile); err != nil {
		log.Debugf("vips: %s in %s (write thumbnail to file)", err, clean.Log(filepath.Base(imageName)))
		return "", thumbBuffer, err
	}

	return thumbName, thumbBuffer, nil
}

// VipsImportParams provides parameters for opening files with libvips.
func VipsImportParams() *vips.ImportParams {
	params := &vips.ImportParams{}
	params.AutoRotate.Set(true)
	params.FailOnError.Set(false)
	return params
}

// VipsRotate rotates a vips image based on the Exif orientation.
func VipsRotate(img *vips.ImageRef, orientation int) error {
	var err error

	switch orientation {
	case OrientationUnspecified:
		// Do nothing.
	case OrientationNormal:
		// Do nothing.
	case OrientationFlipH:
		err = img.Flip(vips.DirectionHorizontal)
	case OrientationFlipV:
		err = img.Flip(vips.DirectionVertical)
	case OrientationRotate90:
		// Rotate the image 90 degrees counter-clockwise.
		err = img.Rotate(vips.Angle270)
	case OrientationRotate180:
		err = img.Rotate(vips.Angle180)
	case OrientationRotate270:
		// Rotate the image 270 degrees counter-clockwise.
		err = img.Rotate(vips.Angle90)
	case OrientationTranspose:
		err = img.Flip(vips.DirectionHorizontal)
		if err == nil {
			// Rotate the image 90 degrees counter-clockwise.
			err = img.Rotate(vips.Angle270)
		}
	case OrientationTransverse:
		err = img.Flip(vips.DirectionVertical)
		if err == nil {
			// Rotate the image 90 degrees counter-clockwise.
			err = img.Rotate(vips.Angle270)
		}
	default:
		log.Debugf("vips: invalid orientation %d (rotate image)", orientation)
	}

	return err
}
