package photoprism

import (
	"os"
	"os/exec"
	"log"
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

func (converter *Converter) ConvertToJpeg(image *MediaFile) (*MediaFile, error) {
	if image.IsJpeg() {
		return image, nil
	}

	extension := image.GetExtension()

	baseFilename := image.filename[0:len(image.filename)-len(extension)]

	jpegFilename := baseFilename + ".jpg"

	if _, err := os.Stat(jpegFilename); err == nil {
		return NewMediaFile(jpegFilename), nil
	}

	xmpFilename := baseFilename + ".xmp"

	var convertCommand *exec.Cmd

	if _, err := os.Stat(xmpFilename); err == nil {
		convertCommand = exec.Command(converter.darktableCli, image.filename, xmpFilename, jpegFilename)
	} else {
		convertCommand = exec.Command(converter.darktableCli, image.filename, jpegFilename)
	}

	if err := convertCommand.Run(); err != nil {
		return nil, err
	}

	return NewMediaFile(jpegFilename), nil
}
