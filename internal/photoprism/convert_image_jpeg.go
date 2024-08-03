package photoprism

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/photoprism/photoprism/internal/ffmpeg"
)

// JpegConvertCommands returns the supported commands for converting a MediaFile to JPEG, sorted by priority.
func (w *Convert) JpegConvertCommands(f *MediaFile, jpegName string, xmpName string) (result ConvertCommands, useMutex bool, err error) {
	result = NewConvertCommands()

	if f == nil {
		return result, useMutex, fmt.Errorf("file is nil - you may have found a bug")
	}

	// Find conversion command depending on the file type and runtime environment.
	fileExt := f.Extension()
	maxSize := strconv.Itoa(w.conf.JpegSize())

	// Apple Scriptable image processing system: https://ss64.com/osx/sips.html
	if (f.IsRaw() || f.IsHEIF()) && w.conf.SipsEnabled() && w.sipsExclude.Allow(fileExt) {
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.SipsBin(), "-Z", maxSize, "-s", "format", "jpeg", "--out", jpegName, f.FileName())),
		)
	}

	// Extract a still image to be used as preview.
	if f.IsAnimated() && !f.IsWebP() && w.conf.FFmpegEnabled() {
		// Use "ffmpeg" to extract a JPEG still image from the video.
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.FFmpegBin(), "-y", "-ss", ffmpeg.PreviewTimeOffset(f.Duration()), "-i", f.FileName(), "-vframes", "1", jpegName)),
		)
	}

	// Use heif-convert for HEIC/HEIF and AVIF image files.
	if (f.IsHEIC() || f.IsAVIF()) && w.conf.HeifConvertEnabled() {
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.HeifConvertBin(), "-q", w.conf.JpegQuality().String(), f.FileName(), jpegName)).
			WithOrientation(w.conf.HeifConvertOrientation()),
		)
	}

	// RAW files may be concerted with Darktable and RawTherapee.
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

			result = append(result, NewConvertCommand(
				exec.Command(w.conf.DarktableBin(), args...)),
			)
		}

		if w.conf.RawTherapeeEnabled() && w.rawTherapeeExclude.Allow(fileExt) {
			jpegQuality := fmt.Sprintf("-j%d", w.conf.JpegQuality())
			profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

			args := []string{"-o", jpegName, "-p", profile, "-s", "-d", jpegQuality, "-js3", "-b8", "-c", f.FileName()}

			result = append(result, NewConvertCommand(
				exec.Command(w.conf.RawTherapeeBin(), args...)),
			)
		}
	}

	// Extract preview image from DNG files.
	if f.IsDNG() && w.conf.ExifToolEnabled() {
		// Example: exiftool -b -PreviewImage -w IMG_4691.DNG.jpg IMG_4691.DNG
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.ExifToolBin(), "-q", "-q", "-b", "-PreviewImage", f.FileName())),
		)
	}

	// Decode JPEG XL image if support is enabled.
	if f.IsJpegXL() && w.conf.JpegXLEnabled() {
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.JpegXLDecoderBin(), f.FileName(), jpegName)),
		)
	}

	// Try ImageMagick for other image file formats if allowed.
	if w.conf.ImageMagickEnabled() && w.imageMagickExclude.Allow(fileExt) &&
		(f.IsImage() && !f.IsJpegXL() && !f.IsRaw() && !f.IsHEIF() || f.IsVector() && w.conf.VectorEnabled()) {
		quality := fmt.Sprintf("%d", w.conf.JpegQuality())
		resize := fmt.Sprintf("%dx%d>", w.conf.JpegSize(), w.conf.JpegSize())
		args := []string{f.FileName(), "-flatten", "-resize", resize, "-quality", quality, jpegName}
		result = append(result, NewConvertCommand(
			exec.Command(w.conf.ImageMagickBin(), args...)),
		)
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
