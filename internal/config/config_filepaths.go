package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// binPaths stores known executable paths.
var (
	binPaths = make(map[string]string, 8)
	binMu    = sync.RWMutex{}
	tempPath = ""
)

// findBin resolves the absolute file path of external binaries.
func findBin(configBin, defaultBin string) (binPath string) {
	cacheKey := defaultBin + configBin
	binMu.RLock()
	cached, found := binPaths[cacheKey]
	binMu.RUnlock()

	// Already found?
	if found {
		return cached
	}

	// Default binary name?
	if configBin == "" {
		binPath = defaultBin
	} else {
		binPath = configBin
	}

	// Search for binary.
	if path, err := exec.LookPath(binPath); err == nil {
		binPath = path
	}

	// Found?
	if !fs.FileExists(binPath) {
		binPath = ""
	} else {
		binMu.Lock()
		binPaths[cacheKey] = binPath
		binMu.Unlock()
	}

	return binPath
}

// CreateDirectories creates directories for storing photos, metadata and cache files.
func (c *Config) CreateDirectories() error {
	if c.AssetsPath() == "" {
		return notFoundError("assets")
	} else if err := os.MkdirAll(c.AssetsPath(), fs.ModeDir); err != nil {
		return createError(c.AssetsPath(), err)
	}

	if c.StoragePath() == "" {
		return notFoundError("storage")
	} else if err := os.MkdirAll(c.StoragePath(), fs.ModeDir); err != nil {
		return createError(c.StoragePath(), err)
	}

	if c.UsersPath() == "" {
		return notFoundError("users")
	} else if err := os.MkdirAll(c.UsersStoragePath(), fs.ModeDir); err != nil {
		return createError(c.UsersStoragePath(), err)
	}

	if c.CmdCachePath() == "" {
		return notFoundError("cmd cache")
	} else if err := os.MkdirAll(c.CmdCachePath(), fs.ModeDir); err != nil {
		return createError(c.CmdCachePath(), err)
	}

	if c.BackupPath() == "" {
		return notFoundError("backup")
	} else if err := os.MkdirAll(c.BackupPath(), fs.ModeDir); err != nil {
		return createError(c.BackupPath(), err)
	}

	if c.OriginalsPath() == "" {
		return notFoundError("originals")
	} else if err := os.MkdirAll(c.OriginalsPath(), fs.ModeDir); err != nil {
		return createError(c.OriginalsPath(), err)
	}

	if c.ImportPath() == "" {
		return notFoundError("import")
	} else if err := os.MkdirAll(c.ImportPath(), fs.ModeDir); err != nil {
		return createError(c.ImportPath(), err)
	}

	if filepath.IsAbs(c.SidecarPath()) {
		if err := os.MkdirAll(c.SidecarPath(), fs.ModeDir); err != nil {
			return createError(c.SidecarPath(), err)
		}
	}

	if c.CachePath() == "" {
		return notFoundError("cache")
	} else if err := os.MkdirAll(c.CachePath(), fs.ModeDir); err != nil {
		return createError(c.CachePath(), err)
	}

	if c.MediaCachePath() == "" {
		return notFoundError("media")
	} else if err := os.MkdirAll(c.MediaCachePath(), fs.ModeDir); err != nil {
		return createError(c.MediaCachePath(), err)
	}

	if c.ThumbCachePath() == "" {
		return notFoundError("thumbs")
	} else if err := os.MkdirAll(c.ThumbCachePath(), fs.ModeDir); err != nil {
		return createError(c.ThumbCachePath(), err)
	}

	if c.ConfigPath() == "" {
		return notFoundError("config")
	} else if err := os.MkdirAll(c.ConfigPath(), fs.ModeDir); err != nil {
		return createError(c.ConfigPath(), err)
	}

	if c.CertificatesPath() == "" {
		return notFoundError("certificates")
	} else if err := os.MkdirAll(c.CertificatesPath(), fs.ModeDir); err != nil {
		return createError(c.CertificatesPath(), err)
	}

	if c.TempPath() == "" {
		return notFoundError("temp")
	} else if err := os.MkdirAll(c.TempPath(), fs.ModeDir); err != nil {
		return createError(c.TempPath(), err)
	}

	if c.AlbumsPath() == "" {
		return notFoundError("albums")
	} else if err := os.MkdirAll(c.AlbumsPath(), fs.ModeDir); err != nil {
		return createError(c.AlbumsPath(), err)
	}

	if c.TensorFlowModelPath() == "" {
		return notFoundError("tensorflow model")
	} else if err := os.MkdirAll(c.TensorFlowModelPath(), fs.ModeDir); err != nil {
		return createError(c.TensorFlowModelPath(), err)
	}

	if c.BuildPath() == "" {
		return notFoundError("build")
	} else if err := os.MkdirAll(c.BuildPath(), fs.ModeDir); err != nil {
		return createError(c.BuildPath(), err)
	}

	if filepath.Dir(c.PIDFilename()) == "" {
		return notFoundError("pid file")
	} else if err := os.MkdirAll(filepath.Dir(c.PIDFilename()), fs.ModeDir); err != nil {
		return createError(filepath.Dir(c.PIDFilename()), err)
	}

	if filepath.Dir(c.LogFilename()) == "" {
		return notFoundError("log file")
	} else if err := os.MkdirAll(filepath.Dir(c.LogFilename()), fs.ModeDir); err != nil {
		return createError(filepath.Dir(c.LogFilename()), err)
	}

	if c.DarktableEnabled() {
		if dir, err := c.CreateDarktableCachePath(); err != nil {
			return fmt.Errorf("could not create darktable cache path %s", clean.Log(dir))
		}

		if dir, err := c.CreateDarktableConfigPath(); err != nil {
			return fmt.Errorf("could not create darktable cache path %s", clean.Log(dir))
		}
	}

	return nil
}

// ConfigPath returns the config path.
func (c *Config) ConfigPath() string {
	if c.options.ConfigPath == "" {
		if fs.PathExists(filepath.Join(c.StoragePath(), "settings")) {
			return filepath.Join(c.StoragePath(), "settings")
		}

		return filepath.Join(c.StoragePath(), "config")
	}

	return fs.Abs(c.options.ConfigPath)
}

// OptionsYaml returns the config options YAML filename.
func (c *Config) OptionsYaml() string {
	return filepath.Join(c.ConfigPath(), "options.yml")
}

// DefaultsYaml returns the default options YAML filename.
func (c *Config) DefaultsYaml() string {
	return fs.Abs(c.options.DefaultsYaml)
}

// HubConfigFile returns the backend api config file name.
func (c *Config) HubConfigFile() string {
	return filepath.Join(c.ConfigPath(), "hub.yml")
}

// SettingsYaml returns the settings YAML filename.
func (c *Config) SettingsYaml() string {
	return filepath.Join(c.ConfigPath(), "settings.yml")
}

// SettingsYamlDefaults returns the default settings YAML filename.
func (c *Config) SettingsYamlDefaults(settingsYml string) string {
	if settingsYml != "" && fs.FileExists(settingsYml) {
		// Use regular settings YAML file.
	} else if defaultsYml := c.DefaultsYaml(); defaultsYml == "" {
		// Use regular settings YAML file.
	} else if dir := filepath.Dir(defaultsYml); dir == "" || dir == "." {
		// Use regular settings YAML file.
	} else if fileName := filepath.Join(dir, "settings.yml"); settingsYml == "" || fs.FileExistsNotEmpty(fileName) {
		// Use default settings YAML file.
		return fileName
	}

	return settingsYml
}

// PIDFilename returns the filename for storing the server process id (pid).
func (c *Config) PIDFilename() string {
	if c.options.PIDFilename == "" {
		return filepath.Join(c.StoragePath(), "photoprism.pid")
	}

	return fs.Abs(c.options.PIDFilename)
}

// LogFilename returns the filename for storing server logs.
func (c *Config) LogFilename() string {
	if c.options.LogFilename == "" {
		return filepath.Join(c.StoragePath(), "photoprism.log")
	}

	return fs.Abs(c.options.LogFilename)
}

// CaseInsensitive checks if the storage path is case-insensitive.
func (c *Config) CaseInsensitive() (result bool, err error) {
	storagePath := c.StoragePath()
	return fs.CaseInsensitive(storagePath)
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	if c.options.OriginalsPath == "" {
		// Try to find the right directory by iterating through a list.
		c.options.OriginalsPath = fs.FindDir(fs.OriginalPaths)
	}

	return fs.Abs(c.options.OriginalsPath)
}

// OriginalsDeletable checks if originals can be deleted.
func (c *Config) OriginalsDeletable() bool {
	return !c.ReadOnly() && fs.Writable(c.OriginalsPath()) && c.Settings().Features.Delete
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	if c.options.ImportPath == "" {
		// Try to find the right directory by iterating through a list.
		c.options.ImportPath = fs.FindDir(fs.ImportPaths)
	}

	return fs.Abs(c.options.ImportPath)
}

// ImportDest returns the relative originals path to which the files should be imported by default.
func (c *Config) ImportDest() string {
	return clean.UserPath(c.options.ImportDest)
}

// SidecarPath returns the storage path for generated sidecar files (relative or absolute).
func (c *Config) SidecarPath() string {
	if c.options.SidecarPath == "" {
		c.options.SidecarPath = filepath.Join(c.StoragePath(), "sidecar")
	}

	return c.options.SidecarPath
}

// SidecarPathIsAbs checks if sidecar path is absolute.
func (c *Config) SidecarPathIsAbs() bool {
	return filepath.IsAbs(c.SidecarPath())
}

// SidecarWritable checks if sidecar files can be created.
func (c *Config) SidecarWritable() bool {
	return !c.ReadOnly() || c.SidecarPathIsAbs()
}

// UsersPath returns the relative base path for user assets.
func (c *Config) UsersPath() string {
	// Set default.
	if c.options.UsersPath == "" {
		return "users"
	}

	return clean.UserPath(c.options.UsersPath)
}

// UsersOriginalsPath returns the users originals base path.
func (c *Config) UsersOriginalsPath() string {
	return filepath.Join(c.OriginalsPath(), c.UsersPath())
}

// UsersStoragePath returns the users storage base path.
func (c *Config) UsersStoragePath() string {
	return filepath.Join(c.StoragePath(), "users")
}

// UserStoragePath returns the storage path for user assets.
func (c *Config) UserStoragePath(userUid string) string {
	if !rnd.IsUID(userUid, 0) {
		return ""
	}

	dir := filepath.Join(c.UsersStoragePath(), userUid)

	if err := os.MkdirAll(dir, fs.ModeDir); err != nil {
		return ""
	}

	return dir
}

// UserUploadPath returns the upload path for the specified user.
func (c *Config) UserUploadPath(userUid, token string) (string, error) {
	if !rnd.IsUID(userUid, 0) {
		return "", fmt.Errorf("invalid uid")
	}

	dir := filepath.Join(c.UserStoragePath(userUid), "upload", clean.Token(token))

	if err := os.MkdirAll(dir, fs.ModeDir); err != nil {
		return "", err
	}

	return dir, nil
}

// TempPath returns the cached temporary directory name e.g. for uploads and downloads.
func (c *Config) TempPath() string {
	// Return cached value?
	if tempPath == "" {
		tempPath = c.tempPath()
	}

	return tempPath
}

// tempPath determines the temporary directory name e.g. for uploads and downloads.
func (c *Config) tempPath() string {
	osTempDir := os.TempDir()

	// Empty default?
	if osTempDir == "" {
		switch runtime.GOOS {
		case "android":
			osTempDir = "/data/local/tmp"
		case "windows":
			osTempDir = "C:/Windows/Temp"
		default:
			osTempDir = "/tmp"
		}

		log.Infof("config: empty default temp folder path, using %s", clean.Log(osTempDir))
	}

	// Check configured temp path first.
	if c.options.TempPath != "" {
		if dir := fs.Abs(c.options.TempPath); dir == "" {
			// Ignore.
		} else if err := os.MkdirAll(dir, fs.ModeDir); err != nil {
			// Ignore.
		} else if fs.PathWritable(dir) {
			return dir
		}
	}

	// Find alternative temp path based on storage serial checksum.
	if dir := filepath.Join(osTempDir, "photoprism_"+c.SerialChecksum()); dir == "" {
		// Ignore.
	} else if err := os.MkdirAll(dir, fs.ModeDir); err != nil {
		// Ignore.
	} else if fs.PathWritable(dir) {
		return dir
	}

	// Find alternative temp path based on built-in TempDir() function.
	if dir, err := ioutil.TempDir(osTempDir, "photoprism_"); err != nil || dir == "" {
		// Ignore.
	} else if err = os.MkdirAll(dir, fs.ModeDir); err != nil {
		// Ignore.
	} else if fs.PathWritable(dir) {
		return dir
	}

	return osTempDir
}

// CachePath returns the path for cache files.
func (c *Config) CachePath() string {
	if c.options.CachePath == "" {
		return filepath.Join(c.StoragePath(), "cache")
	}

	return fs.Abs(c.options.CachePath)
}

// CmdCachePath returns a path that external CLI tools can use as cache directory.
func (c *Config) CmdCachePath() string {
	return filepath.Join(c.CachePath(), "cmd")
}

// CmdLibPath returns the dynamic loader path that external CLI tools should use.
func (c *Config) CmdLibPath() string {
	if dir := os.Getenv("LD_LIBRARY_PATH"); dir != "" {
		return dir
	}

	return "/usr/local/lib:/usr/lib"
}

// MediaCachePath returns the media cache path.
func (c *Config) MediaCachePath() string {
	return filepath.Join(c.CachePath(), "media")
}

// ThumbCachePath returns the thumbnail storage path.
func (c *Config) ThumbCachePath() string {
	return filepath.Join(c.CachePath(), "thumbnails")
}

// StoragePath returns the path for generated files like cache and index.
func (c *Config) StoragePath() string {
	if c.options.StoragePath == "" {
		const dirName = "storage"

		// Default directories.
		originalsDir := fs.Abs(filepath.Join(c.OriginalsPath(), fs.HiddenPath, dirName))
		storageDir := fs.Abs(dirName)

		// Find existing directories.
		if fs.PathWritable(originalsDir) && !c.ReadOnly() {
			return originalsDir
		} else if fs.PathWritable(storageDir) && c.ReadOnly() {
			return storageDir
		}

		// Fallback to backup storage path.
		if fs.PathWritable(c.options.BackupPath) {
			return fs.Abs(filepath.Join(c.options.BackupPath, dirName))
		}

		// Use .photoprism in home directory?
		if usr, _ := user.Current(); usr.HomeDir != "" {
			p := fs.Abs(filepath.Join(usr.HomeDir, fs.HiddenPath, dirName))

			if fs.PathWritable(p) || c.ReadOnly() {
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

	return fs.Abs(c.options.StoragePath)
}

// BackupPath returns the backup storage path.
func (c *Config) BackupPath() string {
	if fs.PathWritable(c.options.BackupPath) {
		return fs.Abs(c.options.BackupPath)
	}

	return filepath.Join(c.StoragePath(), "backup")
}

// AssetsPath returns the path to static assets for models and templates.
func (c *Config) AssetsPath() string {
	if c.options.AssetsPath == "" {
		// Try to find the right directory by iterating through a list.
		c.options.AssetsPath = fs.FindDir(fs.AssetPaths)
	}

	return fs.Abs(c.options.AssetsPath)
}

// CustomAssetsPath returns the path to custom assets such as icons, models and translations.
func (c *Config) CustomAssetsPath() string {
	if c.options.CustomAssetsPath != "" {
		return fs.Abs(c.options.CustomAssetsPath)
	}

	return ""
}

// CustomStaticPath returns the custom static assets' path.
func (c *Config) CustomStaticPath() string {
	if dir := c.CustomAssetsPath(); dir == "" {
		return ""
	} else if dir = filepath.Join(dir, "static"); !fs.PathExists(dir) {
		return ""
	} else {
		return dir
	}
}

// CustomStaticFile returns the path to a custom static file.
func (c *Config) CustomStaticFile(fileName string) string {
	if dir := c.CustomStaticPath(); dir == "" {
		return ""
	} else {
		return filepath.Join(dir, fileName)
	}
}

// CustomStaticUri returns the URI to a custom static resource.
func (c *Config) CustomStaticUri() string {
	if dir := c.CustomAssetsPath(); dir == "" {
		return ""
	} else {
		return c.CdnUrl(c.BaseUri(CustomStaticUri))
	}
}

// CustomStaticAssetUri returns the resource URI of the custom static file asset.
func (c *Config) CustomStaticAssetUri(res string) string {
	if dir := c.CustomAssetsPath(); dir == "" {
		return ""
	} else {
		return c.CdnUrl(c.BaseUri(CustomStaticUri)) + "/" + res
	}
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

// MariadbBin returns the mariadb executable file name.
func (c *Config) MariadbBin() string {
	return findBin("", "mariadb")
}

// MariadbDumpBin returns the mariadb-dump executable file name.
func (c *Config) MariadbDumpBin() string {
	return findBin("", "mariadb-dump")
}

// SqliteBin returns the sqlite executable file name.
func (c *Config) SqliteBin() string {
	return findBin("", "sqlite3")
}

// AlbumsPath returns the storage path for album YAML files.
func (c *Config) AlbumsPath() string {
	return filepath.Join(c.StoragePath(), "albums")
}

// OriginalsAlbumsPath returns the optional album YAML file path inside originals.
func (c *Config) OriginalsAlbumsPath() string {
	return filepath.Join(c.OriginalsPath(), "albums")
}
