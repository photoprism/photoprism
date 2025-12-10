package photoprism

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
)

// JpegConvertCmds returns the supported commands for converting a MediaFile to JPEG, sorted by priority.
func (w *Convert) JpegConvertCmds(f *MediaFile, jpegName string, xmpName string) (result ConvertCmds, useMutex bool, err error) {
	result = NewConvertCmds()

	if f == nil {
		return result, useMutex, fmt.Errorf("no media file provided for processing - you may have found a bug")
	}

	// Add suitable conversion commands depending on the file type, codec, runtime environment, and configuration.
	fileExt := f.Extension()
	maxSize := strconv.Itoa(w.conf.JpegSize())

	// On a Mac, use the Apple Scriptable image processing system to convert images to JPEG,
	// see https://ss64.com/osx/sips.html.
	if (f.IsRaw() || f.IsHeif()) && w.conf.SipsEnabled() && w.sipsExclude.Allow(fileExt) {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.SipsBin(), "-Z", maxSize, "-s", "format", "jpeg", "--out", jpegName, f.FileName())),
		)
	}

	// Use FFmpeg to extract video stills from videos, e.g. to use them as cover images.
	if f.IsAnimated() && !f.IsWebp() && w.conf.FFmpegEnabled() {
		result = append(result, NewConvertCmd(
			ffmpeg.ExtractJpegImageCmd(f.FileName(), jpegName, encode.NewPreviewImageOptions(w.conf.FFmpegBin(), f.Duration()))),
		)
	}

	// Use "heif-dec" or "heif-convert" to convert HEIC/HEIF and AVIF image files to JPEG.
	if (f.IsHeic() || f.IsAvif()) && w.conf.HeifConvertEnabled() {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.HeifConvertBin(), "-q", w.conf.JpegQuality().String(), f.FileName(), jpegName)).
			WithOrientation(w.conf.HeifConvertOrientation()),
		)
	}

	// Convert RAW files to JPEG with Darktable and/or RawTherapee.
	if f.IsRaw() && w.conf.RawEnabled() {
		if w.conf.DarktableEnabled() && w.darktableExclude.Allow(fileExt) {
			var args []string

			// Set RAW, XMP, and JPEG filenames.
			if xmpName != "" {
				args = []string{f.FileName(), xmpName, jpegName}
			} else {
				args = []string{f.FileName(), jpegName}
			}

			// Set RAW to JPEG conversion options.
			if w.conf.RawPresets() {
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

			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.DarktableBin(), args...)),
			)
		}

		if w.conf.RawTherapeeEnabled() && w.rawTherapeeExclude.Allow(fileExt) {
			jpegQuality := fmt.Sprintf("-j%d", w.conf.JpegQuality())
			profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

			args := []string{"-o", jpegName, "-p", profile, "-s", "-d", jpegQuality, "-js3", "-b8", "-c", f.FileName()}

			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.RawTherapeeBin(), args...)),
			)
		}
	}

	// Use ExifTool to extract JPEG thumbnails from Digital Negative (DNG) files.
	if f.IsDng() && w.conf.ExifToolEnabled() {
		// Example: exiftool -b -PreviewImage -w IMG_4691.DNG.jpg IMG_4691.DNG
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.ExifToolBin(), "-q", "-q", "-b", "-PreviewImage", f.FileName())),
		)
	}

	// Use "djxl" to convert JPEG XL images if installed and enabled.
	if f.IsJpegXL() && w.conf.JpegXLEnabled() {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.JpegXLDecoderBin(), f.FileName(), jpegName)),
		)
	}

	// Use ImageMagick for other media file formats if the type and extension are allowed.
	if w.conf.ImageMagickEnabled() && w.imageMagickExclude.Allow(fileExt) {
		resize := fmt.Sprintf("%dx%d>", w.conf.JpegSize(), w.conf.JpegSize())
		quality := fmt.Sprintf("%d", w.conf.JpegQuality())

		switch {
		case f.IsImage() && !f.IsJpegXL() && !f.IsRaw() && !f.IsHeif():
			args := []string{f.FileName() + "[0]", "-background", "white", "-alpha", "remove", "-alpha", "off", "-resize", resize, "-quality", quality, jpegName}
			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		case f.IsVector() && w.conf.VectorEnabled():
			args := []string{f.FileName() + "[0]", "-background", "black", "-alpha", "remove", "-alpha", "off", "-resize", resize, "-quality", quality, jpegName}
			result = append(result, NewConvertCmd(
				// #nosec G204 -- arguments are built from validated config and file paths.
				exec.Command(w.conf.ImageMagickBin(), args...)),
			)
		case f.IsDocument():
			args := []string{"-colorspace", "sRGB", "-density", "300", f.FileName() + "[0]", "-background", "white", "-alpha", "remove", "-alpha", "off", "-resize", resize, "-quality", quality, jpegName}
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
