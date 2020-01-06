package thumb

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/file"

	"github.com/disintegration/imaging"
)

func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter imaging.ResampleFilter, format file.Type) {
	method = ResampleFit
	filter = imaging.Lanczos
	format = file.TypeJpeg

	for _, option := range opts {
		switch option {
		case ResamplePng:
			format = file.TypePng
		case ResampleNearestNeighbor:
			filter = imaging.NearestNeighbor
		case ResampleLanczos:
			filter = imaging.Lanczos
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
		default:
			panic(fmt.Errorf("not a valid resample option: %d", option))
		}
	}

	return method, filter, format
}

func Resample(img image.Image, width, height int, opts ...ResampleOption) (result image.Image) {
	method, filter, _ := ResampleOptions(opts...)

	if method == ResampleFit {
		result = imaging.Fit(img, width, height, filter)
	} else if method == ResampleFillCenter {
		result = imaging.Fill(img, width, height, imaging.Center, filter)
	} else if method == ResampleFillTopLeft {
		result = imaging.Fill(img, width, height, imaging.TopLeft, filter)
	} else if method == ResampleFillBottomRight {
		result = imaging.Fill(img, width, height, imaging.BottomRight, filter)
	} else if method == ResampleResize {
		result = imaging.Resize(img, width, height, filter)
	}

	return result
}

func Postfix(width, height int, opts ...ResampleOption) (result string) {
	method, _, format := ResampleOptions(opts...)

	result = fmt.Sprintf("%dx%d_%s.%s", width, height, ResampleMethods[method], format)

	return result
}

func Filename(hash string, thumbPath string, width, height int, opts ...ResampleOption) (filename string, err error) {
	if width < 0 || width > MaxWidth {
		return "", fmt.Errorf("width has an invalid value: %d", width)
	}

	if height < 0 || height > MaxHeight {
		return "", fmt.Errorf("height has an invalid value: %d", height)
	}

	if len(hash) < 4 {
		return "", fmt.Errorf("file hash is empty or too short: %s", hash)
	}

	if len(thumbPath) == 0 {
		return "", fmt.Errorf("thumbnail path is empty: %s", thumbPath)
	}

	postfix := Postfix(width, height, opts...)
	p := path.Join(thumbPath, hash[0:1], hash[1:2], hash[2:3])

	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		return "", err
	}

	filename = fmt.Sprintf("%s/%s_%s", p, hash, postfix)

	return filename, nil
}

func FromFile(imageFilename string, hash string, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("file hash is empty or too short: %s", hash)
	}

	if len(imageFilename) < 4 {
		return "", fmt.Errorf("image filename is empty or too short: %s", imageFilename)
	}

	fileName, err = Filename(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Errorf("can't determine thumb filename: %s", err)
		return "", err
	}

	if file.Exists(fileName) {
		return fileName, nil
	}

	img, err := imaging.Open(imageFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("can't open original: %s", err)
		return "", err
	}

	if _, err := Create(img, fileName, width, height, opts...); err != nil {
		return "", err
	}

	return fileName, nil
}

func Create(img image.Image, fileName string, width, height int, opts ...ResampleOption) (result image.Image, err error) {
	if width < 0 || width > MaxWidth {
		return img, fmt.Errorf("width has an invalid value: %d", width)
	}

	if height < 0 || height > MaxHeight {
		return img, fmt.Errorf("height has an invalid value: %d", height)
	}

	result = Resample(img, width, height, opts...)

	var saveOption imaging.EncodeOption

	if filepath.Ext(fileName) == "."+string(file.TypePng) {
		saveOption = imaging.PNGCompressionLevel(png.DefaultCompression)
	} else if width <= 150 && height <= 150 {
		saveOption = imaging.JPEGQuality(JpegQualitySmall)
	} else {
		saveOption = imaging.JPEGQuality(JpegQuality)
	}

	err = imaging.Save(result, fileName, saveOption)

	if err != nil {
		log.Errorf("failed to save thumbnail: %v", err)
		return result, err
	}

	return result, nil
}
