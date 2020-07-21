package config

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.params.DarktableBin, "darktable-cli")
}

// DarktablePresets checks if presets should be enabled (disables concurrent raw to jpeg conversion).
func (c *Config) DarktablePresets() bool {
	return c.params.DarktablePresets
}
