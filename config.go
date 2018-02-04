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

func (config *Config) SetValuesFromFile(fileName string) error {
	yamlConfig, err := yaml.ReadFile(fileName)

	if err != nil {
		return err
	}

	config.ConfigFile = fileName

	if OriginalsPath, err := yamlConfig.Get("originals-path"); err == nil {
		config.OriginalsPath = GetExpandedFilename(OriginalsPath)
	}

	if ThumbnailsPath, err := yamlConfig.Get("thumbnails-path"); err == nil {
		config.ThumbnailsPath = GetExpandedFilename(ThumbnailsPath)
	}

	if ImportPath, err := yamlConfig.Get("import-path"); err == nil {
		config.ImportPath = GetExpandedFilename(ImportPath)
	}

	if ExportPath, err := yamlConfig.Get("export-path"); err == nil {
		config.ExportPath = GetExpandedFilename(ExportPath)
	}

	if DarktableCli, err := yamlConfig.Get("darktable-cli"); err == nil {
		config.DarktableCli = GetExpandedFilename(DarktableCli)
	}

	return nil
}

func (config *Config) SetValuesFromCliContext(c *cli.Context) error {
	if c.IsSet("originals-path") {
		config.OriginalsPath = GetExpandedFilename(c.String("originals-path"))
	}

	if c.IsSet("thumbnails-path") {
		config.ThumbnailsPath = GetExpandedFilename(c.String("thumbnails-path"))
	}

	if c.IsSet("import-path") {
		config.ImportPath = GetExpandedFilename(c.String("import-path"))
	}

	if c.IsSet("export-path") {
		config.ExportPath = GetExpandedFilename(c.String("export-path"))
	}

	if c.IsSet("darktable-cli") {
		config.DarktableCli = GetExpandedFilename(c.String("darktable-cli"))
	}

	return nil
}
