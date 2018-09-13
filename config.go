package photoprism

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/urfave/cli"
	"log"
	"os"
	"path"
	"time"
)

type Config struct {
	ConfigFile     string
	DarktableCli   string
	OriginalsPath  string
	ThumbnailsPath string
	ImportPath     string
	ExportPath     string
	DatabaseDriver string
	DatabaseDsn    string
	db             *gorm.DB
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	c.ConfigFile = fileName

	if OriginalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		c.OriginalsPath = GetExpandedFilename(OriginalsPath)
	}

	if ThumbnailsPath, err := yamlConfig.Get("thumbnails-path"); err == nil {
		c.ThumbnailsPath = GetExpandedFilename(ThumbnailsPath)
	}

	if ImportPath, err := yamlConfig.Get("import-path"); err == nil {
		c.ImportPath = GetExpandedFilename(ImportPath)
	}

	if ExportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.ExportPath = GetExpandedFilename(ExportPath)
	}

	if DarktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.DarktableCli = GetExpandedFilename(DarktableCli)
	}

	if DatabaseDriver, err := yamlConfig.Get("database-driver"); err == nil {
		c.DatabaseDriver = DatabaseDriver
	}

	if DatabaseDsn, err := yamlConfig.Get("database-dsn"); err == nil {
		c.DatabaseDsn = DatabaseDsn
	}

	return nil
}

func (c *Config) SetValuesFromCliContext(context *cli.Context) error {
	if context.IsSet("originals-path") {
		c.OriginalsPath = GetExpandedFilename(context.String("originals-path"))
	}

	if context.IsSet("thumbnails-path") {
		c.ThumbnailsPath = GetExpandedFilename(context.String("thumbnails-path"))
	}

	if context.IsSet("import-path") {
		c.ImportPath = GetExpandedFilename(context.String("import-path"))
	}

	if context.IsSet("export-path") {
		c.ExportPath = GetExpandedFilename(context.String("export-path"))
	}

	if context.IsSet("darktable-cli") {
		c.DarktableCli = GetExpandedFilename(context.String("darktable-cli"))
	}

	if context.IsSet("database-driver") {
		c.DatabaseDriver = context.String("database-driver")
	}

	if context.IsSet("database-dsn") {
		c.DatabaseDsn = context.String("database-dsn")
	}

	return nil
}

func (c *Config) CreateDirectories() {
	os.MkdirAll(path.Dir(c.OriginalsPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ThumbnailsPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ImportPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ExportPath), os.ModePerm)
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
