package config

// DisableBackups tests if photo and album metadata backups should be disabled.
func (c *Config) DisableBackups() bool {
	if !c.SidecarWritable() {
		return true
	}

	return c.params.DisableBackups
}

// DisableSettings tests if users should not be allowed to change settings.
func (c *Config) DisableSettings() bool {
	return c.params.DisableSettings
}

// DisablePlaces tests if geocoding and maps should be disabled.
func (c *Config) DisablePlaces() bool {
	return c.params.DisablePlaces
}

// DisableExifTool tests if ExifTool JSON files should not be created for improved metadata extraction.
func (c *Config) DisableExifTool() bool {
	if !c.SidecarWritable() || c.ExifToolBin() == "" {
		return true
	}

	return c.params.DisableExifTool
}

// DisableTensorFlow tests if TensorFlow should not be used for image classification (or anything else).
func (c *Config) DisableTensorFlow() bool {
	return c.params.DisableTensorFlow
}
