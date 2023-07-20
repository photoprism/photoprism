package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ToImage converts a media file to a directly supported image file format.
func (c *Convert) ToImage(f *MediaFile, force bool) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you may have found a bug")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", clean.Log(f.RootRelName()))
	} else if f.Empty() {
		return nil, fmt.Errorf("convert: %s is empty", clean.Log(f.RootRelName()))
	}

	if f.IsPreviewImage() {
		return f, nil
	}

	var err error

	imageName := fs.ImagePNG.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	if imageName == "" {
		imageName = fs.ImageJPEG.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)
	}

	mediaFile, err := NewMediaFile(imageName)

	// Replace existing sidecar if "force" is true.
	if err == nil && mediaFile.IsPreviewImage() {
		if force && mediaFile.InSidecar() {
			if err := mediaFile.Remove(); err != nil {
				return mediaFile, fmt.Errorf("convert: failed removing %s (%s)", clean.Log(mediaFile.RootRelName()), err)
			} else {
				log.Infof("convert: replacing %s", clean.Log(mediaFile.RootRelName()))
			}
		} else {
			return mediaFile, nil
		}
	} else if f.IsVector() {
		if !c.conf.VectorEnabled() {
			return nil, fmt.Errorf("convert: vector graphics support disabled (%s)", clean.Log(f.RootRelName()))
		}
		imageName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.ExtPNG)
	} else {
		imageName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.ExtJPEG)
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read-only mode (%s)", clean.Log(f.RootRelName()))
	}

	fileName := f.RelName(c.conf.OriginalsPath())
	xmpName := fs.SidecarXMP.Find(f.FileName(), false)

	// Publish file conversion event.
	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	start := time.Now()

	// PNG, GIF, BMP, TIFF, and WebP can be handled natively.
	if f.IsImageOther() {
		log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(imageName)), f.FileType())

		// Create PNG or JPEG image from source file.
		switch fs.LowerExt(imageName) {
		case fs.ExtPNG:
			_, err = thumb.Png(f.FileName(), imageName, f.Orientation())
		case fs.ExtJPEG:
			_, err = thumb.Jpeg(f.FileName(), imageName, f.Orientation())
		default:
			return nil, fmt.Errorf("convert: unspported target format %s (%s)", fs.LowerExt(imageName), clean.Log(f.RootRelName()))
		}

		// Check result.
		if err == nil {
			log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(imageName)), time.Since(start), f.FileType())
			return NewMediaFile(imageName)
		} else if !f.IsTIFF() && !f.IsWebP() {
			// See https://github.com/photoprism/photoprism/issues/1612
			// for TIFF file format compatibility.
			return nil, err
		}
	}

	// Run external commands for other formats.
	var cmds []*exec.Cmd
	var useMutex bool
	var expectedMime string

	switch fs.LowerExt(imageName) {
	case fs.ExtPNG:
		cmds, useMutex, err = c.PngConvertCommands(f, imageName)
		expectedMime = fs.MimeTypePNG
	case fs.ExtJPEG:
		cmds, useMutex, err = c.JpegConvertCommands(f, imageName, xmpName)
		expectedMime = fs.MimeTypeJPEG
	default:
		return nil, fmt.Errorf("convert: unspported target format %s (%s)", fs.LowerExt(imageName), clean.Log(f.RootRelName()))
	}

	if err != nil {
		return nil, err
	} else if len(cmds) == 0 {
		return nil, fmt.Errorf("file type %s not supported", f.FileType())
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	if fs.FileExists(imageName) {
		return NewMediaFile(imageName)
	}

	// Try compatible converters.
	for _, cmd := range cmds {
		// Fetch command output.
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		cmd.Env = []string{
			fmt.Sprintf("HOME=%s", c.conf.CmdCachePath()),
			fmt.Sprintf("LD_LIBRARY_PATH=%s", c.conf.CmdLibPath()),
		}

		log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(imageName)), filepath.Base(cmd.Path))

		// Log exact command for debugging in trace mode.
		log.Trace(cmd.String())

		// Run convert command.
		if err = cmd.Run(); err != nil {
			if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
				err = errors.New(errStr)
			}

			log.Tracef("convert: %s (%s)", strings.TrimSpace(err.Error()), filepath.Base(cmd.Path))
			continue
		} else if fs.FileExistsNotEmpty(imageName) {
			log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(imageName)), time.Since(start), filepath.Base(cmd.Path))
			break
		} else if res := out.Bytes(); len(res) < 512 || !mimetype.Detect(res).Is(expectedMime) {
			continue
		} else if err = os.WriteFile(imageName, res, fs.ModeFile); err != nil {
			log.Tracef("convert: %s (%s)", err, filepath.Base(cmd.Path))
			continue
		} else {
			break
		}
	}

	// Ok?
	if err != nil {
		return nil, err
	}

	return NewMediaFile(imageName)
}
