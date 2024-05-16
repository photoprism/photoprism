package thumb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Vips generates a thumbnail image file with libvips, see https://github.com/libvips/libvips.
func Vips(imageFilename, hash, thumbPath string, width, height, orientation int, opts ...ResampleOption) (fileName string, err error) {
	if fileName, err = FromCache(imageFilename, hash, thumbPath, width, height, opts...); err == nil {
		return fileName, err
	} else if !errors.Is(err, ErrNotCached) {
		return "", err
	}

	if InvalidSize(width) {
		return "", fmt.Errorf("thumb: width has an invalid value (%d)", width)
	}

	if InvalidSize(height) {
		return "", fmt.Errorf("thumb: height has an invalid value (%d)", height)
	}

	// Generate thumb cache filename.
	fileName, err = FileName(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Debugf("vips: %s in %s (generate thumbnail filename)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	// Initialize libvips before using it.
	VipsInit()

	// Load image from file.
	img, err := vips.NewImageFromFile(imageFilename)

	if err != nil {
		log.Debugf("vips: %s in %s (new image from file)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	// Adjust orientation.
	if img, err = VipsRotate(img, orientation); err != nil {
		log.Debugf("vips: %s in %s (rotate image)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	var crop vips.Interesting

	method, _, _ := ResampleOptions(opts...)

	if method == ResampleFit {
		crop = vips.InterestingNone
	} else if method == ResampleFillCenter || method == ResampleResize {
		crop = vips.InterestingAttention
	} else if method == ResampleFillTopLeft {
		crop = vips.InterestingHigh
	} else if method == ResampleFillBottomRight {
		crop = vips.InterestingLow
	}

	// Create thumbnail image.
	if err = img.Thumbnail(width, height, crop); err != nil {
		log.Debugf("vips: %s in %s (create thumbnail)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	// Remove metadata from thumbnail.
	if err = img.RemoveMetadata(); err != nil {
		log.Debugf("vips: %s in %s (remove metadata)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	var imageBytes []byte

	// Export to PNG or JPEG.
	if fs.FileType(fileName) == fs.ImagePNG {
		params := vips.NewPngExportParams()
		imageBytes, _, err = img.ExportPng(params)
	} else {
		params := vips.NewJpegExportParams()

		if width <= 150 && height <= 150 {
			params.Quality = JpegQualitySmall.Int()
		} else {
			params.Quality = JpegQuality.Int()
		}

		imageBytes, _, err = img.ExportJpeg(params)
	}

	if err != nil {
		log.Debugf("vips: %s in %s (export thumbnail)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	// Write thumbnail to file.
	if err = os.WriteFile(fileName, imageBytes, fs.ModeFile); err != nil {
		log.Debugf("vips: %s in %s (write thumbnail to file)", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	return fileName, nil
}

// VipsRotate rotates a vips image based on the Exif orientation.
func VipsRotate(img *vips.ImageRef, o int) (*vips.ImageRef, error) {
	var err error

	switch o {
	case OrientationUnspecified:
		err = img.AutoRotate()
	case OrientationNormal:
		// Do nothing.
	case OrientationFlipH:
		err = img.Flip(vips.DirectionHorizontal)
	case OrientationFlipV:
		err = img.Flip(vips.DirectionVertical)
	case OrientationRotate90:
		err = img.Rotate(vips.Angle90)
	case OrientationRotate180:
		err = img.Rotate(vips.Angle180)
	case OrientationRotate270:
		err = img.Rotate(vips.Angle270)
	case OrientationTranspose:
		err = img.Flip(vips.DirectionHorizontal)
		if err == nil {
			err = img.Rotate(vips.Angle270)
		}
	case OrientationTransverse:
		err = img.Flip(vips.DirectionVertical)
		if err == nil {
			err = img.Rotate(vips.Angle270)
		}
	default:
		log.Debugf("vips: invalid orientation %d (rotate image)", o)
	}

	return img, err
}
