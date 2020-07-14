package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/yaml.v2"
)

// SettingsHidden returns true if the user is not allowed to change settings.
func (c *Config) SettingsHidden() bool {
	return c.params.SettingsHidden
}

type TemplateSettings struct {
	Default string `json:"default" yaml:"default"`
}

type MapsSettings struct {
	Animate int    `json:"animate" yaml:"animate"`
	Style   string `json:"style" yaml:"style"`
}

type IndexSettings struct {
	Path      string `json:"path" yaml:"path"`
	Convert   bool   `json:"convert" yaml:"convert"`
	Rescan    bool   `json:"rescan" yaml:"rescan"`
	Sequences bool   `json:"sequences" yaml:"sequences"`
}

type ImportSettings struct {
	Path string `json:"path" yaml:"path"`
	Move bool   `json:"move" yaml:"move"`
}

type FeatureSettings struct {
	Upload   bool `json:"upload" yaml:"upload"`
	Download bool `json:"download" yaml:"download"`
	Archive  bool `json:"archive" yaml:"archive"`
	Private  bool `json:"private" yaml:"private"`
	Review   bool `json:"review" yaml:"review"`
	Files    bool `json:"files" yaml:"files"`
	Moments  bool `json:"moments" yaml:"moments"`
	Labels   bool `json:"labels" yaml:"labels"`
	Places   bool `json:"places" yaml:"places"`
	Edit     bool `json:"edit" yaml:"edit"`
	Share    bool `json:"share" yaml:"share"`
	Library  bool `json:"library" yaml:"library"`
	Import   bool `json:"import" yaml:"import"`
	Logs     bool `json:"logs" yaml:"logs"`
}

// Settings contains Web UI settings
type Settings struct {
	Theme     string           `json:"theme" yaml:"theme"`
	Language  string           `json:"language" yaml:"language"`
	Templates TemplateSettings `json:"templates" yaml:"templates"`
	Maps      MapsSettings     `json:"maps" yaml:"maps"`
	Features  FeatureSettings  `json:"features" yaml:"features"`
	Import    ImportSettings   `json:"import" yaml:"import"`
	Index     IndexSettings    `json:"index" yaml:"index"`
}

// NewSettings returns a empty Settings
func NewSettings() *Settings {
	return &Settings{
		Theme:    "default",
		Language: "en",
		Templates: TemplateSettings{
			Default: "index.tmpl",
		},
		Maps: MapsSettings{
			Animate: 0,
			Style:   "streets",
		},
		Features: FeatureSettings{
			Upload:   true,
			Download: true,
			Archive:  true,
			Review:   true,
			Private:  true,
			Files:    true,
			Moments:  true,
			Labels:   true,
			Places:   true,
			Edit:     true,
			Share:    true,
			Library:  true,
			Import:   true,
			Logs:     true,
		},
		Import: ImportSettings{
			Path: "/",
			Move: false,
		},
		Index: IndexSettings{
			Path:      "/",
			Rescan:    false,
			Convert:   true,
			Sequences: true,
		},
	}
}

// Propagate updates settings in other packages as needed.
func (s Settings) Propagate() {
	i18n.SetLocale(s.Language)
}

// Load uses a yaml config file to initiate the configuration entity.
func (s *Settings) Load(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: %s", txt.Quote(fileName))
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(yamlConfig, s); err != nil {
		return err
	}

	s.Propagate()

	return nil
}

// Save uses a yaml config file to initiate the configuration entity.
func (s *Settings) Save(fileName string) error {
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	s.Propagate()

	if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	s.Propagate()

	return nil
}

// initSettings initializes user settings from a config file.
func (c *Config) initSettings() {
	c.settings = NewSettings()
	p := c.SettingsFile()

	if err := c.settings.Load(p); err != nil {
		log.Info(err)
	}

	i18n.SetDir(c.LocalesPath())

	c.settings.Propagate()
}

// Settings returns the current user settings.
func (c *Config) Settings() *Settings {
	return c.settings
}
