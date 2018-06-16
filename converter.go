package photoprism

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Converter struct {
	darktableCli string
}

func NewConverter(darktableCli string) *Converter {
	if stat, err := os.Stat(darktableCli); err != nil {
		log.Print("Darktable CLI binary could not be found at " + darktableCli)
	} else if stat.IsDir() {
		log.Print("Darktable CLI must be a file, not a directory")
	}

	return &Converter{darktableCli: darktableCli}
}

func (c *Converter) ConvertAll(path string) {
	err := filepath.Walk(path, func(filename string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			log.Print(err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mediaFile := NewMediaFile(filename)

		if !mediaFile.Exists() || !mediaFile.IsRaw() {
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

func (c *Converter) ConvertToJpeg(image *MediaFile) (*MediaFile, error) {
	if !image.Exists() {
		return nil, errors.New("can not convert, file does not exist")
	}

	if image.IsJpeg() {
		return image, nil
	}

	extension := image.GetExtension()

	baseFilename := image.filename[0 : len(image.filename)-len(extension)]

	jpegFilename := baseFilename + ".jpg"

	if _, err := os.Stat(jpegFilename); err == nil {
		return NewMediaFile(jpegFilename), nil
	}

	log.Printf("Converting %s to %s \n", image.filename, jpegFilename)

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

	return NewMediaFile(jpegFilename), nil
}
