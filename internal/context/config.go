package context

import (
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/photoprism/photoprism/internal/frontend"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/tidb"
	"github.com/urfave/cli"
)

const (
	DbTiDB  = "tidb"
	DbMySQL = "mysql"
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
	dbServerIP     string
	dbServerPort   uint
	dbServerPath   string
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

// NewConfig() creates a new configuration entity by using two methods:
//
// 1. SetValuesFromFile: This will initialize values from a yaml config file.
//
// 2. SetValuesFromCliContext: Which comes after SetValuesFromFile and overrides
//    any previous values giving an option two override file configs through the CLI.
func NewConfig(ctx *cli.Context) *Config {
	c := &Config{}
	c.appName = ctx.App.Name
	c.appCopyright = ctx.App.Copyright
	c.appVersion = ctx.App.Version
	c.SetValuesFromFile(fsutil.ExpandedFilename(ctx.GlobalString("config-file")))
	c.SetValuesFromCliContext(ctx)

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

	if dbServerIP, err := yamlConfig.Get("db-host"); err == nil {
		c.dbServerIP = dbServerIP
	}

	if dbServerPort, err := yamlConfig.GetInt("db-port"); err == nil {
		c.dbServerPort = uint(dbServerPort)
	}

	if dbServerPath, err := yamlConfig.Get("db-path"); err == nil {
		c.dbServerPath = dbServerPath
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
		c.assetsPath = fsutil.ExpandedFilename(assetsPath)
	}

	if cachePath, err := yamlConfig.Get("cache-path"); err == nil {
		c.cachePath = fsutil.ExpandedFilename(cachePath)
	}

	if originalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.originalsPath = fsutil.ExpandedFilename(originalsPath)
	}

	if importPath, err := yamlConfig.Get("import-path"); err == nil {
		c.importPath = fsutil.ExpandedFilename(importPath)
	}

	if exportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.exportPath = fsutil.ExpandedFilename(exportPath)
	}

	if darktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.darktableCli = fsutil.ExpandedFilename(darktableCli)
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
func (c *Config) SetValuesFromCliContext(ctx *cli.Context) error {
	if ctx.GlobalBool("debug") {
		c.debug = ctx.GlobalBool("debug")
	}

	if ctx.GlobalIsSet("assets-path") || c.assetsPath == "" {
		c.assetsPath = fsutil.ExpandedFilename(ctx.GlobalString("assets-path"))
	}

	if ctx.GlobalIsSet("cache-path") || c.cachePath == "" {
		c.cachePath = fsutil.ExpandedFilename(ctx.GlobalString("cache-path"))
	}

	if ctx.GlobalIsSet("originals-path") || c.originalsPath == "" {
		c.originalsPath = fsutil.ExpandedFilename(ctx.GlobalString("originals-path"))
	}

	if ctx.GlobalIsSet("import-path") || c.importPath == "" {
		c.importPath = fsutil.ExpandedFilename(ctx.GlobalString("import-path"))
	}

	if ctx.GlobalIsSet("export-path") || c.exportPath == "" {
		c.exportPath = fsutil.ExpandedFilename(ctx.GlobalString("export-path"))
	}

	if ctx.GlobalIsSet("darktable-cli") || c.darktableCli == "" {
		c.darktableCli = fsutil.ExpandedFilename(ctx.GlobalString("darktable-cli"))
	}

	if ctx.GlobalIsSet("database-driver") || c.databaseDriver == "" {
		c.databaseDriver = ctx.GlobalString("database-driver")
	}

	if ctx.GlobalIsSet("database-dsn") || c.databaseDsn == "" {
		c.databaseDsn = ctx.GlobalString("database-dsn")
	}

	if ctx.IsSet("db-host") || c.dbServerIP == "" {
		c.dbServerIP = ctx.String("db-host")
	}

	if ctx.IsSet("db-port") || c.dbServerPort == 0 {
		c.dbServerPort = ctx.Uint("db-port")
	}

	if ctx.IsSet("db-path") || c.dbServerPath == "" {
		c.dbServerPath = ctx.String("db-path")
	}

	if ctx.IsSet("server-host") || c.serverIP == "" {
		c.serverIP = ctx.String("server-host")
	}

	if ctx.IsSet("server-port") || c.serverPort == 0 {
		c.serverPort = ctx.Int("server-port")
	}

	if ctx.IsSet("server-mode") || c.serverMode == "" {
		c.serverMode = ctx.String("server-mode")
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

// connectToDatabase establishes a database connection.
// When used with the tidb driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase() error {
	dbDriver := c.GetDatabaseDriver()
	dbDsn := c.GetDatabaseDsn()

	isTiDB := false
	initSuccess := false

	if dbDriver == DbTiDB {
		isTiDB = true
		dbDriver = DbMySQL
	}

	db, err := gorm.Open(dbDriver, dbDsn)

	if err != nil || db == nil {
		if isTiDB {
			go tidb.Start(c.GetDatabasePath(), 4000, "", c.IsDebug())
		}

		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			if isTiDB && !initSuccess {
				err = tidb.InitDatabase(4000)

				if err != nil {
					log.Println(err)
				} else {
					initSuccess = true
				}
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

// DbServerIP returns the database server IP address (empty for all).
func (c *Config) DbServerIP() string {
	return c.dbServerIP
}

// DbServerPort returns the database server port.
func (c *Config) DbServerPort() uint {
	return c.dbServerPort
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

// GetDb returns the db connection.
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
}

// GetClientConfig returns a loaded and set configuration entity.
func (c *Config) GetClientConfig() frontend.Config {
	db := c.GetDb()

	var cameras []*models.Camera

	type country struct {
		LocCountry     string
		LocCountryCode string
	}

	var countries []country

	db.Model(&models.Location{}).Select("DISTINCT loc_country_code, loc_country").Scan(&countries)

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	jsHash := fsutil.Hash(c.GetPublicBuildPath() + "/app.js")
	cssHash := fsutil.Hash(c.GetPublicBuildPath() + "/app.css")

	result := frontend.Config{
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
