package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
)

// ToImage converts a media file to a directly supported image file format.
func (w *Convert) ToImage(f *MediaFile, force bool) (result *MediaFile, err error) {
	if f == nil {
		return nil, fmt.Errorf("convert: no media file provided for processing - you may have found a bug")
	}

	switch {
	case !f.Exists():
		return nil, fmt.Errorf("convert: %s not found", clean.Log(f.RootRelName()))
	case f.Empty():
		return nil, fmt.Errorf("convert: %s is empty", clean.Log(f.RootRelName()))
	case f.IsThumb():
		return nil, fmt.Errorf("convert: %s is a thumbnail image", clean.Log(f.RootRelName()))
	}

	if f.IsPreviewImage() {
		return f, nil
	}

	imageName := fs.ImagePng.FindFirst(f.FileName(), []string{w.conf.SidecarPath(), fs.PPHiddenPathname}, w.conf.OriginalsPath(), false)

	if imageName == "" {
		imageName = fs.ImageJpeg.FindFirst(f.FileName(), []string{w.conf.SidecarPath(), fs.PPHiddenPathname}, w.conf.OriginalsPath(), false)
	}

	mediaFile, err := NewMediaFile(imageName)

	// Replace existing sidecar if "force" is true.
	if err == nil && mediaFile.IsPreviewImage() {
		if force && mediaFile.InSidecar() {
			if removeErr := mediaFile.Remove(); removeErr != nil {
				return mediaFile, fmt.Errorf("convert: failed removing %s (%s)", clean.Log(mediaFile.RootRelName()), removeErr)
			}

			log.Infof("convert: replacing %s", clean.Log(mediaFile.RootRelName()))
		} else {
			return mediaFile, nil
		}
	} else {
		switch {
		case f.IsVector():
			if !w.conf.VectorEnabled() {
				return nil, fmt.Errorf("convert: vector graphics support disabled (%s)", clean.Log(f.RootRelName()))
			}
			imageName, _ = fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtPng)
		default:
			imageName, _ = fs.FileName(f.FileName(), w.conf.SidecarPath(), w.conf.OriginalsPath(), fs.ExtJpeg)
		}
	}

	// Refuse to write a new file if storage is read-only or insufficient.
	if !w.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read-only mode (%s)", clean.Log(f.RootRelName()))
	} else if w.conf.InsufficientStorage() {
		return nil, status.ErrInsufficientStorage
	}

	fileName := f.RelName(w.conf.OriginalsPath())
	fileOrientation := media.KeepOrientation
	xmpName := fs.SidecarXMP.Find(f.FileName(), false)

	// Publish file conversion event.
	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	start := time.Now()

	// PNG, GIF, BMP, TIFF, HEIC/HEIF, AVIF, and WebP are handled natively. JPEG XL only when
	// "djxl" is unavailable: some libvips builds report JPEG XL support but mis-render files
	// without erroring, so the reference "djxl" decoder is preferred when present.
	if f.IsImageOther() || (f.IsJpegXL() && thumb.JpegXLSupported() && w.conf.JpegXLEnabled() && w.conf.JpegXLDecoderBin() == "") {
		log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(imageName)), f.FileType())

		// Create PNG or JPEG image from source file.
		switch fs.LowerExt(imageName) {
		case fs.ExtPng:
			_, err = thumb.Png(f.FileName(), imageName, f.Orientation())
		case fs.ExtJpeg:
			_, err = thumb.Jpeg(f.FileName(), imageName, f.Orientation())
		default:
			return nil, fmt.Errorf("convert: unsupported target format %s (%s)", fs.LowerExt(imageName), clean.Log(f.RootRelName()))
		}

		// Check result.
		if err == nil {
			log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(imageName)), time.Since(start), f.FileType())
			return NewMediaFile(imageName)
		} else if !f.IsTiff() && !f.IsWebp() && !f.IsHeic() && !f.IsAvif() && !f.IsJpegXL() {
			// See https://github.com/photoprism/photoprism/issues/1612
			// for TIFF file format compatibility. HEIC/HEIF, AVIF, and JPEG XL keep
			// the external conversion fallback until we can rely on native libvips
			// support in every supported runtime.
			return nil, disk.AsInsufficientStorage(err)
		}
	}

	// Run external commands for other formats.
	var cmds ConvertCmds
	var useMutex bool
	var expectedMime string

	switch fs.LowerExt(imageName) {
	case fs.ExtPng:
		cmds, useMutex, err = w.PngConvertCmds(f, imageName)
		expectedMime = header.ContentTypePng
	case fs.ExtJpeg:
		cmds, useMutex, err = w.JpegConvertCmds(f, imageName, xmpName)
		expectedMime = header.ContentTypeJpeg
	default:
		return nil, fmt.Errorf("convert: unsupported target format %s (%s)", fs.LowerExt(imageName), clean.Log(f.RootRelName()))
	}

	if err != nil {
		return nil, err
	} else if len(cmds) == 0 {
		return nil, fmt.Errorf("file type %s not supported", f.FileType())
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		w.cmdMutex.Lock()
		defer w.cmdMutex.Unlock()
	}

	if fs.FileExistsNotEmpty(imageName) {
		return NewMediaFile(imageName)
	}

	// Try compatible converters.
	for _, c := range cmds {
		// Fetch command output.
		var out bytes.Buffer
		var stderr bytes.Buffer

		cmd := c.Cmd
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("HOME=%s", w.conf.CmdCachePath()),
			fmt.Sprintf("LD_LIBRARY_PATH=%s", w.conf.CmdLibPath()),
		}...)

		log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(imageName)), filepath.Base(cmd.Path))

		// Log exact command in debug mode.
		log.Debug(cmd.String())

		// Run convert command.
		if err = cmd.Run(); err != nil {
			if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
				err = errors.New(errStr)
			}

			log.Debugf("convert: %s (%s)", clean.Error(err), filepath.Base(cmd.Path))
			continue
		} else if fs.FileExistsNotEmpty(imageName) {
			// The command wrote the target file directly (e.g. Darktable, RawTherapee).
		} else if res := out.Bytes(); len(res) < 512 || !mimetype.Detect(res).Is(expectedMime) {
			continue
		} else if err = os.WriteFile(imageName, res, fs.ModeFile); err != nil {
			log.Tracef("convert: %s (%s)", err, filepath.Base(cmd.Path))
			continue
		}

		// Reject output flagged untrustworthy by its stderr (e.g. a RawTherapee render of an
		// unsupported sensor) so the loop falls back to the next converter.
		if c.StderrRejected(stderr.String()) {
			log.Debugf("convert: discarding %s from %s (untrustworthy output)", clean.Log(filepath.Base(imageName)), filepath.Base(cmd.Path))
			if removeErr := os.Remove(imageName); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Tracef("convert: %s (%s)", removeErr, filepath.Base(cmd.Path))
			}
			continue
		}

		// Reject undecodable output (e.g. a truncated/bogus embedded RAW preview that passed
		// the MIME sniff) so the loop tries the next converter instead of indexing a file
		// whose thumbnails will fail.
		if c.VerifyImage {
			if err = thumb.Verify(imageName); err != nil {
				log.Debugf("convert: discarding undecodable %s from %s (%s)", clean.Log(filepath.Base(imageName)), filepath.Base(cmd.Path), clean.Error(err))
				if removeErr := os.Remove(imageName); removeErr != nil && !os.IsNotExist(removeErr) {
					log.Tracef("convert: %s (%s)", removeErr, filepath.Base(cmd.Path))
				}
				continue
			}
		}

		log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(imageName)), time.Since(start), filepath.Base(cmd.Path))
		fileOrientation = c.Orientation
		break
	}

	// Ok?
	if err != nil {
		return nil, disk.AsInsufficientStorage(err)
	}

	// Create a MediaFile instance from the generated file.
	if result, err = NewMediaFile(imageName); err != nil {
		return result, err
	}

	// Change the Exif orientation of the generated file if required.
	if fileOrientation == media.ResetOrientation {
		if err = result.ChangeOrientation(1); err != nil {
			log.Warnf("convert: %s in %s (change orientation)", err, clean.Log(result.RootRelName()))
		}
	}

	return result, nil
}
