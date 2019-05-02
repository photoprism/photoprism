package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/frontend"
	"github.com/photoprism/photoprism/internal/fsutil"
	"github.com/photoprism/photoprism/internal/models"
)

const (
	DataURL    = "https://dl.photoprism.org/fixtures/test.zip"
	DataHash   = "1a59b358b80221ab3e76efb683ad72402f0b0844"
	ConfigFile = "../../configs/photoprism.yml"
)

var DarktableCli = "/usr/bin/darktable-cli"
var DataZip = "/tmp/photoprism/testdata.zip"
var AssetsPath = fsutil.ExpandedFilename("../../assets")
var DataPath = AssetsPath + "/testdata"
var CachePath = fsutil.ExpandedFilename(DataPath + "/cache")
var OriginalsPath = fsutil.ExpandedFilename(DataPath + "/originals")
var ImportPath = fsutil.ExpandedFilename(DataPath + "/import")
var ExportPath = fsutil.ExpandedFilename(DataPath + "/export")
var DatabaseDriver = "mysql"
var DatabaseDsn = "photoprism:photoprism@tcp(database:3306)/photoprism?parseTime=true"

func init() {
	conf := NewConfig()
	conf.MigrateDb()
}

type Config struct {
	db *gorm.DB
}

func (c *Config) RemoveTestData(t *testing.T) {
	os.RemoveAll(c.ImportPath())
	os.RemoveAll(c.ExportPath())
	os.RemoveAll(c.OriginalsPath())
	os.RemoveAll(c.CachePath())
}

func (c *Config) DownloadTestData(t *testing.T) {
	if fsutil.Exists(DataZip) {
		hash := fsutil.Hash(DataZip)

		if hash != DataHash {
			os.Remove(DataZip)
			t.Logf("Removed outdated test data zip file (fingerprint %s)\n", hash)
		}
	}

	if !fsutil.Exists(DataZip) {
		fmt.Printf("Downloading latest test data zip file from %s\n", DataURL)

		if err := fsutil.Download(DataZip, DataURL); err != nil {
			fmt.Printf("Download failed: %s\n", err.Error())
		}
	}
}

func (c *Config) UnzipTestData(t *testing.T) {
	if _, err := fsutil.Unzip(DataZip, DataPath); err != nil {
		t.Logf("Could not unzip test data: %s\n", err.Error())
	}
}

func (c *Config) InitializeTestData(t *testing.T) {
	t.Log("Initializing test data")

	c.RemoveTestData(t)

	c.DownloadTestData(t)

	c.UnzipTestData(t)
}

func NewConfig() *Config {
	return &Config{}
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

// connectToDatabase estabilishes a connection to a database given a driver.
// It tries to do this 12 times with a 5 second sleep intervall in between.
func (c *Config) connectToDatabase() error {
	dbDriver := c.DatabaseDriver()
	dbDsn := c.DatabaseDsn()

	db, err := gorm.Open(dbDriver, dbDsn)

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(dbDriver, dbDsn)

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

// AppName returns the application name.
func (c *Config) AppName() string {
	return "PhotoPrism"
}

// AppVersion returns the application version.
func (c *Config) AppVersion() string {
	return "DEVELOPMENT"
}

// AppCopyright returns the application copyright.
func (c *Config) AppCopyright() string {
	return "The PhotoPrism contributors <hello@photoprism.org>"
}

// Debug returns true if debug mode is on.
func (c *Config) Debug() bool {
	return false
}

// LogLevel returns the logrus log level.
func (c *Config) LogLevel() log.Level {
	return log.DebugLevel
}

// ConfigFile returns the config file name.
func (c *Config) ConfigFile() string {
	return ConfigFile
}

// HttpServerHost returns the server IP address (empty for all).
func (c *Config) HttpServerHost() string {
	return "127.0.0.1"
}

// HttpServerPort returns the server port.
func (c *Config) HttpServerPort() int {
	return 80
}

// SqlServerHost returns the database server IP address (empty for all).
func (c *Config) SqlServerHost() string {
	return "127.0.0.1"
}

// SqlServerPort returns the database server port.
func (c *Config) SqlServerPort() uint {
	return 4001
}

// SqlServerPassword returns the password for the built-in database server.
func (c *Config) SqlServerPassword() string {
	return "photoprism"
}

// HttpServerMode returns the server mode.
func (c *Config) HttpServerMode() string {
	return "test"
}

// HttpServerPassword returns the password for the Web UI.
func (c *Config) HttpServerPassword() string {
	return ""
}

// OriginalsPath returns the originals.
func (c *Config) OriginalsPath() string {
	return OriginalsPath
}

// ImportPath returns the import directory.
func (c *Config) ImportPath() string {
	return ImportPath
}

// ExportPath returns the export directory.
func (c *Config) ExportPath() string {
	return ExportPath
}

// DarktableCli returns the darktable-cli binary file name.
func (c *Config) DarktableCli() string {
	return DarktableCli
}

// DatabaseDriver returns the database driver name.
func (c *Config) DatabaseDriver() string {
	return DatabaseDriver
}

// DatabaseDsn returns the database data source name (DSN).
func (c *Config) DatabaseDsn() string {
	return DatabaseDsn
}

// CachePath returns the path to the cache.
func (c *Config) CachePath() string {
	return CachePath
}

// ThumbnailsPath returns the path to the cached thumbnails.
func (c *Config) ThumbnailsPath() string {
	return c.CachePath() + "/thumbnails"
}

// AssetsPath returns the path to the assets.
func (c *Config) AssetsPath() string {
	return AssetsPath
}

// TensorFlowModelPath returns the tensorflow model path.
func (c *Config) TensorFlowModelPath() string {
	return c.AssetsPath() + "/tensorflow"
}

// SqlServerPath returns the database storage path (e.g. for SQLite or Bleve).
func (c *Config) SqlServerPath() string {
	return c.ServerPath() + "/database"
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

// Db gets a db connection. If it already is estabilished it will return that.
func (c *Config) Db() *gorm.DB {
	if c.db == nil {
		c.connectToDatabase()
	}

	return c.db
}

// MigrateDb will start a migration process.
func (c *Config) MigrateDb() {
	db := c.Db()

	db.AutoMigrate(&models.File{},
		&models.Photo{},
		&models.Tag{},
		&models.Album{},
		&models.Location{},
		&models.Camera{},
		&models.Lens{},
		&models.Country{})
}

// ClientConfig returns a loaded and set configuration entity.
func (c *Config) ClientConfig() frontend.Config {
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

	result := frontend.Config{
		"appName":    c.AppName(),
		"appVersion": c.AppVersion(),
		"debug":      c.Debug(),
		"cameras":    cameras,
		"countries":  countries,
		"jsHash":     jsHash,
		"cssHash":    cssHash,
	}

	return result
}
