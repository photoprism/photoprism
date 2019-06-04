package photoprism

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/util"
)

const (
	MaxThumbWidth    = 8192
	MaxThumbHeight   = 8192
	JpegQuality      = 95
	JpegQualitySmall = 80
)

const (
	ResampleFillCenter ResampleOption = iota
	ResampleFillTopLeft
	ResampleFillBottomRight
	ResampleFit
	ResampleResize
	ResampleNearestNeighbor
	ResampleLanczos
	ResamplePng
)

type ResampleOption int

var ResampleMethods = map[ResampleOption]string{
	ResampleFillCenter:      "center",
	ResampleFillTopLeft:     "left",
	ResampleFillBottomRight: "right",
	ResampleFit:             "fit",
	ResampleResize:          "resize",
}

type ThumbnailType struct {
	Source  string
	Width   int
	Height  int
	Public  bool
	Options []ResampleOption
}

var ThumbnailTypes = map[string]ThumbnailType{
	"tile_50":   {"tile_500", 50, 50, false, []ResampleOption{ResampleFillCenter, ResampleLanczos}},
	"tile_100":  {"tile_500", 100, 100, false, []ResampleOption{ResampleFillCenter, ResampleLanczos}},
	"tile_224":  {"tile_500", 224, 224, false, []ResampleOption{ResampleFillCenter, ResampleLanczos}},
	"tile_500":  {"", 500, 500, false, []ResampleOption{ResampleFillCenter, ResampleLanczos}},
	"colors":    {"fit_720", 3, 3, false, []ResampleOption{ResampleResize, ResampleNearestNeighbor, ResamplePng}},
	"left_224":  {"fit_720", 224, 224, false, []ResampleOption{ResampleFillTopLeft, ResampleLanczos}},
	"right_224": {"fit_720", 224, 224, false, []ResampleOption{ResampleFillBottomRight, ResampleLanczos}},
	"fit_720":   {"", 720, 720, true, []ResampleOption{ResampleFit, ResampleLanczos}},
	"fit_1280":  {"fit_2048", 1280, 1024, true, []ResampleOption{ResampleFit, ResampleLanczos}},
	"fit_1920":  {"fit_2048", 1920, 1200, true, []ResampleOption{ResampleFit, ResampleLanczos}},
	"fit_2048":  {"", 2048, 2048, true, []ResampleOption{ResampleFit, ResampleLanczos}},
	"fit_2560":  {"", 2560, 1600, true, []ResampleOption{ResampleFit, ResampleLanczos}},
	"fit_3840":  {"", 3840, 2400, true, []ResampleOption{ResampleFit, ResampleLanczos}},
}

var DefaultThumbnails = []string{
	"fit_3840", "fit_2560", "fit_2048", "fit_1920", "fit_1280", "fit_720", "right_224", "left_224", "colors", "tile_500", "tile_224", "tile_100", "tile_50",
}

func init() {
	for name, t := range ThumbnailTypes {
		if t.Public {
			thumb := config.Thumbnail{Name: name, Width: t.Width, Height: t.Height}
			config.Thumbnails = append(config.Thumbnails, thumb)
		}
	}
}

// CreateThumbnailsFromOriginals creates default thumbnails for all originals.
func CreateThumbnailsFromOriginals(originalsPath string, thumbnailsPath string, force bool) error {
	err := filepath.Walk(originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsJpeg() {
			return nil
		}

		if err := mediaFile.CreateDefaultThumbnails(thumbnailsPath, force); err != nil {
			log.Errorf("could not create default thumbnails: %s", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Error(err)
	}

	return err
}

// Thumbnail returns a thumbnail filename.
func (m *MediaFile) Thumbnail(path string, typeName string) (filename string, err error) {
	thumbType, ok := ThumbnailTypes[typeName]

	if !ok {
		log.Errorf("invalid type: %s", typeName)
		return "", fmt.Errorf("invalid type: %s", typeName)
	}

	thumbnail, err := ThumbnailFromFile(m.Filename(), m.Hash(), path, thumbType.Width, thumbType.Height, thumbType.Options...)

	if err != nil {
		log.Errorf("could not create thumbnail: %s", err)
		return "", fmt.Errorf("could not create thumbnail: %s", err)
	}

	return thumbnail, nil
}

// Thumbnail returns a resampled image of the file.
func (m *MediaFile) Resample(path string, typeName string) (img image.Image, err error) {
	filename, err := m.Thumbnail(path, typeName)

	if err != nil {
		return nil, err
	}

	return imaging.Open(filename, imaging.AutoOrientation(true))
}

func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter imaging.ResampleFilter, format string) {
	method = ResampleFit
	filter = imaging.Lanczos
	format = FileTypeJpeg

	for _, option := range opts {
		switch option {
		case ResamplePng:
			format = FileTypePng
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

func ThumbnailPostfix(width, height int, opts ...ResampleOption) (result string) {
	method, _, format := ResampleOptions(opts...)

	result = fmt.Sprintf("%dx%d_%s.%s", width, height, ResampleMethods[method], format)

	return result
}

func ThumbnailFilename(hash string, thumbPath string, width, height int, opts ...ResampleOption) (filename string, err error) {
	if width < 0 || width > MaxThumbWidth {
		return "", fmt.Errorf("width has an invalid value: %d", width)
	}

	if height < 0 || height > MaxThumbHeight {
		return "", fmt.Errorf("height has an invalid value: %d", height)
	}

	if len(hash) < 4 {
		return "", fmt.Errorf("file hash is empty or too short: %s", hash)
	}

	if len(thumbPath) == 0 {
		return "", fmt.Errorf("thumbnail path is empty: %s", thumbPath)
	}

	postfix := ThumbnailPostfix(width, height, opts...)
	path := fmt.Sprintf("%s/%s/%s/%s", thumbPath, hash[0:1], hash[1:2], hash[2:3])

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}

	filename = fmt.Sprintf("%s/%s_%s", path, hash, postfix)

	return filename, nil
}

func ThumbnailFromFile(imageFilename string, hash string, thumbPath string, width, height int, opts ...ResampleOption) (fileName string, err error) {
	if len(hash) < 4 {
		return "", fmt.Errorf("file hash is empty or too short: %s", hash)
	}

	if len(imageFilename) < 4 {
		return "", fmt.Errorf("image filename is empty or too short: %s", imageFilename)
	}

	fileName, err = ThumbnailFilename(hash, thumbPath, width, height, opts...)

	if err != nil {
		log.Errorf("can't determine thumb filename: %s", err)
		return "", err
	}

	if util.Exists(fileName) {
		return fileName, nil
	}

	img, err := imaging.Open(imageFilename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("can't open original: %s", err)
		return "", err
	}

	if _, err := CreateThumbnail(img, fileName, width, height, opts...); err != nil {
		return "", err
	}

	return fileName, nil
}

func CreateThumbnail(img image.Image, fileName string, width, height int, opts ...ResampleOption) (result image.Image, err error) {
	if width < 0 || width > MaxThumbWidth {
		return img, fmt.Errorf("width has an invalid value: %d", width)
	}

	if height < 0 || height > MaxThumbHeight {
		return img, fmt.Errorf("height has an invalid value: %d", height)
	}

	result = Resample(img, width, height, opts...)

	var saveOption imaging.EncodeOption

	if filepath.Ext(fileName) == "."+FileTypePng {
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

func (m *MediaFile) CreateDefaultThumbnails(thumbPath string, force bool) (err error) {
	defer util.ProfileTime(time.Now(), fmt.Sprintf("creating thumbnails for \"%s\"", m.Filename()))

	hash := m.Hash()

	img, err := imaging.Open(m.Filename(), imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("can't open original: %s", err)
		return err
	}

	var sourceImg image.Image
	var sourceImgType string

	for _, name := range DefaultThumbnails {
		thumbType := ThumbnailTypes[name]

		if fileName, err := ThumbnailFilename(hash, thumbPath, thumbType.Width, thumbType.Height, thumbType.Options...); err != nil {
			log.Errorf("could not create %s thumbnail: \"%s\"", name, err)

			return err
		} else {
			if !force && util.Exists(fileName) {
				continue
			}

			if thumbType.Source != "" {
				if thumbType.Source == sourceImgType && sourceImg != nil {
					_, err = CreateThumbnail(sourceImg, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				} else {
					_, err = CreateThumbnail(img, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				}
			} else {
				sourceImg, err = CreateThumbnail(img, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				sourceImgType = name
			}

			if err != nil {
				log.Errorf("could not create %s thumbnail: \"%s\"", name, err)
				return err
			}
		}
	}

	return nil
}
