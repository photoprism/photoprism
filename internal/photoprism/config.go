package photoprism

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/urfave/cli"
)

// Config provides a struct in which application configuration is stored.
// Application code must use functions to get config values, for two reasons:
//
// 1. Some values are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Config struct {
	appName        string
	appVersion     string
	appCopyright   string
	debug          bool
	configFile     string
	serverIP       string
	serverPort     int
	serverMode     string
	assetsPath     string
	cachePath      string
	originalsPath  string
	importPath     string
	exportPath     string
	darktableCli   string
	databaseDriver string
	databaseDsn    string
	db             *gorm.DB
}

type configValues map[string]interface{}

// NewConfig() creates a new configuration entity by using two methods:
//
// 1. SetValuesFromFile: This will initialize values from a yaml config file.
//
// 2. SetValuesFromCliContext: Which comes after SetValuesFromFile and overrides
//    any previous values giving an option two override file configs through the CLI.
func NewConfig(context *cli.Context) *Config {
	c := &Config{}
	c.appName = context.App.Name
	c.appCopyright = context.App.Copyright
	c.appVersion = context.App.Version
	c.SetValuesFromFile(GetExpandedFilename(context.GlobalString("config-file")))
	c.SetValuesFromCliContext(context)

	return c
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (c *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	c.configFile = fileName
	if debug, err := yamlConfig.GetBool("debug"); err == nil {
		c.debug = debug
	}

	if serverIP, err := yamlConfig.Get("server-host"); err == nil {
		c.serverIP = serverIP
	}

	if serverPort, err := yamlConfig.GetInt("server-port"); err == nil {
		c.serverPort = int(serverPort)
	}

	if serverMode, err := yamlConfig.Get("server-mode"); err == nil {
		c.serverMode = serverMode
	}

	if assetsPath, err := yamlConfig.Get("assets-path"); err == nil {
		c.assetsPath = GetExpandedFilename(assetsPath)
	}

	if cachePath, err := yamlConfig.Get("cache-path"); err == nil {
		c.cachePath = GetExpandedFilename(cachePath)
	}

	if originalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.originalsPath = GetExpandedFilename(originalsPath)
	}

	if importPath, err := yamlConfig.Get("import-path"); err == nil {
		c.importPath = GetExpandedFilename(importPath)
	}

	if exportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.exportPath = GetExpandedFilename(exportPath)
	}

	if darktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.darktableCli = GetExpandedFilename(darktableCli)
	}

	if databaseDriver, err := yamlConfig.Get("database-driver"); err == nil {
		c.databaseDriver = databaseDriver
	}

	if databaseDsn, err := yamlConfig.Get("database-dsn"); err == nil {
		c.databaseDsn = databaseDsn
	}

	return nil
}

// SetValuesFromCliContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Config) SetValuesFromCliContext(context *cli.Context) error {
	if context.GlobalBool("debug") {
		c.debug = context.GlobalBool("debug")
	}

	if context.GlobalIsSet("assets-path") || c.assetsPath == "" {
		c.assetsPath = GetExpandedFilename(context.GlobalString("assets-path"))
	}

	if context.GlobalIsSet("cache-path") || c.cachePath == "" {
		c.cachePath = GetExpandedFilename(context.GlobalString("cache-path"))
	}

	if context.GlobalIsSet("originals-path") || c.originalsPath == "" {
		c.originalsPath = GetExpandedFilename(context.GlobalString("originals-path"))
	}

	if context.GlobalIsSet("import-path") || c.importPath == "" {
		c.importPath = GetExpandedFilename(context.GlobalString("import-path"))
	}

	if context.GlobalIsSet("export-path") || c.exportPath == "" {
		c.exportPath = GetExpandedFilename(context.GlobalString("export-path"))
	}

	if context.GlobalIsSet("darktable-cli") || c.darktableCli == "" {
		c.darktableCli = GetExpandedFilename(context.GlobalString("darktable-cli"))
	}

	if context.GlobalIsSet("database-driver") || c.databaseDriver == "" {
		c.databaseDriver = context.GlobalString("database-driver")
	}

	if context.GlobalIsSet("database-dsn") || c.databaseDsn == "" {
		c.databaseDsn = context.GlobalString("database-dsn")
	}

	if context.IsSet("server-host") || c.serverIP == "" {
		c.serverIP = context.String("server-host")
	}

	if context.IsSet("server-port") || c.serverPort == 0 {
		c.serverPort = context.Int("server-port")
	}

	if context.IsSet("server-mode") || c.serverMode == "" {
		c.serverMode = context.String("server-mode")
	}

	return nil
}

// CreateDirectories creates all the folders that photoprism needs. These are:
// originalsPath
// ThumbnailsPath
// importPath
// exportPath
func (c *Config) CreateDirectories() error {
	if err := os.MkdirAll(c.GetOriginalsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetImportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetExportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetThumbnailsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetDatabasePath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetTensorFlowModelPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.GetPublicBuildPath(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// connectToDatabase estabilishes a connection to a database given a driver.
// It tries to do this 12 times with a 5 second sleep intervall in between.
func (c *Config) connectToDatabase() error {
	db, err := gorm.Open(c.databaseDriver, c.databaseDsn)

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(c.databaseDriver, c.databaseDsn)

			if db != nil && err == nil {
				break
			}
		}

		if err != nil || db == nil {
			log.Fatal(err)
		}
	}

	c.db = db

	return err
}

// GetAppName returns the application name.
func (c *Config) GetAppName() string {
	return c.appName
}

// GetAppVersion returns the application version.
func (c *Config) GetAppVersion() string {
	return c.appVersion
}

// GetAppCopyright returns the application copyright.
func (c *Config) GetAppCopyright() string {
	return c.appCopyright
}

// IsDebug returns true if debug mode is on.
func (c *Config) IsDebug() bool {
	return c.debug
}

// GetConfigFile returns the config file name.
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// GetServerIP returns the server IP address (empty for all).
func (c *Config) GetServerIP() string {
	return c.serverIP
}

// GetServerPort returns the server port.
func (c *Config) GetServerPort() int {
	return c.serverPort
}

// GetServerMode returns the server mode.
func (c *Config) GetServerMode() string {
	return c.serverMode
}

// GetOriginalsPath returns the originals.
func (c *Config) GetOriginalsPath() string {
	return c.originalsPath
}

// GetImportPath returns the import directory.
func (c *Config) GetImportPath() string {
	return c.importPath
}

// GetExportPath returns the export directory.
func (c *Config) GetExportPath() string {
	return c.exportPath
}

// GetDarktableCli returns the darktable-cli binary file name.
func (c *Config) GetDarktableCli() string {
	return c.darktableCli
}

// GetDatabaseDriver returns the database driver name.
func (c *Config) GetDatabaseDriver() string {
	return c.databaseDriver
}

// GetDatabaseDsn returns the database data source name (DSN).
func (c *Config) GetDatabaseDsn() string {
	return c.databaseDsn
}

// GetCachePath returns the path to the cache.
func (c *Config) GetCachePath() string {
	return c.cachePath
}

// GetThumbnailsPath returns the path to the cached thumbnails.
func (c *Config) GetThumbnailsPath() string {
	return c.GetCachePath() + "/thumbnails"
}

// GetAssetsPath returns the path to the assets.
func (c *Config) GetAssetsPath() string {
	return c.assetsPath
}

// GetTensorFlowModelPath returns the tensorflow model path.
func (c *Config) GetTensorFlowModelPath() string {
	return c.GetAssetsPath() + "/tensorflow"
}

// GetDatabasePath returns the database storage path (e.g. for SQLite or Bleve).
func (c *Config) GetDatabasePath() string {
	return c.GetAssetsPath() + "/database"
}

// GetServerAssetsPath returns the server assets path (public files, favicons, templates,...).
func (c *Config) GetServerAssetsPath() string {
	return c.GetAssetsPath() + "/server"
}

// GetTemplatesPath returns the server templates path.
func (c *Config) GetTemplatesPath() string {
	return c.GetServerAssetsPath() + "/templates"
}

// GetFaviconsPath returns the favicons path.
func (c *Config) GetFaviconsPath() string {
	return c.GetServerAssetsPath() + "/favicons"
}

// GetPublicPath returns the public server path (//server/assets/*).
func (c *Config) GetPublicPath() string {
	return c.GetServerAssetsPath() + "/public"
}

// GetPublicBuildPath returns the public build path (//server/assets/build/*).
func (c *Config) GetPublicBuildPath() string {
	return c.GetPublicPath() + "/build"
}

// GetDb gets a db connection. If it already is estabilished it will return that.
func (c *Config) GetDb() *gorm.DB {
	if c.db == nil {
		c.connectToDatabase()
	}

	return c.db
}

// MigrateDb will start a migration process.
func (c *Config) MigrateDb() {
	db := c.GetDb()

	db.AutoMigrate(&models.File{},
		&models.Photo{},
		&models.Tag{},
		&models.Album{},
		&models.Location{},
		&models.Camera{},
		&models.Lens{},
		&models.Country{})

	if !db.Dialect().HasIndex("photos", "photos_fulltext") {
		db.Exec("CREATE FULLTEXT INDEX photos_fulltext ON photos (photo_title, photo_description, photo_artist, photo_colors)")
	}
}

// GetClientConfig returns a loaded and set configuration entity.
func (c *Config) GetClientConfig() map[string]interface{} {
	db := c.GetDb()

	var cameras []*models.Camera

	type country struct {
		LocCountry     string
		LocCountryCode string
	}

	var countries []country

	db.Model(&models.Location{}).Select("DISTINCT loc_country_code, loc_country").Scan(&countries)

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	jsHash := fileHash(c.GetPublicBuildPath() + "/app.js")
	cssHash := fileHash(c.GetPublicBuildPath() + "/app.css")

	result := configValues{
		"appName":    c.GetAppName(),
		"appVersion": c.GetAppVersion(),
		"debug":      c.IsDebug(),
		"cameras":    cameras,
		"countries":  countries,
		"jsHash":     jsHash,
		"cssHash":    cssHash,
	}

	return result
}

// GetExpandedFilename returns the expanded format for a filename.
func GetExpandedFilename(filename string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	if filename == "" {
		panic("filename was empty")
	}

	if len(filename) > 2 && filename[:2] == "~/" {
		filename = filepath.Join(dir, filename[2:])
	}

	result, _ := filepath.Abs(filename)

	return result
}
