package config

// RawtherapeeBin returns the rawtherapee-cli executable file name.
func (c *Config) RawtherapeeBin() string {
	return findExecutable(c.options.RawtherapeeBin, "rawtherapee-cli")
}

// RawtherapeeEnabled tests if Rawtherapee is enabled for RAW conversion.
func (c *Config) RawtherapeeEnabled() bool {
	return !c.DisableRawtherapee()
}

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.options.DarktableBin, "darktable-cli")
}

// DarktableEnabled tests if Darktable is enabled for RAW conversion.
func (c *Config) DarktableEnabled() bool {
	return !c.DisableDarktable()
}

// DarktablePresets checks if Darktable presets are enabled (disables concurrent RAW conversion).
func (c *Config) DarktablePresets() bool {
	return c.options.DarktablePresets
}

// SipsEnabled tests if SIPS is enabled for RAW conversion.
func (c *Config) SipsEnabled() bool {
	return !c.DisableSips()
}

// SipsBin returns the SIPS executable file name.
func (c *Config) SipsBin() string {
	return findExecutable(c.options.SipsBin, "sips")
}

// HeifConvertBin returns the heif-convert executable file name.
func (c *Config) HeifConvertBin() string {
	return findExecutable(c.options.HeifConvertBin, "heif-convert")
}

// HeifConvertEnabled tests if heif-convert is enabled for HEIF conversion.
func (c *Config) HeifConvertEnabled() bool {
	return !c.DisableHeifConvert()
}
