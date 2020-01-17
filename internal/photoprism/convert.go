package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
)

// Convert represents a converter that can convert RAW/HEIF images to JPEG.
type Convert struct {
	conf     *config.Config
	cmdMutex sync.Mutex
}

// NewConvert returns a new converter and expects the config as argument.
func NewConvert(conf *config.Config) *Convert {
	return &Convert{conf: conf}
}

// Start converts all files in a directory to JPEG if possible.
func (c *Convert) Start(path string) error {
	if err := mutex.Worker.Start(); err != nil {
		return err
	}

	defer mutex.Worker.Stop()

	err := filepath.Walk(path, func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("convert: %s [panic]", err)
			}
		}()

		if mutex.Worker.Canceled() {
			return errors.New("convert: canceled")
		}

		if err != nil {
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mf, err := NewMediaFile(filename)

		if err != nil || !(mf.IsRaw() || mf.IsHEIF() || mf.IsImageOther()) {
			return nil
		}

		if _, err := c.ToJpeg(mf); err != nil {
			log.Warnf("convert: %s (%s)", err.Error(), filename)
		}

		return nil
	})

	return err
}

// ConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) ConvertCommand(image *MediaFile, jpegFilename string, xmpFilename string) (result *exec.Cmd, err error) {
	if image.IsRaw() {
		if c.conf.SipsBin() != "" {
			result = exec.Command(c.conf.SipsBin(), "-s format jpeg", image.filename, "--out "+jpegFilename)
		} else if c.conf.DarktableBin() != "" {
			if xmpFilename != "" {
				result = exec.Command(c.conf.DarktableBin(), image.filename, xmpFilename, jpegFilename)
			} else {
				result = exec.Command(c.conf.DarktableBin(), image.filename, jpegFilename)
			}
		} else {
			return nil, fmt.Errorf("convert: no binary for raw to jpeg could be found (%s)", image.Filename())
		}
	} else if image.IsHEIF() {
		result = exec.Command(c.conf.HeifConvertBin(), image.filename, jpegFilename)
	} else {
		return nil, fmt.Errorf("convert: image type not supported for conversion (%s)", image.Type())
	}

	return result, nil
}

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(image *MediaFile) (*MediaFile, error) {
	if !image.Exists() {
		return nil, fmt.Errorf("convert: can not convert to jpeg, file does not exist (%s)", image.Filename())
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
		return nil, fmt.Errorf("convert: disabled in read only mode (%s)", image.Filename())
	}

	log.Infof("convert: \"%s\" -> \"%s\"", image.filename, jpegFilename)

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

	cmd, err := c.ConvertCommand(image, jpegFilename, xmpFilename)

	if err != nil {
		return nil, err
	}

	// Unclear if this is really necessary here, but safe is safe.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Make sure only one command is executed at a time.
	// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
	c.cmdMutex.Lock()
	defer c.cmdMutex.Unlock()

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Run convert command.
	if err := cmd.Run(); err != nil {
		return nil, errors.New(stderr.String())
	}

	return NewMediaFile(jpegFilename)
}
