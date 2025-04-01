package photoprism

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// PngConvertCmds returns commands for converting a media file to PNG, if possible.
func (w *Convert) PngConvertCmds(f *MediaFile, pngName string) (result ConvertCmds, useMutex bool, err error) {
	result = NewConvertCmds()

	if f == nil {
		return result, useMutex, fmt.Errorf("no media file provided for processing - you may have found a bug")
	}

	// Find conversion command depending on the file type and runtime environment.
	fileExt := f.Extension()
	maxSize := strconv.Itoa(w.conf.PngSize())

	// Apple Scriptable image processing system: https://ss64.com/osx/sips.html
	if (f.IsRaw() || f.IsHeif()) && w.conf.SipsEnabled() && w.sipsExclude.Allow(fileExt) {
		result = append(result, NewConvertCmd(
			exec.Command(w.conf.SipsBin(), "-Z", maxSize, "-s", "format", "png", "--out", pngName, f.FileName())),
		)
	}

	// Extract a video still image that can be used as preview.
	if f.IsAnimated() && !f.IsWebp() && w.conf.FFmpegEnabled() {
		// Use "ffmpeg" to extract a PNG still image from the video.
		result = append(result, NewConvertCmd(
			ffmpeg.ExtractPngImageCmd(f.FileName(), pngName, encode.NewPreviewImageOptions(w.conf.FFmpegBin(), f.Duration()))),
		)
	}

	// Use heif-convert for HEIC/HEIF and AVIF image files.
	if (f.IsHeic() || f.IsAvif()) && w.conf.HeifConvertEnabled() {
		result = append(result, NewConvertCmd(
			exec.Command(w.conf.HeifConvertBin(), f.FileName(), pngName)).
			WithOrientation(w.conf.HeifConvertOrientation()),
		)
	}

	// Decode JPEG XL image if support is enabled.
	if f.IsJpegXL() && w.conf.JpegXLEnabled() {
		result = append(result, NewConvertCmd(
			exec.Command(w.conf.JpegXLDecoderBin(), f.FileName(), pngName)),
		)
	}

	// SVG vector graphics can be converted with librsvg if installed,
	// otherwise try to convert the media file with ImageMagick.
	if w.conf.RsvgConvertEnabled() && f.IsSVG() {
		args := []string{"-a", "-f", "png", "-o", pngName, f.FileName()}
		result = append(result, NewConvertCmd(
			exec.Command(w.conf.RsvgConvertBin(), args...)),
		)
	} else if w.conf.ImageMagickEnabled() && w.imageMagickExclude.Allow(fileExt) {
		if f.IsImage() && !f.IsJpegXL() && !f.IsRaw() && !f.IsHeif() || f.IsVector() && w.conf.VectorEnabled() {
			resize := fmt.Sprintf("%dx%d>", w.conf.PngSize(), w.conf.PngSize())
			args := []string{f.FileName(), "-flatten", "-resize", resize, pngName}
			result = append(result, NewConvertCmd(
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		} else if f.IsDocument() {
			resize := fmt.Sprintf("%dx%d>", w.conf.PngSize(), w.conf.PngSize())
			args := []string{f.FileName() + "[0]", "-background", "white", "-alpha", "remove", "-alpha", "off", "-resize", resize, pngName}
			result = append(result, NewConvertCmd(
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		}
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
