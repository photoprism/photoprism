package config

import (
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/media"
)

// VectorEnabled checks if indexing and conversion of vector graphics is enabled.
func (c *Config) VectorEnabled() bool {
	return !c.DisableVectors()
}

// RsvgConvertBin returns the rsvg-convert executable file name.
func (c *Config) RsvgConvertBin() string {
	return FindBin(c.options.RsvgConvertBin, "rsvg-convert")
}

// RsvgConvertEnabled checks if rsvg-convert is enabled for SVG conversion.
func (c *Config) RsvgConvertEnabled() bool {
	return !c.DisableVectors()
}

// ImageMagickBin returns the ImageMagick "convert" executable file name.
func (c *Config) ImageMagickBin() string {
	return FindBin(c.options.ImageMagickBin, "convert", "magick")
}

// ImageMagickExclude returns the file extensions not to be used with ImageMagick.
func (c *Config) ImageMagickExclude() string {
	return c.options.ImageMagickExclude
}

// ImageMagickEnabled checks if ImageMagick can be used for converting media files.
func (c *Config) ImageMagickEnabled() bool {
	return !c.DisableImageMagick()
}

// JpegXLDecoderBin returns the external JPEG XL decoder executable file name,
// which is used as a fallback when libvips lacks native JPEG XL support.
func (c *Config) JpegXLDecoderBin() string {
	return FindBin("", "djxl")
}

// JpegXLEnabled checks if JPEG XL file format support is enabled.
func (c *Config) JpegXLEnabled() bool {
	return !c.DisableJpegXL()
}

// DisableJpegXL checks if JPEG XL file format support is disabled. It stays enabled
// as long as the external "djxl" decoder is available or libvips can decode JPEG XL
// natively, so the format only turns off when neither option exists.
func (c *Config) DisableJpegXL() bool {
	if jpegXLDisabled(c.options.DisableJpegXL, c.JpegXLDecoderBin() != "", thumb.JpegXLSupported) {
		c.options.DisableJpegXL = true
	}

	return c.options.DisableJpegXL
}

// jpegXLDisabled reports whether JPEG XL support is unavailable, given the explicit
// disable option and external decoder availability. nativeSupported is a thunk so the
// libvips capability probe runs only when no external decoder is present, keeping
// config introspection on standard installs (which ship "djxl") free of a libvips start.
func jpegXLDisabled(explicit, decoderAvailable bool, nativeSupported func() bool) bool {
	switch {
	case explicit:
		return true
	case decoderAvailable:
		return false
	default:
		return !nativeSupported()
	}
}

// HeifConvertBin returns the name of the "heif-dec" executable ("heif-convert" in earlier libheif versions).
// see https://github.com/photoprism/photoprism/issues/4439
func (c *Config) HeifConvertBin() string {
	return FindBin(c.options.HeifConvertBin, "heif-dec", "heif-convert")
}

// HeifConvertOrientation returns the Exif orientation of images generated with libheif (keep, reset).
func (c *Config) HeifConvertOrientation() media.Orientation {
	return media.ParseOrientation(c.options.HeifConvertOrientation, media.KeepOrientation)
}

// HeifConvertEnabled checks if heif-convert is enabled for HEIF conversion.
func (c *Config) HeifConvertEnabled() bool {
	return !c.DisableHeifConvert()
}

// SipsEnabled checks if SIPS is enabled for RAW conversion.
func (c *Config) SipsEnabled() bool {
	return !c.DisableSips()
}

// SipsBin returns the SIPS executable file name.
func (c *Config) SipsBin() string {
	return FindBin(c.options.SipsBin, "sips")
}

// SipsExclude returns the file extensions no not be used with Sips.
func (c *Config) SipsExclude() string {
	return c.options.SipsExclude
}
