package config

var Sponsor = Env(EnvDemo, EnvSponsor, EnvTest)

// DisableWebDAV checks if the built-in WebDAV server should be disabled.
func (c *Config) DisableWebDAV() bool {
	if c.Public() || c.ReadOnly() || c.Demo() {
		return true
	}

	return c.options.DisableWebDAV
}

// DisableBackups checks if photo and album metadata files should be disabled.
func (c *Config) DisableBackups() bool {
	if !c.SidecarWritable() {
		return true
	}

	return c.options.DisableBackups
}

// DisableSettings checks if users should not be allowed to change settings.
func (c *Config) DisableSettings() bool {
	return c.options.DisableSettings
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

// DisableRaw checks if indexing and conversion of RAW files is disabled.
func (c *Config) DisableRaw() bool {
	if LowMem && !c.options.DisableRaw {
		c.options.DisableRaw = true
		return true
	}

	return c.options.DisableRaw
}

// DisableDarktable checks if conversion of RAW files with Darktable is disabled.
func (c *Config) DisableDarktable() bool {
	if c.DisableRaw() || c.options.DisableDarktable {
		return true
	} else if c.DarktableBin() == "" {
		c.options.DisableDarktable = true
	}

	return c.options.DisableDarktable
}

// DisableRawtherapee checks if conversion of RAW files with Rawtherapee is disabled.
func (c *Config) DisableRawtherapee() bool {
	if c.DisableRaw() || c.options.DisableRawtherapee {
		return true
	} else if c.RawtherapeeBin() == "" {
		c.options.DisableRawtherapee = true
	}

	return c.options.DisableRawtherapee
}

// DisableSips checks if conversion of RAW files with SIPS is disabled.
func (c *Config) DisableSips() bool {
	if c.options.DisableSips {
		return true
	} else if c.SipsBin() == "" {
		c.options.DisableSips = true
	}

	return c.options.DisableSips
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
