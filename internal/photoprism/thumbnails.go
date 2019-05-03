package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/fsutil"
)

// CreateThumbnailsFromOriginals create thumbnails.
func CreateThumbnailsFromOriginals(originalsPath string, thumbnailsPath string, size int, square bool) {
	err := filepath.Walk(originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsJpeg() {
			return nil
		}

		if square {
			if thumbnail, err := mediaFile.SquareThumbnail(thumbnailsPath, size); err != nil {
				log.Errorf("could not create thumbnail: %s", err.Error())
			} else {
				log.Infof("created %dx%d thumbnail for \"%s\"", thumbnail.Width(), thumbnail.Height(), mediaFile.RelativeFilename(originalsPath))
			}
		} else {
			if thumbnail, err := mediaFile.Thumbnail(thumbnailsPath, size); err != nil {
				log.Errorf("could not create thumbnail: %s", err.Error())
			} else {
				log.Infof("created %dx%d thumbnail for \"%s\"", thumbnail.Width(), thumbnail.Height(), mediaFile.RelativeFilename(originalsPath))
			}
		}

		return nil
	})

	if err != nil {
		log.Error(err.Error())
	}
}

// Thumbnail get the thumbnail for a path.
func (m *MediaFile) Thumbnail(path string, size int) (result *MediaFile, err error) {
	canonicalName := m.CanonicalName()
	dateCreated := m.DateCreated()

	thumbnailPath := fmt.Sprintf("%s/%dpx/%s", path, size, dateCreated.UTC().Format("2006/01"))

	os.MkdirAll(thumbnailPath, os.ModePerm)

	thumbnailFilename := fmt.Sprintf("%s/%s_%dpx.jpg", thumbnailPath, canonicalName, size)

	if fsutil.Exists(thumbnailFilename) {
		return NewMediaFile(thumbnailFilename)
	}

	return m.CreateThumbnail(thumbnailFilename, size)
}

// CreateThumbnail Resize preserving the aspect ratio
func (m *MediaFile) CreateThumbnail(filename string, size int) (result *MediaFile, err error) {
	img, err := imaging.Open(m.filename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("can't open original: %s", err.Error())
		return nil, err
	}

	img = imaging.Fit(img, size, size, imaging.Lanczos)

	err = imaging.Save(img, filename)

	if err != nil {
		log.Fatalf("failed to save thumbnail: %v", err)
		return nil, err
	}

	return NewMediaFile(filename)
}

// SquareThumbnail return the square thumbnail for a path and size.
func (m *MediaFile) SquareThumbnail(path string, size int) (result *MediaFile, err error) {
	canonicalName := m.CanonicalName()
	dateCreated := m.DateCreated()

	thumbnailPath := fmt.Sprintf("%s/square/%dpx/%s", path, size, dateCreated.UTC().Format("2006/01"))

	os.MkdirAll(thumbnailPath, os.ModePerm)

	thumbnailFilename := fmt.Sprintf("%s/%s_square_%dpx.jpg", thumbnailPath, canonicalName, size)

	if fsutil.Exists(thumbnailFilename) {
		return NewMediaFile(thumbnailFilename)
	}

	return m.CreateSquareThumbnail(thumbnailFilename, size)
}

// CreateSquareThumbnail Resize and crop to square format
func (m *MediaFile) CreateSquareThumbnail(filename string, size int) (result *MediaFile, err error) {
	img, err := imaging.Open(m.filename, imaging.AutoOrientation(true))

	if err != nil {
		log.Errorf("can't open original: %s", err.Error())
		return nil, err
	}

	img = imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)

	err = imaging.Save(img, filename)

	if err != nil {
		log.Fatalf("failed to save thumbnail: %v", err)
		return nil, err
	}

	return NewMediaFile(filename)
}
