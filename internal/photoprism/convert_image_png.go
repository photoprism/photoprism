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

	// Add suitable conversion commands depending on the file type, codec, runtime environment, and configuration.
	fileExt := f.Extension()
	maxSize := strconv.Itoa(w.conf.PngSize())

	// On a Mac, use the Apple Scriptable image processing system to convert images to PNG,
	// see https://ss64.com/osx/sips.html.
	if (f.IsRaw() || f.IsHeif()) && w.conf.SipsEnabled() && w.sipsExclude.Allow(fileExt) {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.SipsBin(), "-Z", maxSize, "-s", "format", "png", "--out", pngName, f.FileName())),
		)
	}

	// Use FFmpeg to extract video stills from videos, e.g. to use them as cover images.
	if f.IsAnimated() && !f.IsWebp() && w.conf.FFmpegEnabled() {
		// Use "ffmpeg" to extract a PNG still image from the video.
		result = append(result, NewConvertCmd(
			ffmpeg.ExtractPngImageCmd(f.FileName(), pngName, encode.NewPreviewImageOptions(w.conf.FFmpegBin(), f.Duration()))),
		)
	}

	// Use "heif-dec" or "heif-convert" to convert HEIC/HEIF and AVIF image files to PNG.
	if (f.IsHeic() || f.IsAvif()) && w.conf.HeifConvertEnabled() {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.HeifConvertBin(), f.FileName(), pngName)).
			WithOrientation(w.conf.HeifConvertOrientation()),
		)
	}

	// Use "djxl" to convert JPEG XL images if installed and enabled.
	if f.IsJpegXL() && w.conf.JpegXLEnabled() {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.JpegXLDecoderBin(), f.FileName(), pngName)),
		)
	}

	// Use librsvg to convert SVG vector graphics to PNG if installed and enabled.
	if w.conf.RsvgConvertEnabled() && f.IsSVG() {
		args := []string{"-a", "-f", "png", "-o", pngName, f.FileName()}
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.RsvgConvertBin(), args...)),
		)
	}

	// Use ImageMagick for other media file formats if the type and extension are allowed.
	if w.conf.ImageMagickEnabled() && w.imageMagickExclude.Allow(fileExt) {
		resize := fmt.Sprintf("%dx%d>", w.conf.PngSize(), w.conf.PngSize())
		switch {
		case f.IsImage() && !f.IsJpegXL() && !f.IsRaw() && !f.IsHeif():
			args := []string{f.FileName(), "-flatten", "-resize", resize, pngName}
			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		case f.IsVector() && w.conf.VectorEnabled():
			args := []string{f.FileName() + "[0]", "-resize", resize, pngName}
			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		case f.IsDocument():
			args := []string{"-colorspace", "sRGB", "-density", "300", f.FileName() + "[0]", "-background", "white", "-alpha", "remove", "-alpha", "off", "-resize", resize, pngName}
			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
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
