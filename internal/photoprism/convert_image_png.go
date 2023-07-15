package photoprism

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/photoprism/photoprism/internal/ffmpeg"
)

// PngConvertCommands returns commands for converting a media file to PNG, if possible.
func (c *Convert) PngConvertCommands(f *MediaFile, pngName string) (result []*exec.Cmd, useMutex bool, err error) {
	result = make([]*exec.Cmd, 0, 2)

	if f == nil {
		return result, useMutex, fmt.Errorf("file is nil - you may have found a bug")
	}

	// Find conversion command depending on the file type and runtime environment.
	fileExt := f.Extension()
	maxSize := strconv.Itoa(c.conf.PngSize())

	// Apple Scriptable image processing system: https://ss64.com/osx/sips.html
	if (f.IsRaw() || f.IsHEIF()) && c.conf.SipsEnabled() && c.sipsBlacklist.Allow(fileExt) {
		result = append(result, exec.Command(c.conf.SipsBin(), "-Z", maxSize, "-s", "format", "png", "--out", pngName, f.FileName()))
	}

	// Extract a video still image that can be used as preview.
	if f.IsAnimated() && !f.IsWebP() && c.conf.FFmpegEnabled() {
		// Use "ffmpeg" to extract a PNG still image from the video.
		result = append(result, exec.Command(c.conf.FFmpegBin(), "-y", "-i", f.FileName(), "-ss", ffmpeg.PreviewTimeOffset(f.Duration()), "-vframes", "1", pngName))
	}

	// Use heif-convert for HEIC/HEIF and AVIF image files.
	if (f.IsHEIC() || f.IsAVIF()) && c.conf.HeifConvertEnabled() {
		result = append(result, exec.Command(c.conf.HeifConvertBin(), f.FileName(), pngName))
	}

	// Decode JPEG XL image if support is enabled.
	if f.IsJpegXL() && c.conf.JpegXLEnabled() {
		result = append(result, exec.Command(c.conf.JpegXLDecoderBin(), f.FileName(), pngName))
	}

	// Try ImageMagick for other image file formats if allowed.
	if c.conf.ImageMagickEnabled() && c.imagemagickBlacklist.Allow(fileExt) &&
		(f.IsImage() && !f.IsJpegXL() && !f.IsRaw() && !f.IsHEIF() || f.IsVector() && c.conf.VectorEnabled()) {
		resize := fmt.Sprintf("%dx%d>", c.conf.PngSize(), c.conf.PngSize())
		args := []string{f.FileName(), "-flatten", "-resize", resize, pngName}
		result = append(result, exec.Command(c.conf.ImageMagickBin(), args...))
	} else if f.IsVector() && c.conf.RsvgConvertEnabled() {
		// Vector graphics may be also be converted with librsvg if installed.
		args := []string{"-a", "-f", "png", "-o", pngName, f.FileName()}
		result = append(result, exec.Command(c.conf.RsvgConvertBin(), args...))
	}

	// No suitable converter found?
	if len(result) == 0 {
		return result, useMutex, fmt.Errorf("file type %s not supported", f.FileType())
	}

	// Log convert command in trace mode only as it exposes server internals.
	for i, cmd := range result {
		if i == 0 {
			log.Tracef("convert: %s", cmd.String())
		} else {
			log.Tracef("convert: %s (alternative)", cmd.String())
		}
	}

	return result, useMutex, nil
}
