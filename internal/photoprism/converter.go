package photoprism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/config"
)

// Converter wraps a darktable cli binary.
type Converter struct {
	conf *config.Config
}

// NewConverter returns a new converter by setting the darktable
// cli binary location.
func NewConverter(conf *config.Config) *Converter {
	return &Converter{conf: conf}
}

// ConvertAll converts all the files given a path to JPEG. This function
// ignores error during this process.
func (c *Converter) ConvertAll(path string) {
	err := filepath.Walk(path, func(filename string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			log.Error("Walk", err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !(mediaFile.IsRaw() || mediaFile.IsHEIF()) {
			return nil
		}

		if _, err := c.ConvertToJpeg(mediaFile); err != nil {
			log.Warnf("file could not be converted to JPEG: \"%s\"", filename)
		}

		return nil
	})

	if err != nil {
		log.Error(err.Error())
	}
}

func (c *Converter) ConvertCommand(image *MediaFile, jpegFilename string, xmpFilename string) (result *exec.Cmd, err error) {
	if image.IsRaw() {
		if c.conf.SipsBin() != "" {
			result = exec.Command(c.conf.SipsBin(), "-s format jpeg", image.filename, "--out "+jpegFilename)
		} else if c.conf.DarktableBin() != "" && xmpFilename != "" {
			result = exec.Command(c.conf.DarktableBin(), image.filename, xmpFilename, jpegFilename)
		} else if c.conf.DarktableBin() != "" {
			result = exec.Command(c.conf.DarktableBin(), image.filename, jpegFilename)
		} else {
			return nil, fmt.Errorf("no binary for raw to jpeg conversion could be found: %s", image.Filename())
		}
	} else if image.IsHEIF() {
		result = exec.Command(c.conf.HeifConvertBin(), image.filename, jpegFilename)
	} else {
		return nil, fmt.Errorf("image type not supported for conversion: %s", image.Type())
	}

	return result, nil
}

// ConvertToJpeg converts a single image the JPEG format.
func (c *Converter) ConvertToJpeg(image *MediaFile) (*MediaFile, error) {
	if !image.Exists() {
		return nil, fmt.Errorf("can not convert to jpeg, file does not exist: %s", image.Filename())
	}

	if image.IsJpeg() {
		return image, nil
	}

	baseFilename := image.DirectoryBasename()

	jpegFilename := baseFilename + ".jpg"

	mediaFile, err := NewMediaFile(jpegFilename)

	if err == nil {
		return mediaFile, nil
	}

	if c.conf.ReadOnly() {
		return nil, fmt.Errorf("can not convert to jpeg in read only mode: %s", image.Filename())
	}

	log.Infof("converting \"%s\" to \"%s\"", image.filename, jpegFilename)

	xmpFilename := baseFilename + ".xmp"

	if _, err := os.Stat(xmpFilename); err != nil {
		xmpFilename = ""
	}

	if convertCommand, err := c.ConvertCommand(image, jpegFilename, xmpFilename); err != nil {
		return nil, err
	} else if err := convertCommand.Run(); err != nil {
		return nil, err
	}

	return NewMediaFile(jpegFilename)
}
