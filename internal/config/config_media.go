package config

import (
	"github.com/photoprism/photoprism/pkg/media"
)

// VectorEnabled checks if indexing and conversion of vector graphics is enabled.
func (c *Config) VectorEnabled() bool {
	return !c.DisableVectors()
}

// RsvgConvertBin returns the rsvg-convert executable file name.
func (c *Config) RsvgConvertBin() string {
	return findBin(c.options.RsvgConvertBin, "rsvg-convert")
}

// RsvgConvertEnabled checks if rsvg-convert is enabled for SVG conversion.
func (c *Config) RsvgConvertEnabled() bool {
	return !c.DisableVectors()
}

// ImageMagickBin returns the ImageMagick "convert" executable file name.
func (c *Config) ImageMagickBin() string {
	return findBin(c.options.ImageMagickBin, "convert")
}

// ImageMagickExclude returns the file extensions not to be used with ImageMagick.
func (c *Config) ImageMagickExclude() string {
	return c.options.ImageMagickExclude
}

// ImageMagickEnabled checks if ImageMagick can be used for converting media files.
func (c *Config) ImageMagickEnabled() bool {
	return !c.DisableImageMagick()
}

// JpegXLDecoderBin returns the JPEG XL decoder executable file name.
func (c *Config) JpegXLDecoderBin() string {
	return findBin("", "djxl")
}

// JpegXLEnabled checks if JPEG XL file format support is enabled.
func (c *Config) JpegXLEnabled() bool {
	return !c.DisableJpegXL()
}

// DisableJpegXL checks if JPEG XL file format support is disabled.
func (c *Config) DisableJpegXL() bool {
	if c.options.DisableJpegXL {
		return true
	} else if c.JpegXLDecoderBin() == "" {
		c.options.DisableJpegXL = true
	}

	return c.options.DisableJpegXL
}

// HeifConvertBin returns the name of the "heif-dec" executable ("heif-convert" in earlier libheif versions).
// see https://github.com/photoprism/photoprism/issues/4439
func (c *Config) HeifConvertBin() string {
	return findBin(c.options.HeifConvertBin, "heif-dec", "heif-convert")
}

// HeifConvertOrientation returns the Exif orientation of images generated with libheif (auto, strip, keep).
func (c *Config) HeifConvertOrientation() media.Orientation {
	return media.ParseOrientation(c.options.HeifConvertOrientation, media.ResetOrientation)
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
	return findBin(c.options.SipsBin, "sips")
}

// SipsExclude returns the file extensions no not be used with Sips.
func (c *Config) SipsExclude() string {
	return c.options.SipsExclude
}
