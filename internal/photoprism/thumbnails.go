package photoprism

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/capture"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/file"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/disintegration/imaging"
)

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

		fileName := mediaFile.RelativeFilename(originalsPath)

		event.Publish("index.thumbnails", event.Data{
			"fileName": fileName,
			"baseName": filepath.Base(fileName),
			"force":    force,
		})

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
	thumbType, ok := thumb.Types[typeName]

	if !ok {
		log.Errorf("invalid type: %s", typeName)
		return "", fmt.Errorf("invalid type: %s", typeName)
	}

	thumbnail, err := thumb.FromFile(m.Filename(), m.Hash(), path, thumbType.Width, thumbType.Height, thumbType.Options...)

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

func (m *MediaFile) CreateDefaultThumbnails(thumbPath string, force bool) (err error) {
	defer capture.Time(time.Now(), fmt.Sprintf("thumbs: creating thumbnails for \"%s\"", m.Filename()))

	hash := m.Hash()

	img, err := imaging.Open(m.Filename(), imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("thumbs: can't open original \"%s\"", err)
		return err
	}

	var sourceImg image.Image
	var sourceImgType string

	for _, name := range thumb.DefaultTypes {
		thumbType := thumb.Types[name]

		if thumbType.Height > thumb.MaxHeight || thumbType.Width > thumb.MaxWidth {
			log.Debugf("thumbs: size exceeds limit (width %d, height %d)", thumbType.Width, thumbType.Height)
			continue
		}

		if fileName, err := thumb.Filename(hash, thumbPath, thumbType.Width, thumbType.Height, thumbType.Options...); err != nil {
			log.Errorf("thumbs: could not create \"%s\" (%s)", name, err)

			return err
		} else {
			if !force && file.Exists(fileName) {
				continue
			}

			if thumbType.Source != "" {
				if thumbType.Source == sourceImgType && sourceImg != nil {
					_, err = thumb.Create(sourceImg, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				} else {
					_, err = thumb.Create(img, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				}
			} else {
				sourceImg, err = thumb.Create(img, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				sourceImgType = name
			}

			if err != nil {
				log.Errorf("thumbs: could not create \"%s\" (%s)", name, err)
				return err
			}
		}
	}

	return nil
}
