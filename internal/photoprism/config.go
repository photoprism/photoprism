package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	. "github.com/photoprism/photoprism/internal/models"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

type Config struct {
	Debug          bool
	ConfigFile     string
	ServerIP       string
	ServerPort     int
	ServerMode     string
	AssetsPath     string
	ThumbnailsPath string
	OriginalsPath  string
	ImportPath     string
	ExportPath     string
	DarktableCli   string
	DatabaseDriver string
	DatabaseDsn    string
	db             *gorm.DB
}

type ConfigValues map[string]interface{}

func NewConfig(context *cli.Context) *Config {
	c := &Config{}
	c.SetValuesFromFile(GetExpandedFilename(context.GlobalString("config-file")))
	c.SetValuesFromCliContext(context)

	return c
}

func (c *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	c.ConfigFile = fileName

	if debug, err := yamlConfig.GetBool("debug"); err == nil {
		c.Debug = debug
	}

	if serverIP, err := yamlConfig.Get("server-host"); err == nil {
		c.ServerIP = serverIP
	}

	if serverPort, err := yamlConfig.GetInt("server-port"); err == nil {
		c.ServerPort = int(serverPort)
	}

	if serverMode, err := yamlConfig.Get("server-mode"); err == nil {
		c.ServerMode = serverMode
	}

	if assetsPath, err := yamlConfig.Get("assets-path"); err == nil {
		c.AssetsPath = GetExpandedFilename(assetsPath)
	}

	if thumbnailsPath, err := yamlConfig.Get("thumbnails-path"); err == nil {
		c.ThumbnailsPath = GetExpandedFilename(thumbnailsPath)
	}

	if originalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.OriginalsPath = GetExpandedFilename(originalsPath)
	}

	if importPath, err := yamlConfig.Get("import-path"); err == nil {
		c.ImportPath = GetExpandedFilename(importPath)
	}

	if exportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.ExportPath = GetExpandedFilename(exportPath)
	}

	if darktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.DarktableCli = GetExpandedFilename(darktableCli)
	}

	if databaseDriver, err := yamlConfig.Get("database-driver"); err == nil {
		c.DatabaseDriver = databaseDriver
	}

	if databaseDsn, err := yamlConfig.Get("database-dsn"); err == nil {
		c.DatabaseDsn = databaseDsn
	}

	return nil
}

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

func (c *Config) ConnectToDatabase() error {
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

func (c *Config) GetTensorFlowModelPath() string {
	return c.AssetsPath + "/tensorflow"
}

func (c *Config) GetDb() *gorm.DB {
	if c.db == nil {
		c.ConnectToDatabase()
	}

	return c.db
}

func (c *Config) MigrateDb() {
	db := c.GetDb()

	db.AutoMigrate(&File{}, &Photo{}, &Tag{}, &Album{}, &Location{}, &Camera{})

	if !db.Dialect().HasIndex("photos", "photos_fulltext") {
		db.Exec("CREATE FULLTEXT INDEX photos_fulltext ON photos (photo_title, photo_description, photo_artist, photo_colors)")
	}
}

func (c *Config) GetClientConfig() ConfigValues {
	db := c.GetDb()

	var cameras []*Camera
	// var countries map[string]string

	db.Where("deleted_at IS NULL").Limit(1000).Order("camera_model").Find(&cameras)

	result := ConfigValues{
		"title":   "PhotoPrism",
		"debug":   c.Debug,
		"cameras": cameras,
	}

	return result
}
