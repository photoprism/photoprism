package test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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
	os.RemoveAll(c.GetImportPath())
	os.RemoveAll(c.GetExportPath())
	os.RemoveAll(c.GetOriginalsPath())
	os.RemoveAll(c.GetCachePath())
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
	dbDriver := c.GetDatabaseDriver()
	dbDsn := c.GetDatabaseDsn()

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

// GetAppName returns the application name.
func (c *Config) GetAppName() string {
	return "PhotoPrism"
}

// GetAppVersion returns the application version.
func (c *Config) GetAppVersion() string {
	return "DEVELOPMENT"
}

// GetAppCopyright returns the application copyright.
func (c *Config) GetAppCopyright() string {
	return "The PhotoPrism contributors <hello@photoprism.org>"
}

// IsDebug returns true if debug mode is on.
func (c *Config) IsDebug() bool {
	return false
}

// GetConfigFile returns the config file name.
func (c *Config) GetConfigFile() string {
	return ConfigFile
}

// GetServerIP returns the server IP address (empty for all).
func (c *Config) GetServerIP() string {
	return "127.0.0.1"
}

// GetServerPort returns the server port.
func (c *Config) GetServerPort() int {
	return 80
}

// DbServerIP returns the database server IP address (empty for all).
func (c *Config) DbServerIP() string {
	return "127.0.0.1"
}

// DbServerPort returns the database server port.
func (c *Config) DbServerPort() uint {
	return 4001
}

// GetServerMode returns the server mode.
func (c *Config) GetServerMode() string {
	return "test"
}

// GetOriginalsPath returns the originals.
func (c *Config) GetOriginalsPath() string {
	return OriginalsPath
}

// GetImportPath returns the import directory.
func (c *Config) GetImportPath() string {
	return ImportPath
}

// GetExportPath returns the export directory.
func (c *Config) GetExportPath() string {
	return ExportPath
}

// GetDarktableCli returns the darktable-cli binary file name.
func (c *Config) GetDarktableCli() string {
	return DarktableCli
}

// GetDatabaseDriver returns the database driver name.
func (c *Config) GetDatabaseDriver() string {
	return DatabaseDriver
}

// GetDatabaseDsn returns the database data source name (DSN).
func (c *Config) GetDatabaseDsn() string {
	return DatabaseDsn
}

// GetCachePath returns the path to the cache.
func (c *Config) GetCachePath() string {
	return CachePath
}

// GetThumbnailsPath returns the path to the cached thumbnails.
func (c *Config) GetThumbnailsPath() string {
	return c.GetCachePath() + "/thumbnails"
}

// GetAssetsPath returns the path to the assets.
func (c *Config) GetAssetsPath() string {
	return AssetsPath
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
