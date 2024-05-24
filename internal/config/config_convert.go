package config

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
	return !c.DisableImageMagick()
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
