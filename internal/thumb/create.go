package thumb

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Suffix returns the thumb cache file suffix.
func Suffix(width, height int, opts ...ResampleOption) (result string) {
	method, _, format := ResampleOptions(opts...)

	result = fmt.Sprintf("%dx%d_%s.%s", width, height, ResampleMethods[method], format)

	return result
}

// FileName returns the thumb cache file name based on path, size, and options.
func FileName(hash string, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if InvalidSize(width) {
		return "", fmt.Errorf("resample: width exceeds limit (%d)", width)
	}

	if InvalidSize(height) {
		return "", fmt.Errorf("resample: height exceeds limit (%d)", height)
	}

	if len(hash) < 4 {
		return "", fmt.Errorf("resample: file hash is empty or too short (%s)", sanitize.Log(hash))
	}

	if len(thumbPath) == 0 {
		return "", errors.New("resample: folder is empty")
	}

	suffix := Suffix(width, height, opts...)
	p := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])

	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return "", err
	}

	fileName = fmt.Sprintf("%s/%s_%s", p, hash, suffix)

	return fileName, nil
}

// FromCache returns the thumb cache file name for an image.
func FromCache(imageFilename, hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("resample: invalid file hash %s", sanitize.Log(hash))
	}

	if len(imageFilename) < 4 {
		return "", fmt.Errorf("resample: invalid file name %s", sanitize.Log(imageFilename))
	}

	fileName, err = FileName(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Error(err)
		return "", err
	}

	if fs.FileExists(fileName) {
		return fileName, nil
	}

	return "", ErrThumbNotCached
}

// FromFile returns the thumb cache file name for an image, and creates it if needed.
func FromFile(imageFilename, hash, thumbPath string, width, height, orientation int, opts ...ResampleOption) (fileName string, err error) {
	if fileName, err := FromCache(imageFilename, hash, thumbPath, width, height, opts...); err == nil {
		return fileName, err
	} else if err != ErrThumbNotCached {
		return "", err
	}

	// Generate thumb cache filename.
	fileName, err = FileName(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Error(err)
		return "", err
	}

	// Load image from storage.
	img, err := Open(imageFilename, orientation)

	if err != nil {
		log.Error(err)
		return "", err
	}

	// Create thumb from image.
	if _, err := Create(img, fileName, width, height, opts...); err != nil {
		return "", err
	}

	return fileName, nil
}

// Create creates an image thumbnail.
func Create(img image.Image, fileName string, width, height int, opts ...ResampleOption) (result image.Image, err error) {
	if InvalidSize(width) {
		return img, fmt.Errorf("resample: width has an invalid value (%d)", width)
	}

	if InvalidSize(height) {
		return img, fmt.Errorf("resample: height has an invalid value (%d)", height)
	}

	result = Resample(img, width, height, opts...)

	var saveOption imaging.EncodeOption

	if filepath.Ext(fileName) == "."+string(fs.FormatPng) {
		saveOption = imaging.PNGCompressionLevel(png.DefaultCompression)
	} else if width <= 150 && height <= 150 {
		saveOption = imaging.JPEGQuality(JpegQualitySmall)
	} else {
		saveOption = imaging.JPEGQuality(JpegQuality)
	}

	err = imaging.Save(result, fileName, saveOption)

	if err != nil {
		log.Errorf("resample: failed to save %s", sanitize.Log(filepath.Base(fileName)))
		return result, err
	}

	return result, nil
}
