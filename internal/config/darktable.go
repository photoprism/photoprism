package config

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.params.DarktableBin, "darktable-cli")
}

// DarktableUnlock checks if presets should be disabled to run multiple instances concurrently.
func (c *Config) DarktableUnlock() bool {
	return c.params.DarktableUnlock
}
