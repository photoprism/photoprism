package thumb

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/disintegration/imaging"
)

func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter imaging.ResampleFilter, format fs.FileFormat) {
	method = ResampleFit
	filter = imaging.Lanczos
	format = fs.FormatJpeg

	for _, option := range opts {
		switch option {
		case ResamplePng:
			format = fs.FormatPng
		case ResampleNearestNeighbor:
			filter = imaging.NearestNeighbor
		case ResampleDefault:
			filter = Filter.Imaging()
		case ResampleFillTopLeft:
			method = ResampleFillTopLeft
		case ResampleFillCenter:
			method = ResampleFillCenter
		case ResampleFillBottomRight:
			method = ResampleFillBottomRight
		case ResampleFit:
			method = ResampleFit
		case ResampleResize:
			method = ResampleResize
		}
	}

	return method, filter, format
}

func Resample(img image.Image, width, height int, opts ...ResampleOption) image.Image {
	var resImg image.Image

	method, filter, _ := ResampleOptions(opts...)

	if method == ResampleFit {
		resImg = imaging.Fit(img, width, height, filter)
	} else if method == ResampleFillCenter {
		resImg = imaging.Fill(img, width, height, imaging.Center, filter)
	} else if method == ResampleFillTopLeft {
		resImg = imaging.Fill(img, width, height, imaging.TopLeft, filter)
	} else if method == ResampleFillBottomRight {
		resImg = imaging.Fill(img, width, height, imaging.BottomRight, filter)
	} else if method == ResampleResize {
		resImg = imaging.Resize(img, width, height, filter)
	}

	return resImg
}

func Postfix(width, height int, opts ...ResampleOption) (result string) {
	method, _, format := ResampleOptions(opts...)

	result = fmt.Sprintf("%dx%d_%s.%s", width, height, ResampleMethods[method], format)

	return result
}

func Filename(hash string, thumbPath string, width, height int, opts ...ResampleOption) (filename string, err error) {
	if InvalidSize(width) {
		return "", fmt.Errorf("resample: width exceeds limit (%d)", width)
	}

	if InvalidSize(height) {
		return "", fmt.Errorf("resample: height exceeds limit (%d)", height)
	}

	if len(hash) < 4 {
		return "", fmt.Errorf("resample: file hash is empty or too short (%s)", txt.Quote(hash))
	}

	if len(thumbPath) == 0 {
		return "", errors.New("resample: folder is empty")
	}

	postfix := Postfix(width, height, opts...)
	p := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])

	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return "", err
	}

	filename = fmt.Sprintf("%s/%s_%s", p, hash, postfix)

	return filename, nil
}

func FromCache(imageFilename, hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("resample: file hash is empty or too short (%s)", txt.Quote(hash))
	}

	if len(imageFilename) < 4 {
		return "", fmt.Errorf("resample: image filename is empty or too short (%s)", txt.Quote(imageFilename))
	}

	fileName, err = Filename(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Error(err)
		return "", err
	}

	if fs.FileExists(fileName) {
		return fileName, nil
	}

	return "", ErrThumbNotCached
}

func FromFile(imageFilename, hash, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if fileName, err := FromCache(imageFilename, hash, thumbPath, width, height, opts...); err == nil {
		return fileName, err
	} else if err != ErrThumbNotCached {
		return "", err
	}

	fileName, err = Filename(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Error(err)
		return "", err
	}

	img, err := imaging.Open(imageFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("resample: %s in %s", err, txt.Quote(filepath.Base(imageFilename)))
		return "", err
	}

	if _, err := Create(img, fileName, width, height, opts...); err != nil {
		return "", err
	}

	return fileName, nil
}

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
		log.Errorf("resample: failed to save %s", txt.Quote(filepath.Base(fileName)))
		return result, err
	}

	return result, nil
}
