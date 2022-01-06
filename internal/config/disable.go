package config

// DisableWebDAV tests if the built-in WebDAV server should be disabled.
func (c *Config) DisableWebDAV() bool {
	if c.ReadOnly() || c.Demo() {
		return true
	}

	return c.options.DisableWebDAV
}

// DisableBackups tests if photo and album metadata files should be disabled.
func (c *Config) DisableBackups() bool {
	if !c.SidecarWritable() {
		return true
	}

	return c.options.DisableBackups
}

// DisableSettings tests if users should not be allowed to change settings.
func (c *Config) DisableSettings() bool {
	return c.options.DisableSettings
}

// DisablePlaces tests if geocoding and maps should be disabled.
func (c *Config) DisablePlaces() bool {
	return c.options.DisablePlaces
}

// DisableExifTool tests if ExifTool JSON files should not be created for improved metadata extraction.
func (c *Config) DisableExifTool() bool {
	if !c.SidecarWritable() || c.ExifToolBin() == "" {
		return true
	}

	return c.options.DisableExifTool
}

// DisableTensorFlow tests if all features depending on TensorFlow should be disabled.
func (c *Config) DisableTensorFlow() bool {
	if LowMem && !c.options.DisableTensorFlow {
		c.options.DisableTensorFlow = true
		log.Warnf("config: disabled tensorflow due to memory constraints")
	}

	return c.options.DisableTensorFlow
}

// DisableFaces tests if facial recognition is disabled.
func (c *Config) DisableFaces() bool {
	if c.DisableTensorFlow() || c.options.DisableFaces {
		return true
	}

	return false
}

// DisableClassification tests if image classification is disabled.
func (c *Config) DisableClassification() bool {
	if c.DisableTensorFlow() || c.options.DisableClassification {
		return true
	}

	return false
}

// DisableFFmpeg tests if FFmpeg is disabled for video transcoding.
func (c *Config) DisableFFmpeg() bool {
	return c.options.DisableFFmpeg || c.FFmpegBin() == ""
}

// DisableDarktable tests if Darktable is disabled for RAW conversion.
func (c *Config) DisableDarktable() bool {
	if LowMem && !c.options.DisableDarktable {
		c.options.DisableDarktable = true
		log.Warnf("config: disabled file conversion with Darktable due to memory constraints")
	}

	return c.options.DisableDarktable || c.DarktableBin() == ""
}

// DisableRawtherapee tests if Rawtherapee is disabled for RAW conversion.
func (c *Config) DisableRawtherapee() bool {
	if LowMem && !c.options.DisableRawtherapee {
		c.options.DisableRawtherapee = true
		log.Warnf("config: disabled file conversion with RawTherapee due to memory constraints")
	}

	return c.options.DisableRawtherapee || c.RawtherapeeBin() == ""
}

// DisableSips tests if SIPS is disabled for RAW conversion.
func (c *Config) DisableSips() bool {
	return c.options.DisableSips || c.SipsBin() == ""
}

// DisableHeifConvert tests if heif-convert is disabled for HEIF conversion.
func (c *Config) DisableHeifConvert() bool {
	return c.options.DisableHeifConvert || c.HeifConvertBin() == ""
}
