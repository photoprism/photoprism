package context

import (
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/tidb"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Context struct {
	db     *gorm.DB
	config *Config
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

func NewContext(ctx *cli.Context) *Context {
	initLogger(ctx.GlobalBool("debug"))

	c := &Context{config: NewConfig(ctx)}

	log.SetLevel(c.LogLevel())

	return c
}

// CreateDirectories creates all the folders that photoprism needs. These are:
// OriginalsPath
// ThumbnailsPath
// ImportPath
// ExportPath
func (c *Context) CreateDirectories() error {
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
func (c *Context) connectToDatabase() error {
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
func (c *Context) Name() string {
	return c.config.Name
}

// Version returns the application version.
func (c *Context) Version() string {
	return c.config.Version
}

// Copyright returns the application copyright.
func (c *Context) Copyright() string {
	return c.config.Copyright
}

// Debug returns true if Debug mode is on.
func (c *Context) Debug() bool {
	return c.config.Debug
}

// ReadOnly returns true if photo directories are write protected.
func (c *Context) ReadOnly() bool {
	return c.config.ReadOnly
}

// LogLevel returns the logrus log level.
func (c *Context) LogLevel() log.Level {
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
func (c *Context) ConfigFile() string {
	return c.config.ConfigFile
}

// SqlServerHost returns the built-in SQL server host name or IP address (empty for all interfaces).
func (c *Context) SqlServerHost() string {
	return c.config.SqlServerHost
}

// SqlServerPort returns the built-in SQL server port.
func (c *Context) SqlServerPort() uint {
	return c.config.SqlServerPort
}

// SqlServerPath returns the database storage path for TiDB.
func (c *Context) SqlServerPath() string {
	if c.config.SqlServerPath != "" {
		return c.config.SqlServerPath
	}

	return c.ServerPath() + "/database"
}

// SqlServerPassword returns the password for the built-in database server.
func (c *Context) SqlServerPassword() string {
	return c.config.SqlServerPassword
}

// HttpServerHost returns the built-in HTTP server host name or IP address (empty for all interfaces).
func (c *Context) HttpServerHost() string {
	if c.config.HttpServerHost == "" {
		return "0.0.0.0"
	}

	return c.config.HttpServerHost
}

// HttpServerPort returns the built-in HTTP server port.
func (c *Context) HttpServerPort() int {
	return c.config.HttpServerPort
}

// HttpServerMode returns the server mode.
func (c *Context) HttpServerMode() string {
	return c.config.HttpServerMode
}

// HttpServerPassword returns the password for the user interface (optional).
func (c *Context) HttpServerPassword() string {
	return c.config.HttpServerPassword
}

// OriginalsPath returns the originals.
func (c *Context) OriginalsPath() string {
	return c.config.OriginalsPath
}

// ImportPath returns the import directory.
func (c *Context) ImportPath() string {
	return c.config.ImportPath
}

// ExportPath returns the export directory.
func (c *Context) ExportPath() string {
	return c.config.ExportPath
}

// DarktableCli returns the darktable-cli binary file name.
func (c *Context) DarktableCli() string {
	if c.config.DarktableCli == "" {
		return "/usr/bin/darktable-cli"
	}
	return c.config.DarktableCli
}

// DatabaseDriver returns the database driver name.
func (c *Context) DatabaseDriver() string {
	if c.config.DatabaseDriver == "" {
		return DbTiDB
	}

	return c.config.DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Context) DatabaseDsn() string {
	if c.config.DatabaseDsn == "" {
		return "root:photoprism@tcp(localhost:4000)/photoprism?parseTime=true"
	}

	return c.config.DatabaseDsn
}

// CachePath returns the path to the cache.
func (c *Context) CachePath() string {
	return c.config.CachePath
}

// ThumbnailsPath returns the path to the cached thumbnails.
func (c *Context) ThumbnailsPath() string {
	return c.CachePath() + "/thumbnails"
}

// AssetsPath returns the path to the assets.
func (c *Context) AssetsPath() string {
	return c.config.AssetsPath
}

// TensorFlowModelPath returns the tensorflow model path.
func (c *Context) TensorFlowModelPath() string {
	return c.AssetsPath() + "/tensorflow"
}

// ServerPath returns the server assets path (public files, favicons, templates,...).
func (c *Context) ServerPath() string {
	return c.AssetsPath() + "/server"
}

// HttpTemplatesPath returns the server templates path.
func (c *Context) HttpTemplatesPath() string {
	return c.ServerPath() + "/templates"
}

// HttpFaviconsPath returns the favicons path.
func (c *Context) HttpFaviconsPath() string {
	return c.HttpPublicPath() + "/favicons"
}

// HttpPublicPath returns the public server path (//server/assets/*).
func (c *Context) HttpPublicPath() string {
	return c.ServerPath() + "/public"
}

// HttpPublicBuildPath returns the public build path (//server/assets/build/*).
func (c *Context) HttpPublicBuildPath() string {
	return c.HttpPublicPath() + "/build"
}

// Db returns the db connection.
func (c *Context) Db() *gorm.DB {
	if c.db == nil {
		if err := c.connectToDatabase(); err != nil {
			log.Fatal(err)
		}
	}

	return c.db
}

// CloseDb closes the db connection (if any).
func (c *Context) CloseDb() error {
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
func (c *Context) MigrateDb() {
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
func (c *Context) ClientConfig() ClientConfig {
	db := c.Db()

	var cameras []*models.Camera

	type country struct {
		LocCountry     string
		LocCountryCode string
	}

	var countries []country

	db.Model(&models.Location{}).Select("DISTINCT loc_country_code, loc_country").Scan(&countries)

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	jsHash := fsutil.Hash(c.HttpPublicBuildPath() + "/app.js")
	cssHash := fsutil.Hash(c.HttpPublicBuildPath() + "/app.css")

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

func (c *Context) Shutdown() {
	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
		log.Error(err)
	}
}
