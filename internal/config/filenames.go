package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

func findExecutable(configBin, defaultBin string) (result string) {
	if configBin == "" {
		result = defaultBin
	} else {
		result = configBin
	}

	if path, err := exec.LookPath(result); err == nil {
		result = path
	}

	if !fs.FileExists(result) {
		result = ""
	}

	return result
}

// CreateDirectories creates directories for storing photos, metadata and cache files.
func (c *Config) CreateDirectories() error {
	createError := func(path string, err error) (result error) {
		if fs.FileExists(path) {
			result = fmt.Errorf("%s is a file, not a folder: please check your configuration", txt.Quote(path))
		} else {
			result = fmt.Errorf("can't create %s: please check configuration and permissions", txt.Quote(path))
		}

		log.Debug(err)

		return result
	}

	if err := os.MkdirAll(c.OriginalsPath(), os.ModePerm); err != nil {
		return createError(c.OriginalsPath(), err)
	}

	if err := os.MkdirAll(c.ImportPath(), os.ModePerm); err != nil {
		return createError(c.ImportPath(), err)
	}

	if err := os.MkdirAll(c.TempPath(), os.ModePerm); err != nil {
		return createError(c.TempPath(), err)
	}

	if err := os.MkdirAll(c.ThumbPath(), os.ModePerm); err != nil {
		return createError(c.ThumbPath(), err)
	}

	if err := os.MkdirAll(c.ResourcesPath(), os.ModePerm); err != nil {
		return createError(c.ResourcesPath(), err)
	}

	if err := os.MkdirAll(c.TidbServerPath(), os.ModePerm); err != nil {
		return createError(c.TidbServerPath(), err)
	}

	if err := os.MkdirAll(c.TensorFlowModelPath(), os.ModePerm); err != nil {
		return createError(c.TensorFlowModelPath(), err)
	}

	if err := os.MkdirAll(c.HttpStaticBuildPath(), os.ModePerm); err != nil {
		return createError(c.HttpStaticBuildPath(), err)
	}

	if err := os.MkdirAll(filepath.Dir(c.PIDFilename()), os.ModePerm); err != nil {
		return createError(filepath.Dir(c.PIDFilename()), err)
	}

	if err := os.MkdirAll(filepath.Dir(c.LogFilename()), os.ModePerm); err != nil {
		return createError(filepath.Dir(c.LogFilename()), err)
	}

	return nil
}

// ConfigFile returns the config file name.
func (c *Config) ConfigFile() string {
	return c.params.ConfigFile
}

// SettingsFile returns the user settings file name.
func (c *Config) SettingsFile() string {
	return filepath.Join(c.ConfigPath(), "settings.yml")
}

// ConfigPath returns the config path.
func (c *Config) ConfigPath() string {
	if c.params.ConfigPath == "" {
		return filepath.Join(c.AssetsPath(), "config")
	}

	return fs.Abs(c.params.ConfigPath)
}

// PIDFilename returns the filename for storing the server process id (pid).
func (c *Config) PIDFilename() string {
	if c.params.PIDFilename == "" {
		return filepath.Join(c.AssetsPath(), "photoprism.pid")
	}

	return fs.Abs(c.params.PIDFilename)
}

// LogFilename returns the filename for storing server logs.
func (c *Config) LogFilename() string {
	if c.params.LogFilename == "" {
		return filepath.Join(c.AssetsPath(), "photoprism.log")
	}

	return fs.Abs(c.params.LogFilename)
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	return fs.Abs(c.params.OriginalsPath)
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	return fs.Abs(c.params.ImportPath)
}

// SipsBin returns the sips executable file name.
func (c *Config) SipsBin() string {
	return findExecutable(c.params.SipsBin, "sips")
}

// DarktableBin returns the darktable-cli executable file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.params.DarktableBin, "darktable-cli")
}

// ExifToolBin returns the exiftool executable file name.
func (c *Config) ExifToolBin() string {
	return findExecutable(c.params.ExifToolBin, "exiftool")
}

// SidecarJson returns true if metadata should be synced with json sidecar files as used by exiftool.
func (c *Config) SidecarJson() bool {
	if c.ReadOnly() || c.ExifToolBin() == "" {
		return false
	}

	return c.params.SidecarJson
}

// HeifConvertBin returns the heif-convert executable file name.
func (c *Config) HeifConvertBin() string {
	return findExecutable(c.params.HeifConvertBin, "heif-convert")
}

// FFmpegBin returns the ffmpeg executable file name.
func (c *Config) FFmpegBin() string {
	return findExecutable(c.params.FFmpegBin, "ffmpeg")
}

// TempPath returns a temporary directory name for uploads and downloads.
func (c *Config) TempPath() string {
	if c.params.TempPath == "" {
		return filepath.Join(os.TempDir(), "photoprism")
	}

	return fs.Abs(c.params.TempPath)
}

// CachePath returns the path to the cache.
func (c *Config) CachePath() string {
	return fs.Abs(c.params.CachePath)
}

// AssetsPath returns the path to the assets.
func (c *Config) AssetsPath() string {
	return fs.Abs(c.params.AssetsPath)
}

// ResourcesPath returns the path to the app resources like static files.
func (c *Config) ResourcesPath() string {
	if c.params.ResourcesPath == "" {
		return filepath.Join(c.AssetsPath(), "resources")
	}

	return fs.Abs(c.params.ResourcesPath)
}

// ExamplesPath returns the example files path.
func (c *Config) ExamplesPath() string {
	return filepath.Join(c.ResourcesPath(), "examples")
}

// TensorFlowModelPath returns the tensorflow model path.
func (c *Config) TensorFlowModelPath() string {
	return filepath.Join(c.ResourcesPath(), "nasnet")
}

// NSFWModelPath returns the NSFW tensorflow model path.
func (c *Config) NSFWModelPath() string {
	return filepath.Join(c.ResourcesPath(), "nsfw")
}
