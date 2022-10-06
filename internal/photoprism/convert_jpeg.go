package photoprism

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/thumb"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ToJpeg converts a single image file to JPEG if possible.
func (c *Convert) ToJpeg(f *MediaFile, force bool) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: file is nil - possible bug")
	}

	if !f.Exists() {
		return nil, fmt.Errorf("convert: %s not found", clean.Log(f.RootRelName()))
	} else if f.Empty() {
		return nil, fmt.Errorf("convert: %s is empty", clean.Log(f.RootRelName()))
	}

	if f.IsJpeg() {
		return f, nil
	}

	jpegName := fs.ImageJPEG.FindFirst(f.FileName(), []string{c.conf.SidecarPath(), fs.HiddenPath}, c.conf.OriginalsPath(), false)

	mediaFile, err := NewMediaFile(jpegName)

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
	} else {
		jpegName = fs.FileName(f.FileName(), c.conf.SidecarPath(), c.conf.OriginalsPath(), fs.ExtJPEG)
	}

	if !c.conf.SidecarWritable() {
		return nil, fmt.Errorf("convert: disabled in read-only mode (%s)", clean.Log(f.RootRelName()))
	}

	fileName := f.RelName(c.conf.OriginalsPath())
	xmpName := fs.XmpFile.Find(f.FileName(), false)

	event.Publish("index.converting", event.Data{
		"fileType": f.FileType(),
		"fileName": fileName,
		"baseName": filepath.Base(fileName),
		"xmpName":  filepath.Base(xmpName),
	})

	start := time.Now()

	if f.IsImageOther() {
		log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(jpegName)), f.FileType())

		_, err = thumb.Jpeg(f.FileName(), jpegName, f.Orientation())

		if err != nil {
			return nil, err
		}

		log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(jpegName)), time.Since(start), f.FileType())

		return NewMediaFile(jpegName)
	}

	cmd, useMutex, err := c.JpegConvertCommand(f, jpegName, xmpName)

	if err != nil {
		return nil, err
	}

	if useMutex {
		// Make sure only one command is executed at a time.
		// See https://photo.stackexchange.com/questions/105969/darktable-cli-fails-because-of-locked-database-file
		c.cmdMutex.Lock()
		defer c.cmdMutex.Unlock()
	}

	if fs.FileExists(jpegName) {
		return NewMediaFile(jpegName)
	}

	// Fetch command output.
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Env = []string{fmt.Sprintf("HOME=%s", c.conf.CmdCachePath())}

	log.Infof("convert: converting %s to %s (%s)", clean.Log(filepath.Base(fileName)), clean.Log(filepath.Base(jpegName)), filepath.Base(cmd.Path))

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run convert command.
	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return nil, errors.New(stderr.String())
		} else {
			return nil, err
		}
	}

	log.Infof("convert: %s created in %s (%s)", clean.Log(filepath.Base(jpegName)), time.Since(start), filepath.Base(cmd.Path))

	return NewMediaFile(jpegName)
}

// JpegConvertCommand returns the command for converting files to JPEG, depending on the format.
func (c *Convert) JpegConvertCommand(f *MediaFile, jpegName string, xmpName string) (result *exec.Cmd, useMutex bool, err error) {
	if f == nil {
		return result, useMutex, fmt.Errorf("file is nil - possible bug")
	}

	fileExt := f.Extension()
	maxSize := strconv.Itoa(c.conf.JpegSize())

	// Select conversion command depending on the file type and runtime environment.
	if (f.IsRaw() || f.IsHEIF() || f.IsAVIF()) && c.conf.SipsEnabled() && c.sipsBlacklist.Ok(fileExt) {
		result = exec.Command(c.conf.SipsBin(), "-Z", maxSize, "-s", "format", "jpeg", "--out", jpegName, f.FileName())
	} else if f.IsRaw() && c.conf.RawEnabled() || f.IsAVIF() {
		if c.conf.DarktableEnabled() && c.darktableBlacklist.Ok(fileExt) {

			var args []string

			// Set RAW, XMP, and JPEG filenames.
			if xmpName != "" {
				args = []string{f.FileName(), xmpName, jpegName}
			} else {
				args = []string{f.FileName(), jpegName}
			}

			// Set RAW to JPEG conversion options.
			if c.conf.RawPresets() {
				useMutex = true // can run one instance only with presets enabled
				args = append(args, "--width", maxSize, "--height", maxSize, "--hq", "true", "--upscale", "false")
			} else {
				useMutex = false // --apply-custom-presets=false disables locking
				args = append(args, "--apply-custom-presets", "false", "--width", maxSize, "--height", maxSize, "--hq", "true", "--upscale", "false")
			}

			// Set library, config, and cache location.
			args = append(args, "--core", "--library", ":memory:")

			if dir := conf.DarktableConfigPath(); dir != "" {
				args = append(args, "--configdir", dir)
			}

			if dir := conf.DarktableCachePath(); dir != "" {
				args = append(args, "--cachedir", dir)
			}

			result = exec.Command(c.conf.DarktableBin(), args...)
		} else if c.conf.RawtherapeeEnabled() && c.rawtherapeeBlacklist.Ok(fileExt) {
			jpegQuality := fmt.Sprintf("-j%d", c.conf.JpegQuality())
			profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

			args := []string{"-o", jpegName, "-p", profile, "-s", "-d", jpegQuality, "-js3", "-b8", "-c", f.FileName()}

			result = exec.Command(c.conf.RawtherapeeBin(), args...)
		} else {
			return nil, useMutex, fmt.Errorf("no suitable converter found")
		}
	} else if f.IsVideo() && c.conf.FFmpegEnabled() {
		result = exec.Command(c.conf.FFmpegBin(), "-y", "-i", f.FileName(), "-ss", "00:00:00.001", "-vframes", "1", jpegName)
	} else if f.IsHEIF() && c.conf.HeifConvertEnabled() {
		result = exec.Command(c.conf.HeifConvertBin(), "-q", c.conf.JpegQuality().String(), f.FileName(), jpegName)
	} else {
		return nil, useMutex, fmt.Errorf("file type %s not supported", f.FileType())
	}

	// Log convert command in trace mode only as it exposes server internals.
	if result != nil {
		log.Tracef("convert: %s", result.String())
	}

	return result, useMutex, nil
}
