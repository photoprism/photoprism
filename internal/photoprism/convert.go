package photoprism

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/thumb"
)

// Convert represents a converter that can convert RAW/HEIF images to JPEG.
type Convert struct {
	conf *config.Config
}

// NewConvert returns a new converter and expects the config as argument.
func NewConvert(conf *config.Config) *Convert {
	return &Convert{conf: conf}
}

// Path converts all files in a directory to JPEG if possible.
func (c *Convert) Path(path string) {
	err := filepath.Walk(path, func(filename string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			log.Error("Walk", err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !(mediaFile.IsRaw() || mediaFile.IsHEIF() || mediaFile.IsImageOther()) {
			return nil
		}

		if _, err := c.ToJpeg(mediaFile); err != nil {
			log.Warnf("file could not be converted to JPEG: \"%s\"", filename)
		}

		return nil
	})

	if err != nil {
		log.Error(err.Error())
	}
}

// ConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) ConvertCommand(image *MediaFile, jpegFilename string, xmpFilename string) (result *exec.Cmd, err error) {
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

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(image *MediaFile) (*MediaFile, error) {
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

	fileName := image.RelativeFilename(c.conf.OriginalsPath())

	xmpFilename := baseFilename + ".xmp"

	if _, err := os.Stat(xmpFilename); err != nil {
		xmpFilename = ""
	}

	event.Publish("index.converting", event.Data{
		"fileType": image.Type(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpFilename),
	})

	if image.IsImageOther() {
		_, err = thumb.Jpeg(image.Filename(), jpegFilename)

		if err != nil {
			return nil, err
		}

		return NewMediaFile(jpegFilename)
	}

	if convertCommand, err := c.ConvertCommand(image, jpegFilename, xmpFilename); err != nil {
		return nil, err
	} else if err := convertCommand.Run(); err != nil {
		return nil, err
	}

	return NewMediaFile(jpegFilename)
}
