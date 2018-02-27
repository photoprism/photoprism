package photoprism

import (
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/urfave/cli"
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
		c.OriginalsPath = getExpandedFilename(OriginalsPath)
	}

	if ThumbnailsPath, err := yamlConfig.Get("thumbnails-path"); err == nil {
		c.ThumbnailsPath = getExpandedFilename(ThumbnailsPath)
	}

	if ImportPath, err := yamlConfig.Get("import-path"); err == nil {
		c.ImportPath = getExpandedFilename(ImportPath)
	}

	if ExportPath, err := yamlConfig.Get("export-path"); err == nil {
		c.ExportPath = getExpandedFilename(ExportPath)
	}

	if DarktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		c.DarktableCli = getExpandedFilename(DarktableCli)
	}

	return nil
}

func (c *Config) SetValuesFromCliContext(context *cli.Context) error {
	if context.IsSet("originals-path") {
		c.OriginalsPath = getExpandedFilename(context.String("originals-path"))
	}

	if context.IsSet("thumbnails-path") {
		c.ThumbnailsPath = getExpandedFilename(context.String("thumbnails-path"))
	}

	if context.IsSet("import-path") {
		c.ImportPath = getExpandedFilename(context.String("import-path"))
	}

	if context.IsSet("export-path") {
		c.ExportPath = getExpandedFilename(context.String("export-path"))
	}

	if context.IsSet("darktable-cli") {
		c.DarktableCli = getExpandedFilename(context.String("darktable-cli"))
	}

	return nil
}
