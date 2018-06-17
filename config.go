package photoprism

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/urfave/cli"
	"os"
	"path"
)

type Config struct {
	ConfigFile     string
	DarktableCli   string
	OriginalsPath  string
	ThumbnailsPath string
	ImportPath     string
	ExportPath     string
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

	return nil
}

func (c *Config) CreateDirectories() {
	os.MkdirAll(path.Dir(c.OriginalsPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ThumbnailsPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ImportPath), os.ModePerm)
	os.MkdirAll(path.Dir(c.ExportPath), os.ModePerm)
}
