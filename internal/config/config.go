package config

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/tidb"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Config struct {
	db     *gorm.DB
	config *Params
}

func initLogger(debug bool) {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func findExecutable(configBin, defaultBin string) (result string) {
	if configBin == "" {
		result = defaultBin
	} else {
		result = configBin
	}

	if path, err := exec.LookPath(result); err == nil {
		result = path
	}

	if !util.Exists(result) {
		result = ""
	}

	return result
}

func NewConfig(ctx *cli.Context) *Config {
	initLogger(ctx.GlobalBool("debug"))

	c := &Config{
		config: NewParams(ctx),
	}

	log.SetLevel(c.LogLevel())

	return c
}

// CreateDirectories creates all the folders that photoprism needs. These are:
// OriginalsPath
// ThumbnailsPath
// ImportPath
// ExportPath
func (c *Config) CreateDirectories() error {
	if err := os.MkdirAll(c.OriginalsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ImportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ExportPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ThumbnailsPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ResourcesPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.SqlServerPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.TensorFlowModelPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.HttpStaticBuildPath(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// connectToDatabase establishes a database connection.
// When used with the internal driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase(ctx context.Context) error {
	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	if dbDriver == "" {
		return errors.New("can't connect: database driver not specified")
	}

	if dbDsn == "" {
		return errors.New("can't connect: database DSN not specified")
	}

	isTiDB := false
	initSuccess := false

	if dbDriver == DbTiDB {
		isTiDB = true
		dbDriver = DbMySQL
	}

	db, err := gorm.Open(dbDriver, dbDsn)
	if err != nil || db == nil {
		if isTiDB {
			log.Infof("starting database server at %s:%d\n", c.SqlServerHost(), c.SqlServerPort())

			go tidb.Start(ctx, c.SqlServerPath(), c.SqlServerPort(), c.SqlServerHost(), c.Debug())
		}

		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

			if db != nil && err == nil {
				break
			}

			if isTiDB && !initSuccess {
				err = tidb.InitDatabase(c.SqlServerPort(), c.SqlServerPassword())

				if err != nil {
					log.Debug(err)
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

// Name returns the application name.
func (c *Config) Name() string {
	return c.config.Name
}

// Version returns the application version.
func (c *Config) Version() string {
	return c.config.Version
}

// Copyright returns the application copyright.
func (c *Config) Copyright() string {
	return c.config.Copyright
}

// Debug returns true if Debug mode is on.
func (c *Config) Debug() bool {
	return c.config.Debug
}

// ReadOnly returns true if photo directories are write protected.
func (c *Config) ReadOnly() bool {
	return c.config.ReadOnly
}

// LogLevel returns the logrus log level.
func (c *Config) LogLevel() log.Level {
	if c.Debug() {
		c.config.LogLevel = "debug"
	}

	if logLevel, err := log.ParseLevel(c.config.LogLevel); err == nil {
		return logLevel
	} else {
		return log.InfoLevel
	}
}

// ConfigFile returns the config file name.
func (c *Config) ConfigFile() string {
	return c.config.ConfigFile
}

// ConfigPath returns the config path.
func (c *Config) ConfigPath() string {
	if c.config.ConfigPath == "" {
		return c.AssetsPath() + "/config"
	}

	return c.config.ConfigPath
}

// SqlServerHost returns the built-in SQL server host name or IP address (empty for all interfaces).
func (c *Config) SqlServerHost() string {
	return c.config.SqlServerHost
}

// SqlServerPort returns the built-in SQL server port.
func (c *Config) SqlServerPort() uint {
	return c.config.SqlServerPort
}

// SqlServerPath returns the database storage path for TiDB.
func (c *Config) SqlServerPath() string {
	if c.config.SqlServerPath != "" {
		return c.config.SqlServerPath
	}

	return c.ResourcesPath() + "/database"
}

// SqlServerPassword returns the password for the built-in database server.
func (c *Config) SqlServerPassword() string {
	return c.config.SqlServerPassword
}

// HttpServerHost returns the built-in HTTP server host name or IP address (empty for all interfaces).
func (c *Config) HttpServerHost() string {
	if c.config.HttpServerHost == "" {
		return "0.0.0.0"
	}

	return c.config.HttpServerHost
}

// HttpServerPort returns the built-in HTTP server port.
func (c *Config) HttpServerPort() int {
	return c.config.HttpServerPort
}

// HttpServerMode returns the server mode.
func (c *Config) HttpServerMode() string {
	return c.config.HttpServerMode
}

// HttpServerPassword returns the password for the user interface (optional).
func (c *Config) HttpServerPassword() string {
	return c.config.HttpServerPassword
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	return c.config.OriginalsPath
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	return c.config.ImportPath
}

// ExportPath returns the export directory.
func (c *Config) ExportPath() string {
	return c.config.ExportPath
}

// SipsBin returns the sips binary file name.
func (c *Config) SipsBin() string {
	return findExecutable(c.config.SipsBin, "sips")
}

// DarktableBin returns the darktable-cli binary file name.
func (c *Config) DarktableBin() string {
	return findExecutable(c.config.DarktableBin, "darktable-cli")
}

// HeifConvertBin returns the heif-convert binary file name.
func (c *Config) HeifConvertBin() string {
	return findExecutable(c.config.HeifConvertBin, "heif-convert")
}

// ExifToolBin returns the exiftool binary file name.
func (c *Config) ExifToolBin() string {
	return findExecutable(c.config.ExifToolBin, "exiftool")
}

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	if c.config.DatabaseDriver == "" {
		return DbTiDB
	}

	return c.config.DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	if c.config.DatabaseDsn == "" {
		return "root:photoprism@tcp(localhost:4000)/photoprism?parseTime=true"
	}

	return c.config.DatabaseDsn
}

// CachePath returns the path to the cache.
func (c *Config) CachePath() string {
	return c.config.CachePath
}

// ThumbnailsPath returns the path to the cached thumbnails.
func (c *Config) ThumbnailsPath() string {
	return c.CachePath() + "/thumbnails"
}

// AssetsPath returns the path to the assets.
func (c *Config) AssetsPath() string {
	return c.config.AssetsPath
}

// ResourcesPath returns the path to the app resources like static files.
func (c *Config) ResourcesPath() string {
	if c.config.ResourcesPath == "" {
		return c.AssetsPath() + "/resources"
	}

	return c.config.ResourcesPath
}

// ExamplesPath returns the example files path.
func (c *Config) ExamplesPath() string {
	return c.ResourcesPath() + "/examples"
}

// TensorFlowModelPath returns the tensorflow model path.
func (c *Config) TensorFlowModelPath() string {
	return c.ResourcesPath() + "/nasnet"
}

// HttpTemplatesPath returns the server templates path.
func (c *Config) HttpTemplatesPath() string {
	return c.ResourcesPath() + "/templates"
}

// HttpFaviconsPath returns the favicons path.
func (c *Config) HttpFaviconsPath() string {
	return c.HttpStaticPath() + "/favicons"
}

// HttpStaticPath returns the static server assets path (//server/static/*).
func (c *Config) HttpStaticPath() string {
	return c.ResourcesPath() + "/static"
}

// HttpStaticBuildPath returns the static build path (//server/static/build/*).
func (c *Config) HttpStaticBuildPath() string {
	return c.HttpStaticPath() + "/build"
}

// Db returns the db connection.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		log.Fatal("database not initialised.")
	}

	return c.db
}

// CloseDb closes the db connection (if any).
func (c *Config) CloseDb() error {
	if c.db != nil {
		if err := c.db.Close(); err == nil {
			c.db = nil
		} else {
			return err
		}
	}

	return nil
}

// MigrateDb will start a migration process.
func (c *Config) MigrateDb() {
	db := c.Db()

	// db.LogMode(true)

	db.AutoMigrate(
		&models.File{},
		&models.Photo{},
		&models.Label{},
		&models.Category{},
		&models.PhotoLabel{},
		&models.Album{},
		&models.Location{},
		&models.Camera{},
		&models.Lens{},
		&models.Country{},
	)
}

// ClientConfig returns a loaded and set configuration entity.
func (c *Config) ClientConfig() ClientConfig {
	db := c.Db()

	var cameras []*models.Camera

	type country struct {
		LocCountry     string
		LocCountryCode string
	}

	var countries []country

	db.Model(&models.Location{}).Select("DISTINCT loc_country_code, loc_country").Scan(&countries)

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	jsHash := util.Hash(c.HttpStaticBuildPath() + "/app.js")
	cssHash := util.Hash(c.HttpStaticBuildPath() + "/app.css")

	result := ClientConfig{
		"name":       c.Name(),
		"version":    c.Version(),
		"copyright":  c.Copyright(),
		"debug":      c.Debug(),
		"readonly":   c.ReadOnly(),
		"cameras":    cameras,
		"countries":  countries,
		"thumbnails": Thumbnails,
		"jsHash":     jsHash,
		"cssHash":    cssHash,
	}

	return result
}

// Init initialises the Database.
func (c *Config) Init(ctx context.Context) error {
	return c.connectToDatabase(ctx)
}

func (c *Config) Shutdown() {
	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}
}
