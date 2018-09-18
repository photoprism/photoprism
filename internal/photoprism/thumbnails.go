package photoprism

import (
	"fmt"
	"github.com/disintegration/imaging"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
			if thumbnail, err := mediaFile.GetSquareThumbnail(thumbnailsPath, size); err != nil {
				log.Printf("Could not create thumbnail: %s", err.Error())
			} else {
				log.Printf("Created %dx%d px thumbnail for \"%s\"", thumbnail.GetWidth(), thumbnail.GetHeight(), mediaFile.GetRelativeFilename(originalsPath))
			}
		} else {
			if thumbnail, err := mediaFile.GetThumbnail(thumbnailsPath, size); err != nil {
				log.Printf("Could not create thumbnail: %s", err.Error())
			} else {
				log.Printf("Created %dx%d px thumbnail for \"%s\"", thumbnail.GetWidth(), thumbnail.GetHeight(), mediaFile.GetRelativeFilename(originalsPath))
			}
		}

		return nil
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func (m *MediaFile) GetThumbnail(path string, size int) (result *MediaFile, err error) {
	canonicalName := m.GetCanonicalName()
	dateCreated := m.GetDateCreated()

	thumbnailPath := fmt.Sprintf("%s/%dpx/%s", path, size, dateCreated.UTC().Format("2006/01"))

	os.MkdirAll(thumbnailPath, os.ModePerm)

	thumbnailFilename := fmt.Sprintf("%s/%s_%dpx.jpg", thumbnailPath, canonicalName, size)

	if fileExists(thumbnailFilename) {
		return NewMediaFile(thumbnailFilename)
	}

	return m.CreateThumbnail(thumbnailFilename, size)
}

// Resize preserving the aspect ratio
func (m *MediaFile) CreateThumbnail(filename string, size int) (result *MediaFile, err error) {
	img, err := imaging.Open(m.filename, imaging.AutoOrientation(true))

	if err != nil {
		log.Printf("open failed: %s", err.Error())
		return nil, err
	}

	img = imaging.Fit(img, size, size, imaging.Lanczos)

	err = imaging.Save(img, filename)

	if err != nil {
		log.Fatalf("failed to save img: %v", err)
		return nil, err
	}

	return NewMediaFile(filename)
}

func (m *MediaFile) GetSquareThumbnail(path string, size int) (result *MediaFile, err error) {
	canonicalName := m.GetCanonicalName()
	dateCreated := m.GetDateCreated()

	thumbnailPath := fmt.Sprintf("%s/square/%dpx/%s", path, size, dateCreated.UTC().Format("2006/01"))

	os.MkdirAll(thumbnailPath, os.ModePerm)

	thumbnailFilename := fmt.Sprintf("%s/%s_square_%dpx.jpg", thumbnailPath, canonicalName, size)

	if fileExists(thumbnailFilename) {
		return NewMediaFile(thumbnailFilename)
	}

	return m.CreateSquareThumbnail(thumbnailFilename, size)
}

// Resize and crop to square format
func (m *MediaFile) CreateSquareThumbnail(filename string, size int) (result *MediaFile, err error) {
	img, err := imaging.Open(m.filename, imaging.AutoOrientation(true))

	if err != nil {
		log.Printf("open failed: %s", err.Error())
		return nil, err
	}

	img = imaging.Fill(img, size, size, imaging.Center, imaging.Lanczos)

	err = imaging.Save(img, filename)

	if err != nil {
		log.Fatalf("failed to save img: %v", err)
		return nil, err
	}

	return NewMediaFile(filename)
}
