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

		mediaFile := NewMediaFile(filename)

		if !mediaFile.Exists() || !mediaFile.IsJpeg() {
			return nil
		}

		if square {
			log.Printf("Creating square %dpx thumbnail for %s", size, filename)

			if _, err := mediaFile.GetSquareThumbnail(thumbnailsPath, size); err != nil {
				log.Print(err.Error())
			}
		} else {
			log.Printf("Creating %dpx thumbnail for %s", size, filename)

			if _, err := mediaFile.GetThumbnail(thumbnailsPath, size); err != nil {
				log.Print(err.Error())
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
		return NewMediaFile(thumbnailFilename), nil
	}

	return m.CreateThumbnail(thumbnailFilename, size)
}

// Resize preserving the aspect ratio
func (m *MediaFile) CreateThumbnail(filename string, size int) (result *MediaFile, err error) {
	image, err := imaging.Open(m.filename)

	if err != nil {
		log.Printf("open failed: %s", err.Error())
		return nil, err
	}

	image = imaging.Fit(image, size, size, imaging.Lanczos)

	err = imaging.Save(image, filename)

	if err != nil {
		log.Fatalf("failed to save image: %v", err)
		return nil, err
	}

	result = NewMediaFile(filename)

	return result, nil
}

func (m *MediaFile) GetSquareThumbnail(path string, size int) (result *MediaFile, err error) {
	canonicalName := m.GetCanonicalName()
	dateCreated := m.GetDateCreated()

	thumbnailPath := fmt.Sprintf("%s/square/%dpx/%s", path, size, dateCreated.UTC().Format("2006/01"))

	os.MkdirAll(thumbnailPath, os.ModePerm)

	thumbnailFilename := fmt.Sprintf("%s/%s_square_%dpx.jpg", thumbnailPath, canonicalName, size)

	if fileExists(thumbnailFilename) {
		return NewMediaFile(thumbnailFilename), nil
	}

	return m.CreateSquareThumbnail(thumbnailFilename, size)
}

// Resize and crop to square format
func (m *MediaFile) CreateSquareThumbnail(filename string, size int) (result *MediaFile, err error) {
	image, err := imaging.Open(m.filename)

	if err != nil {
		log.Printf("open failed: %s", err.Error())
		return nil, err
	}

	image = imaging.Fill(image, size, size, imaging.Center, imaging.Lanczos)

	err = imaging.Save(image, filename)

	if err != nil {
		log.Fatalf("failed to save image: %v", err)
		return nil, err
	}

	result = NewMediaFile(filename)

	return result, nil
}
