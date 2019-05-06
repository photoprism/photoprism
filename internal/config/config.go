package config

import (
	"errors"
	"os"
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

func NewConfig(ctx *cli.Context) *Config {
	initLogger(ctx.GlobalBool("debug"))

	c := &Config{config: NewParams(ctx)}

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

	if err := os.MkdirAll(c.SqlServerPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.TensorFlowModelPath(), os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.HttpPublicBuildPath(), os.ModePerm); err != nil {
		return err
	}

	return nil
}

// connectToDatabase establishes a database connection.
// When used with the internal driver, it may create a new database server instance.
// It tries to do this 12 times with a 5 second sleep interval in between.
func (c *Config) connectToDatabase() error {
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

			go tidb.Start(c.SqlServerPath(), c.SqlServerPort(), c.SqlServerHost(), c.Debug())
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

// TestConfigFile returns the config file name.
func (c *Config) ConfigFile() string {
	return c.config.ConfigFile
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

	return c.ServerPath() + "/database"
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

// DarktableCli returns the darktable-cli binary file name.
func (c *Config) DarktableCli() string {
	if c.config.DarktableCli == "" {
		return "/usr/bin/darktable-cli"
	}
	return c.config.DarktableCli
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

// TensorFlowModelPath returns the tensorflow model path.
func (c *Config) TensorFlowModelPath() string {
	return c.AssetsPath() + "/tensorflow"
}

// ServerPath returns the server assets path (public files, favicons, templates,...).
func (c *Config) ServerPath() string {
	return c.AssetsPath() + "/server"
}

// HttpTemplatesPath returns the server templates path.
func (c *Config) HttpTemplatesPath() string {
	return c.ServerPath() + "/templates"
}

// HttpFaviconsPath returns the favicons path.
func (c *Config) HttpFaviconsPath() string {
	return c.HttpPublicPath() + "/favicons"
}

// HttpPublicPath returns the public server path (//server/assets/*).
func (c *Config) HttpPublicPath() string {
	return c.ServerPath() + "/public"
}

// HttpPublicBuildPath returns the public build path (//server/assets/build/*).
func (c *Config) HttpPublicBuildPath() string {
	return c.HttpPublicPath() + "/build"
}

// Db returns the db connection.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		if err := c.connectToDatabase(); err != nil {
			log.Fatal(err)
		}
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

	db.AutoMigrate(
		&models.File{},
		&models.Photo{},
		&models.Tag{},
		&models.PhotoTag{},
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

	jsHash := util.Hash(c.HttpPublicBuildPath() + "/app.js")
	cssHash := util.Hash(c.HttpPublicBuildPath() + "/app.css")

	result := ClientConfig{
		"name":      c.Name(),
		"version":   c.Version(),
		"copyright": c.Copyright(),
		"debug":     c.Debug(),
		"readonly":  c.ReadOnly(),
		"cameras":   cameras,
		"countries": countries,
		"jsHash":    jsHash,
		"cssHash":   cssHash,
	}

	return result
}

func (c *Config) Shutdown() {
	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}
}
