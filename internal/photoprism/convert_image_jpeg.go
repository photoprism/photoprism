package photoprism

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/raw"
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
	rawEnabled := w.conf.RawEnabled()

	// On a Mac, use the Apple Scriptable image processing system to convert images to JPEG,
	// see https://ss64.com/osx/sips.html.
	if (f.IsRaw() && rawEnabled || f.IsHeif()) && w.conf.SipsEnabled() && w.sipsExclude.Allow(fileExt) {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.SipsBin(), "-Z", maxSize, "-s", "format", "jpeg", "--out", jpegName, f.FileName())),
		)
	}

	// Use FFmpeg to extract video stills from videos, e.g. to use them as cover images.
	if f.IsAnimated() && !f.IsWebp() && w.conf.FFmpegEnabled() && w.FFmpegAllowed(f) {
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

	// Convert RAW files to JPEG: Darktable/RawTherapee render when RAW conversion is enabled, while
	// ExifTool preview extraction also runs when rendering is disabled, so existing previews stay usable.
	if f.IsRaw() {
		if rawEnabled {
			if w.conf.DarktableEnabled() && w.darktableExclude.Allow(fileExt) {
				cmd, mutex := raw.DarktableCmd(raw.DarktableOptions{
					Bin:       w.conf.DarktableBin(),
					RawName:   f.FileName(),
					XmpName:   xmpName,
					JpegName:  jpegName,
					MaxSize:   w.conf.JpegSize(),
					Presets:   w.conf.RawPresets(),
					ConfigDir: conf.DarktableConfigPath(),
					CacheDir:  conf.DarktableCachePath(),
				})
				useMutex = mutex
				result = append(result, NewConvertCmd(cmd))
			}

			// Render the RAW with RawTherapee when Darktable is unavailable or fails. For formats in the
			// discard set (e.g. CR3) the output is rejected on an untrustworthy decode (raw.DecoderErrors)
			// so the embedded preview wins; other RAW formats (e.g. .raw/.kdc) keep the render because
			// RawTherapee may be their only working decoder.
			if w.conf.RawTherapeeEnabled() && w.rawTherapeeExclude.Allow(fileExt) {
				profile := filepath.Join(conf.AssetsPath(), "profiles", "raw.pp3")

				rtCmd := NewConvertCmd(
					raw.TherapeeCmd(w.conf.RawTherapeeBin(), f.FileName(), jpegName, profile, int(w.conf.JpegQuality())),
				)

				if raw.DiscardRenderOnWarning(fileExt) {
					rtCmd = rtCmd.WithStderrRejection(raw.DecoderErrors...)
				}

				result = append(result, rtCmd)
			}
		}

		// Extract the embedded camera preview (largest first): the fallback after the RAW developers,
		// and the only option when RAW rendering is disabled. Colors stay correct for sensors they
		// cannot identify (recent Canon CR3 bodies otherwise come out magenta). Skipped if unusable.
		if w.conf.ExifToolEnabled() && raw.PreviewExtAllowed(fileExt) {
			result = append(result, NewConvertCmd(raw.ExifToolJpgFromRawCmd(w.conf.ExifToolBin(), f.FileName())).WithImageVerification())
			result = append(result, NewConvertCmd(raw.ExifToolPreviewImageCmd(w.conf.ExifToolBin(), f.FileName())).WithImageVerification())
		}
	}

	// Use "djxl" to convert JPEG XL images as a fallback when libvips lacks native support.
	if f.IsJpegXL() && w.conf.JpegXLEnabled() && w.conf.JpegXLDecoderBin() != "" {
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

	// Use ExifTool to extract embedded Photoshop thumbnails as a lightweight fallback
	// when full rasterization of TIFF/PSD previews is unavailable or fails.
	if (f.IsTiff() || f.IsPsd()) && w.conf.ExifToolEnabled() {
		result = append(result, NewConvertCmd(
			// #nosec G204 -- arguments are built from validated config and file paths.
			exec.Command(w.conf.ExifToolBin(), "-q", "-q", "-b", "-PhotoshopThumbnail", f.FileName())).WithImageVerification(),
		)
	}

	// No suitable converter found?
	if len(result) == 0 {
		if f.IsAnimated() && w.conf.FFmpegEnabled() && !w.FFmpegAllowed(f) {
			return result, useMutex, fmt.Errorf("format %s is on the FFmpeg exclude list", w.ffmpegExclude.Match(f.MetaData().Codec, f.VideoInfo().VideoCodec, f.FileType().String()))
		}
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
