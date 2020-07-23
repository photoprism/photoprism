package config

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
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

	if err := os.MkdirAll(c.AssetsPath(), os.ModePerm); err != nil {
		return createError(c.AssetsPath(), err)
	}

	if err := os.MkdirAll(c.StoragePath(), os.ModePerm); err != nil {
		return createError(c.StoragePath(), err)
	}

	if err := os.MkdirAll(c.OriginalsPath(), os.ModePerm); err != nil {
		return createError(c.OriginalsPath(), err)
	}

	if err := os.MkdirAll(c.ImportPath(), os.ModePerm); err != nil {
		return createError(c.ImportPath(), err)
	}

	if filepath.IsAbs(c.SidecarPath()) {
		if err := os.MkdirAll(c.SidecarPath(), os.ModePerm); err != nil {
			return createError(c.SidecarPath(), err)
		}
	}

	if err := os.MkdirAll(c.CachePath(), os.ModePerm); err != nil {
		return createError(c.CachePath(), err)
	}

	if err := os.MkdirAll(c.ThumbPath(), os.ModePerm); err != nil {
		return createError(c.ThumbPath(), err)
	}

	if err := os.MkdirAll(c.SettingsPath(), os.ModePerm); err != nil {
		return createError(c.SettingsPath(), err)
	}

	if err := os.MkdirAll(c.TempPath(), os.ModePerm); err != nil {
		return createError(c.TempPath(), err)
	}

	if err := os.MkdirAll(c.TensorFlowModelPath(), os.ModePerm); err != nil {
		return createError(c.TensorFlowModelPath(), err)
	}

	if err := os.MkdirAll(c.BuildPath(), os.ModePerm); err != nil {
		return createError(c.BuildPath(), err)
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
	if c.params.ConfigFile == "" || !fs.FileExists(c.params.ConfigFile) {
		return filepath.Join(c.SettingsPath(), "photoprism.yml")
	}

	return c.params.ConfigFile
}

// SettingsFile returns the user settings file name.
func (c *Config) SettingsFile() string {
	return filepath.Join(c.SettingsPath(), "settings.yml")
}

// SettingsPath returns the config path.
func (c *Config) SettingsPath() string {
	if c.params.SettingsPath == "" {
		return filepath.Join(c.StoragePath(), "settings")
	}

	return fs.Abs(c.params.SettingsPath)
}

// PIDFilename returns the filename for storing the server process id (pid).
func (c *Config) PIDFilename() string {
	if c.params.PIDFilename == "" {
		return filepath.Join(c.StoragePath(), "photoprism.pid")
	}

	return fs.Abs(c.params.PIDFilename)
}

// LogFilename returns the filename for storing server logs.
func (c *Config) LogFilename() string {
	if c.params.LogFilename == "" {
		return filepath.Join(c.StoragePath(), "photoprism.log")
	}

	return fs.Abs(c.params.LogFilename)
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	if c.params.OriginalsPath == "" {
		// Try to find the right directory by iterating through a list.
		c.params.OriginalsPath = fs.FindDir(fs.OriginalPaths)
	}

	return fs.Abs(c.params.OriginalsPath)
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	if c.params.ImportPath == "" {
		// Try to find the right directory by iterating through a list.
		c.params.ImportPath = fs.FindDir(fs.ImportPaths)
	}

	return fs.Abs(c.params.ImportPath)
}

// SipsBin returns the sips executable file name.
func (c *Config) SipsBin() string {
	return findExecutable(c.params.SipsBin, "sips")
}

// ExifToolBin returns the exiftool executable file name.
func (c *Config) ExifToolBin() string {
	return findExecutable(c.params.ExifToolBin, "exiftool")
}

// SidecarJson returns true if metadata should be synced with json sidecar files as used by exiftool.
func (c *Config) SidecarJson() bool {
	if !c.SidecarWritable() || c.ExifToolBin() == "" {
		return false
	}

	return c.params.SidecarJson
}

// SidecarYaml returns true if metadata should be synced with PhotoPrism YAML sidecar files.
func (c *Config) SidecarYaml() bool {
	if !c.SidecarWritable() {
		return false
	}

	return c.params.SidecarYaml
}

// SidecarPath returns the storage path for automatically created sidecar files.
func (c *Config) SidecarPath() string {
	if c.params.SidecarPath == "" {
		c.params.SidecarPath = filepath.Join(c.StoragePath(), "sidecar")
	}

	return c.params.SidecarPath
}

// SidecarPathIsAbs returns true if sidecar path is absolute.
func (c *Config) SidecarPathIsAbs() bool {
	return filepath.IsAbs(c.SidecarPath())
}

// SidecarWritable returns true if sidecar files can be created.
func (c *Config) SidecarWritable() bool {
	return !c.ReadOnly() || c.SidecarPathIsAbs()
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
	if c.params.CachePath == "" {
		return filepath.Join(c.StoragePath(), "cache")
	}

	return fs.Abs(c.params.CachePath)
}

// StoragePath returns the path for generated files like cache and index.
func (c *Config) StoragePath() string {
	if c.params.StoragePath == "" {
		const dirName = "storage"

		// Default directories.
		originalsDir := fs.Abs(filepath.Join(c.OriginalsPath(), fs.HiddenPath, dirName))
		storageDir := fs.Abs(dirName)

		// Find existing directories.
		if fs.PathExists(originalsDir) && !c.ReadOnly() {
			return originalsDir
		} else if fs.PathExists(storageDir) && c.ReadOnly() {
			return storageDir
		}

		// Use .photoprism in home directory?
		if usr, _ := user.Current(); usr.HomeDir != "" {
			p := fs.Abs(filepath.Join(usr.HomeDir, fs.HiddenPath, dirName))

			if fs.PathExists(p) || c.ReadOnly() {
				return p
			}
		}

		// Fallback directory in case nothing else works.
		if c.ReadOnly() {
			return fs.Abs(filepath.Join(fs.HiddenPath, dirName))
		}

		// Store cache and index in "originals/.photoprism/storage".
		return originalsDir
	}

	return fs.Abs(c.params.StoragePath)
}

// AssetsPath returns the path to static assets.
func (c *Config) AssetsPath() string {
	return fs.Abs(c.params.AssetsPath)
}

// LocalesPath returns the translation locales path.
func (c *Config) LocalesPath() string {
	return filepath.Join(c.AssetsPath(), "locales")
}

// ExamplesPath returns the example files path.
func (c *Config) ExamplesPath() string {
	return filepath.Join(c.AssetsPath(), "examples")
}

// TestdataPath returns the test files path.
func (c *Config) TestdataPath() string {
	return filepath.Join(c.StoragePath(), "testdata")
}
