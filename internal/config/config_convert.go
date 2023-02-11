package config

// VectorEnabled checks if indexing and conversion of vector graphics is enabled.
func (c *Config) VectorEnabled() bool {
	return !c.DisableVector()
}

// RsvgConvertBin returns the rsvg-convert executable file name.
func (c *Config) RsvgConvertBin() string {
	return findBin(c.options.RsvgConvertBin, "rsvg-convert")
}

// RsvgConvertEnabled checks if rsvg-convert is enabled for SVG conversion.
func (c *Config) RsvgConvertEnabled() bool {
	return !c.DisableVector()
}

// ImageMagickBin returns the ImageMagick "convert" executable file name.
func (c *Config) ImageMagickBin() string {
	return findBin(c.options.ImageMagickBin, "convert")
}

// ImageMagickBlacklist returns the ImageMagick file extension blacklist.
func (c *Config) ImageMagickBlacklist() string {
	return c.options.ImageMagickBlacklist
}

// ImageMagickEnabled checks if ImageMagick can be used for converting media files.
func (c *Config) ImageMagickEnabled() bool {
	return !c.DisableImageMagick()
}
