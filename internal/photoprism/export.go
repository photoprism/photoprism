package photoprism

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FindOriginalsByDate searches the originalsPath given a time frame in the format of
// after <=> before and returns a list of results.
func FindOriginalsByDate(originalsPath string, after time.Time, before time.Time) (result []*MediaFile) {
	filepath.Walk(originalsPath, func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil || fileInfo.IsDir() || strings.HasPrefix(filepath.Base(filename), ".") {
			return nil
		}

		_, basename := filepath.Split(filename)

		if basename <= after.Format("20060102_150405") || basename >= before.Format("20060102_150405") {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)
		if err != nil || !mediaFile.IsJpeg() {
			return nil
		}

		result = append(result, mediaFile)

		return nil
	})

	return result
}

// ExportPhotosFromOriginals takes a list of original mediafiles and exports
// them to JPEG.
func ExportPhotosFromOriginals(originals []*MediaFile, thumbnailsPath string, exportPath string, size int) (err error) {
	for _, mediaFile := range originals {

		if !mediaFile.Exists() || !mediaFile.IsJpeg() {
			return nil
		}

		log.Printf("Exporting %s as %dpx JPEG", mediaFile.GetFilename(), size)

		thumbnail, err := mediaFile.GetThumbnail(thumbnailsPath, size)

		if err != nil {
			log.Print(err.Error())
		}

		os.MkdirAll(exportPath, os.ModePerm)

		destinationFilename := fmt.Sprintf("%s/%s_%dpx.jpg", exportPath, mediaFile.GetCanonicalName(), size)

		thumbnail.Copy(destinationFilename)
	}

	return nil
}
