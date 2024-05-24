package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FixJpeg tries to re-encode a broken JPEG and returns the cached image file.
func (c *Convert) FixJpeg(f *MediaFile, force bool) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - you may have found a bug")
	}

	logName := clean.Log(f.RootRelName())

	if c.conf.DisableImageMagick() || !c.imagemagickSkip.Allow(fs.ExtJPEG) {
		return nil, fmt.Errorf("convert: ImageMagick must be enabled to re-encode %s", logName)
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", logName)
	} else if f.Empty() {
		return nil, fmt.Errorf("convert: %s is empty", logName)
	} else if !f.IsJpeg() {
		return nil, fmt.Errorf("convert: %s is not a jpeg", logName)
	}

	var err error

	// Get SHA1 file hash.
	fileHash := f.Hash()

	// Get cache path based on config and file hash.
	cacheDir := c.conf.MediaFileCachePath(fileHash)

	// Compose cache filename.
	cacheName := filepath.Join(cacheDir, fileHash+fs.ExtJPEG)

	mediaFile, err := NewMediaFile(cacheName)

	// Replace existing sidecar if "force" is true.
	if err == nil && mediaFile.IsJpeg() {
		if force && mediaFile.InSidecar() {
			if err := mediaFile.Remove(); err != nil {
				return mediaFile, fmt.Errorf("convert: failed removing %s (%s)", clean.Log(mediaFile.RootRelName()), err)
			} else {
				log.Infof("convert: replacing %s", clean.Log(mediaFile.RootRelName()))
			}
		} else {
			return mediaFile, nil
		}
	}

	fileName := f.RelName(c.conf.OriginalsPath())

	// Publish file conversion event.
	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  "",
	})

	start := time.Now()

	// Try ImageMagick for other image file formats if allowed.
	quality := fmt.Sprintf("%d", c.conf.JpegQuality())
	resize := fmt.Sprintf("%dx%d>", c.conf.JpegSize(), c.conf.JpegSize())
	args := []string{f.FileName(), "-flatten", "-resize", resize, "-quality", quality, cacheName}
	cmd := exec.Command(c.conf.ImageMagickBin(), args...)

	if fs.FileExists(cacheName) {
		return NewMediaFile(cacheName)
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{
		fmt.Sprintf("HOME=%s", c.conf.CmdCachePath()),
		fmt.Sprintf("LD_LIBRARY_PATH=%s", c.conf.CmdLibPath()),
	}

	log.Infof("convert: re-encoding %s to %s (%s)", logName, clean.Log(filepath.Base(cacheName)), filepath.Base(cmd.Path))

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err = cmd.Run(); err != nil {
		if stderr.String() != "" {
			err = errors.New(stderr.String())
		}

		log.Tracef("convert: %s (%s)", err, filepath.Base(cmd.Path))
	} else if fs.FileExistsNotEmpty(cacheName) {
		log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(cacheName)), time.Since(start), filepath.Base(cmd.Path))
	}

	// Ok?
	if err != nil {
		return nil, err
	}

	return NewMediaFile(cacheName)
}
