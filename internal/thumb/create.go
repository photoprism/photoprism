package thumb

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"path"
	"path/filepath"

	"github.com/disintegration/imaging"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Suffix returns the thumb cache file suffix.
func Suffix(width, height int, opts ...ResampleOption) (result string) {
	method, _, format := ResampleOptions(opts...)

	result = fmt.Sprintf("%dx%d_%s.%s", width, height, ResampleMethods[method], format)

	return result
}

// FileName returns the file name of the thumbnail for the matching size.
func FileName(hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if InvalidSize(width) {
		return "", fmt.Errorf("thumb: width exceeds limit (%d)", width)
	}

	if InvalidSize(height) {
		return "", fmt.Errorf("thumb: height exceeds limit (%d)", height)
	}

	if len(hash) < 4 {
		return "", fmt.Errorf("thumb: file hash is empty or too short (%s)", clean.Log(hash))
	}

	if len(thumbPath) == 0 {
		return "", errors.New("thumb: folder is empty")
	}

	suffix := Suffix(width, height, opts...)
	p := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])

	if err = fs.MkdirAll(p); err != nil {
		return "", err
	}

	fileName = fmt.Sprintf("%s/%s_%s", p, hash, suffix)

	return fileName, nil
}

// ResolvedName returns the file name of the thumbnail for the matching size with all symlinks resolved.
func ResolvedName(hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if fileName, err = FileName(hash, thumbPath, width, height, opts...); err != nil {
		return fileName, err
	} else {
		return fs.Resolve(fileName)
	}
}

// FromCache returns the filename if a thumbnail image with the matching size is in the cache.
func FromCache(imageFilename, hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("thumb: invalid file hash %s", clean.Log(hash))
	}

	if len(imageFilename) < 4 {
		return "", fmt.Errorf("thumb: invalid file name %s", clean.Log(imageFilename))
	}

	if fileName, err = FileName(hash, thumbPath, width, height, opts...); err != nil {
		log.Debugf("thumb: %s in %s (get filename)", err, clean.Log(imageFilename))
		return "", err
	} else if fileName, err = fs.Resolve(fileName); err != nil {
		return "", ErrNotCached
	} else if fs.FileExistsNotEmpty(fileName) {
		return fileName, nil
	}

	return "", ErrNotCached
}

// FromFile creates a new thumbnail with the specified size if it was not found in the cache, and returns the filename.
func FromFile(imageFilename, hash, thumbPath string, width, height, orientation int, opts ...ResampleOption) (fileName string, err error) {
	if fileName, err = FromCache(imageFilename, hash, thumbPath, width, height, opts...); err == nil {
		return fileName, err
	} else if !errors.Is(err, ErrNotCached) {
		return "", err
	}

	// Generate thumb cache filename.
	fileName, err = FileName(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Error(err)
		return "", err
	}

	// Load image from file.
	img, err := Open(imageFilename, orientation)

	if err != nil {
		log.Debugf("thumb: %s in %s", err, clean.Log(filepath.Base(imageFilename)))
		return "", err
	}

	// Create thumb from image.
	if _, err = Create(img, fileName, width, height, opts...); err != nil {
		return "", err
	}

	return fileName, nil
}

// Create creates an image thumbnail.
func Create(img image.Image, fileName string, width, height int, opts ...ResampleOption) (result image.Image, err error) {
	if InvalidSize(width) {
		return img, fmt.Errorf("thumb: width has an invalid value (%d)", width)
	}

	if InvalidSize(height) {
		return img, fmt.Errorf("thumb: height has an invalid value (%d)", height)
	}

	result = Resample(img, width, height, opts...)

	var quality imaging.EncodeOption

	if fs.FileType(fileName) == fs.ImagePNG {
		quality = imaging.PNGCompressionLevel(png.DefaultCompression)
	} else if width <= 150 && height <= 150 {
		quality = JpegQualitySmall.EncodeOption()
	} else {
		quality = JpegQuality.EncodeOption()
	}

	err = imaging.Save(result, fileName, quality)

	if err != nil {
		log.Debugf("thumb: failed to save %s", clean.Log(filepath.Base(fileName)))
		return result, err
	}

	return result, nil
}
