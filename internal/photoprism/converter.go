package photoprism

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Converter wraps a darktable cli binary.
type Converter struct {
	darktableCli string
}

// NewConverter returns a new converter by setting the darktable
// cli binary location.
func NewConverter(darktableCli string) *Converter {
	if stat, err := os.Stat(darktableCli); err != nil {
		log.Print("Darktable CLI binary could not be found at " + darktableCli)
	} else if stat.IsDir() {
		log.Print("Darktable CLI must be a file, not a directory")
	}

	return &Converter{darktableCli: darktableCli}
}

// ConvertAll converts all the files given a path to JPEG. This function
// ignores error during this process.
func (c *Converter) ConvertAll(path string) {
	err := filepath.Walk(path, func(filename string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			log.Print(err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsRaw() {
			return nil
		}

		if _, err := c.ConvertToJpeg(mediaFile); err != nil {
			log.Print(err.Error())
		}

		return nil
	})

	if err != nil {
		log.Print(err.Error())
	}
}

// ConvertToJpeg converts a single image the JPEG format.
func (c *Converter) ConvertToJpeg(image *MediaFile) (*MediaFile, error) {
	if !image.Exists() {
		return nil, fmt.Errorf("can not convert, file does not exist: %s", image.GetFilename())
	}

	if image.IsJpeg() {
		return image, nil
	}

	baseFilename := image.GetCanonicalNameFromFileWithDirectory()

	jpegFilename := baseFilename + ".jpg"

	mediaFile, err := NewMediaFile(jpegFilename)

	if err == nil {
		return mediaFile, nil
	}

	log.Printf("Converting \"%s\" to \"%s\"\n", image.filename, jpegFilename)

	xmpFilename := baseFilename + ".xmp"

	var convertCommand *exec.Cmd

	if _, err := os.Stat(xmpFilename); err == nil {
		convertCommand = exec.Command(c.darktableCli, image.filename, xmpFilename, jpegFilename)
	} else {
		convertCommand = exec.Command(c.darktableCli, image.filename, jpegFilename)
	}

	if err := convertCommand.Run(); err != nil {
		return nil, err
	}

	return NewMediaFile(jpegFilename)
}
