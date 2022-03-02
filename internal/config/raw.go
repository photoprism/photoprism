package config

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/sanitize"
)

// RawPresets tests if RAW converter presents should be used (may reduce performance).
func (c *Config) RawPresets() bool {
	return c.options.RawPresets
}

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.options.DarktableBin, "darktable-cli")
}

// DarktableBlacklist returns the darktable file extension blacklist.
func (c *Config) DarktableBlacklist() string {
	return c.options.DarktableBlacklist
}

// DarktableConfigPath returns the darktable config directory.
func (c *Config) DarktableConfigPath() string {
	dir := filepath.Join(c.ConfigPath(), "darktable")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Errorf("darktable: cannot create config directory %s, check permissions", sanitize.Log(dir))
		return c.ConfigPath()
	}
	return dir
}

// DarktableCachePath returns the darktable cache directory.
func (c *Config) DarktableCachePath() string {
	dir := filepath.Join(c.CachePath(), "darktable")

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Errorf("darktable: cannot create cache directory %s, check permissions", sanitize.Log(dir))
		return c.ConfigPath()
	}

	return dir
}

// DarktableEnabled tests if Darktable is enabled for RAW conversion.
func (c *Config) DarktableEnabled() bool {
	return !c.DisableDarktable()
}

// RawtherapeeBin returns the rawtherapee-cli executable file name.
func (c *Config) RawtherapeeBin() string {
	return findExecutable(c.options.RawtherapeeBin, "rawtherapee-cli")
}

// RawtherapeeBlacklist returns the RawTherapee file extension blacklist.
func (c *Config) RawtherapeeBlacklist() string {
	return c.options.RawtherapeeBlacklist
}

// RawtherapeeEnabled tests if Rawtherapee is enabled for RAW conversion.
func (c *Config) RawtherapeeEnabled() bool {
	return !c.DisableRawtherapee()
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
