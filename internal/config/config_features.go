package config

var Sponsor = Env(EnvDemo, EnvSponsor, EnvTest)

// DisableSettings checks if users should not be allowed to change settings.
func (c *Config) DisableSettings() bool {
	return c.options.DisableSettings
}

// DisableRestart checks if users should not be allowed to restart the server from the user interface.
func (c *Config) DisableRestart() bool {
	return c.options.DisableRestart
}

// DisableWebDAV checks if the built-in WebDAV server should be disabled.
func (c *Config) DisableWebDAV() bool {
	if c.Public() || c.Demo() {
		return true
	}

	return c.options.DisableWebDAV
}

// DisablePlaces checks if geocoding and maps should be disabled.
func (c *Config) DisablePlaces() bool {
	return c.options.DisablePlaces
}

// DisableExifTool checks if ExifTool JSON files should not be created for improved metadata extraction.
func (c *Config) DisableExifTool() bool {
	if c.options.DisableExifTool {
		return true
	} else if !c.SidecarWritable() || c.ExifToolBin() == "" {
		c.options.DisableExifTool = true
	}

	return c.options.DisableExifTool
}

// ExifToolEnabled checks if the use of ExifTool is possible.
func (c *Config) ExifToolEnabled() bool {
	return !c.DisableExifTool()
}

// DisableTensorFlow checks if all features depending on TensorFlow should be disabled.
func (c *Config) DisableTensorFlow() bool {
	if LowMem && !c.options.DisableTensorFlow {
		c.options.DisableTensorFlow = true
	}

	return c.options.DisableTensorFlow
}

// DisableFaces checks if face recognition is disabled.
func (c *Config) DisableFaces() bool {
	if c.DisableTensorFlow() || c.options.DisableFaces {
		return true
	}

	return false
}

// DisableClassification checks if image classification is disabled.
func (c *Config) DisableClassification() bool {
	if c.DisableTensorFlow() || c.options.DisableClassification {
		return true
	}

	return false
}

// DisableFFmpeg checks if FFmpeg is disabled for video transcoding.
func (c *Config) DisableFFmpeg() bool {
	if c.options.DisableFFmpeg {
		return true
	} else if c.FFmpegBin() == "" {
		c.options.DisableFFmpeg = true
	}

	return c.options.DisableFFmpeg
}

// DisableDarktable checks if conversion of RAW images with Darktable is disabled.
func (c *Config) DisableDarktable() bool {
	if c.DisableRaw() || c.options.DisableDarktable {
		return true
	} else if c.DarktableBin() == "" {
		c.options.DisableDarktable = true
	}

	return c.options.DisableDarktable
}

// DisableRawTherapee checks if conversion of RAW images with RawTherapee is disabled.
func (c *Config) DisableRawTherapee() bool {
	if c.DisableRaw() || c.options.DisableRawTherapee {
		return true
	} else if c.RawTherapeeBin() == "" {
		c.options.DisableRawTherapee = true
	}

	return c.options.DisableRawTherapee
}

// DisableImageMagick checks if conversion of files with ImageMagick is disabled.
func (c *Config) DisableImageMagick() bool {
	if c.options.DisableImageMagick {
		return true
	} else if c.ImageMagickBin() == "" {
		c.options.DisableImageMagick = true
	}

	return c.options.DisableImageMagick
}

// DisableHeifConvert checks if heif-convert is disabled for HEIF conversion.
func (c *Config) DisableHeifConvert() bool {
	if c.options.DisableHeifConvert {
		return true
	} else if c.HeifConvertBin() == "" {
		c.options.DisableHeifConvert = true
	}

	return c.options.DisableHeifConvert
}

// DisableVips checks if the use of libvips is disabled.
func (c *Config) DisableVips() bool {
	return c.options.DisableVips
}

// DisableSips checks if conversion of RAW images with SIPS is disabled.
func (c *Config) DisableSips() bool {
	if c.options.DisableSips {
		return true
	} else if c.SipsBin() == "" {
		c.options.DisableSips = true
	}

	return c.options.DisableSips
}

// DisableVectors checks if vector graphics support is disabled.
func (c *Config) DisableVectors() bool {
	if c.options.DisableVectors || !c.Sponsor() {
		return true
	} else if c.RsvgConvertBin() == "" {
		c.options.DisableVectors = true
	}

	return c.options.DisableVectors
}

// DisableRsvgConvert checks if rsvg-convert is disabled for SVG conversion.
func (c *Config) DisableRsvgConvert() bool {
	if c.options.DisableVectors || !c.Sponsor() {
		return true
	}

	return c.RsvgConvertBin() == ""
}

// DisableRaw checks if indexing and conversion of RAW images is disabled.
func (c *Config) DisableRaw() bool {
	if LowMem && !c.options.DisableRaw {
		c.options.DisableRaw = true
		return true
	}

	return c.options.DisableRaw
}
