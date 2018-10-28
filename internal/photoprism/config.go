package photoprism

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql" // Import gorm drivers
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Config provides a struct in which application configuration is stored.
type Config struct {
	Debug          bool `yaml:"debug"`
	ConfigFile     string
	ServerIP       string
	ServerPort     int    `yaml:"server-port"`
	ServerMode     string `yaml:"server-mode"`
	AssetsPath     string `yaml:"assets-path"`
	ThumbnailsPath string `yaml:"thumbnails-path"`
	OriginalsPath  string `yaml:"originals-path"`
	ImportPath     string `yaml:"import-path"`
	ExportPath     string `yaml:"export-path"`
	DarktableCli   string `yaml:"darktable-cli"`
	DatabaseDriver string `yaml:"database-driver"`
	DatabaseDsn    string `yaml:"database-dsn"`
	db             *gorm.DB
}

type configValues map[string]interface{}

// NewConfig creates a new configuration entity by using two methods.
// 1: SetValuesFromFile: This will initialize values from a yaml config file.
// 2: SetValuesFromCliContext: Which comes after SetValuesFromFile and overrides
// any previous values giving an option two override file configs through the CLI.
func NewConfig(context *cli.Context) *Config {
	c := &Config{}
	c.SetValuesFromFile(GetExpandedFilename(context.GlobalString("config-file")))
	c.SetValuesFromCliContext(context)

	return c
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (c *Config) SetValuesFromFile(fileName string) error {
	content, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(content, c)
	if err != nil {
		return err
	}

	c.ConfigFile = fileName

	return nil
}

// SetValuesFromCliContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Config) SetValuesFromCliContext(context *cli.Context) error {
	if context.GlobalBool("debug") {
		c.Debug = context.GlobalBool("debug")
	}

	if context.GlobalIsSet("assets-path") || c.AssetsPath == "" {
		c.AssetsPath = GetExpandedFilename(context.GlobalString("assets-path"))
	}

	if context.GlobalIsSet("thumbnails-path") || c.ThumbnailsPath == "" {
		c.ThumbnailsPath = GetExpandedFilename(context.GlobalString("thumbnails-path"))
	}

	if context.GlobalIsSet("originals-path") || c.OriginalsPath == "" {
		c.OriginalsPath = GetExpandedFilename(context.GlobalString("originals-path"))
	}

	if context.GlobalIsSet("import-path") || c.ImportPath == "" {
		c.ImportPath = GetExpandedFilename(context.GlobalString("import-path"))
	}

	if context.GlobalIsSet("export-path") || c.ExportPath == "" {
		c.ExportPath = GetExpandedFilename(context.GlobalString("export-path"))
	}

	if context.GlobalIsSet("darktable-cli") || c.DarktableCli == "" {
		c.DarktableCli = GetExpandedFilename(context.GlobalString("darktable-cli"))
	}

	if context.GlobalIsSet("database-driver") || c.DatabaseDriver == "" {
		c.DatabaseDriver = context.GlobalString("database-driver")
	}

	if context.GlobalIsSet("database-dsn") || c.DatabaseDsn == "" {
		c.DatabaseDsn = context.GlobalString("database-dsn")
	}

	return nil
}

// CreateDirectories creates all the folders that photoprism needs. These are:
// OriginalsPath
// ThumbnailsPath
// ImportPath
// ExportPath
func (c *Config) CreateDirectories() error {
	if err := os.MkdirAll(c.OriginalsPath, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ThumbnailsPath, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ImportPath, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(c.ExportPath, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// connectToDatabase estabilishes a connection to a database given a driver.
// It tries to do this 12 times with a 5 second sleep intervall in between.
func (c *Config) connectToDatabase() error {
	db, err := gorm.Open(c.DatabaseDriver, c.DatabaseDsn)

	if err != nil || db == nil {
		for i := 1; i <= 12; i++ {
			time.Sleep(5 * time.Second)

			db, err = gorm.Open(c.DatabaseDriver, c.DatabaseDsn)

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

// GetTensorFlowModelPath returns the tensorflow model path.
func (c *Config) GetTensorFlowModelPath() string {
	return c.AssetsPath + "/tensorflow"
}

// GetTemplatesPath returns the templates path.
func (c *Config) GetTemplatesPath() string {
	return c.AssetsPath + "/templates"
}

// GetFaviconsPath returns the favicons path.
func (c *Config) GetFaviconsPath() string {
	return c.AssetsPath + "/favicons"
}

// GetPublicPath returns the public path.
func (c *Config) GetPublicPath() string {
	return c.AssetsPath + "/public"
}

// GetPublicBuildPath returns the public build path.
func (c *Config) GetPublicBuildPath() string {
	return c.AssetsPath + "/build"
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
		"title":     "PhotoPrism",
		"debug":     c.Debug,
		"cameras":   cameras,
		"countries": countries,
		"jsHash":    jsHash,
		"cssHash":   cssHash,
	}

	return result
}
