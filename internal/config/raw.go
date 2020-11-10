package config

// RawtherapeeBin returns the rawtherapee-cli executable file name.
func (c *Config) RawtherapeeBin() string {
	return findExecutable(c.params.RawtherapeeBin, "rawtherapee-cli")
}

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.params.DarktableBin, "darktable-cli")
}

// DarktablePresets checks if presets should be enabled (disables concurrent raw to jpeg conversion).
func (c *Config) DarktablePresets() bool {
	return c.params.DarktablePresets
}

// SipsBin returns the sips executable file name.
func (c *Config) SipsBin() string {
	return findExecutable(c.params.SipsBin, "sips")
}

// HeifConvertBin returns the heif-convert executable file name.
func (c *Config) HeifConvertBin() string {
	return findExecutable(c.params.HeifConvertBin, "heif-convert")
}
