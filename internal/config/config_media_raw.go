package config

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// RawEnabled checks if indexing and conversion of RAW images is enabled.
func (c *Config) RawEnabled() bool {
	return !c.DisableRaw()
}

// RawPresets checks if RAW converter presents should be used (may reduce performance).
func (c *Config) RawPresets() bool {
	return c.options.RawPresets
}

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findBin(c.options.DarktableBin, "darktable-cli")
}

// DarktableExclude returns the file extensions no not be used with Darktable.
func (c *Config) DarktableExclude() string {
	return c.options.DarktableExclude
}

// DarktableConfigPath returns the darktable config directory.
func (c *Config) DarktableConfigPath() string {
	return fs.Abs(c.options.DarktableConfigPath)
}

// DarktableCachePath returns the darktable cache directory.
func (c *Config) DarktableCachePath() string {
	return fs.Abs(c.options.DarktableCachePath)
}

// CreateDarktableCachePath creates and returns the darktable cache directory.
func (c *Config) CreateDarktableCachePath() (string, error) {
	cachePath := c.DarktableCachePath()

	if cachePath == "" {
		return "", nil
	} else if err := fs.MkdirAll(cachePath); err != nil {
		return cachePath, err
	} else {
		c.options.DarktableCachePath = cachePath
	}

	return cachePath, nil
}

// CreateDarktableConfigPath creates and returns the darktable config directory.
func (c *Config) CreateDarktableConfigPath() (string, error) {
	configPath := c.DarktableConfigPath()

	if configPath == "" {
		return "", nil
	} else if err := fs.MkdirAll(configPath); err != nil {
		return configPath, err
	} else {
		c.options.DarktableConfigPath = configPath
	}

	return configPath, nil
}

// DarktableEnabled checks if Darktable is enabled for RAW conversion.
func (c *Config) DarktableEnabled() bool {
	return !c.DisableDarktable()
}

// RawTherapeeBin returns the rawtherapee-cli executable file name.
func (c *Config) RawTherapeeBin() string {
	return findBin(c.options.RawTherapeeBin, "rawtherapee-cli")
}

// RawTherapeeExclude returns the file extensions no not be used with RawTherapee.
func (c *Config) RawTherapeeExclude() string {
	return c.options.RawTherapeeExclude
}

// RawTherapeeEnabled checks if RawTherapee is enabled for RAW conversion.
func (c *Config) RawTherapeeEnabled() bool {
	return !c.DisableRawTherapee()
}
