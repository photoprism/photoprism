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

// TemplateSettings represents HTML template settings for the Web UI.
type TemplateSettings struct {
	Default string `json:"default" yaml:"default"`
}

// MapsSettings represents maps settings (for places).
type MapsSettings struct {
	Animate int    `json:"animate" yaml:"animate"`
	Style   string `json:"style" yaml:"style"`
}

// FeatureSettings represents feature flags, mainly for the Web UI.
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

// ImportSettings represents import settings.
type ImportSettings struct {
	Path string `json:"path" yaml:"path"`
	Move bool   `json:"move" yaml:"move"`
}

// IndexSettings represents indexing settings.
type IndexSettings struct {
	Path    string `json:"path" yaml:"path"`
	Convert bool   `json:"convert" yaml:"convert"`
	Rescan  bool   `json:"rescan" yaml:"rescan"`
	Stacks  bool   `json:"stacks" yaml:"stacks"`
}

// StackSettings represents settings for files that belong to the same photo.
type StackSettings struct {
	UUID bool `json:"uuid" yaml:"uuid"`
	Meta bool `json:"meta" yaml:"meta"`
	Name bool `json:"name" yaml:"name"`
}

// Settings represents user settings for Web UI, indexing, and import.
type Settings struct {
	Theme     string           `json:"theme" yaml:"theme"`
	Language  string           `json:"language" yaml:"language"`
	Templates TemplateSettings `json:"templates" yaml:"templates"`
	Maps      MapsSettings     `json:"maps" yaml:"maps"`
	Features  FeatureSettings  `json:"features" yaml:"features"`
	Import    ImportSettings   `json:"import" yaml:"import"`
	Index     IndexSettings    `json:"index" yaml:"index"`
	Stack     StackSettings    `json:"stack" yaml:"stack"`
}

// NewSettings creates a new Settings instance.
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
			Path:    "/",
			Rescan:  false,
			Convert: true,
			Stacks:  true,
		},
		Stack: StackSettings{
			UUID: true,
			Meta: true,
			Name: false,
		},
	}
}

// Propagate updates settings in other packages as needed.
func (s *Settings) Propagate() {
	i18n.SetLocale(s.Language)
}

// StackSequences tests if files should be stacked based on their file name prefix (sequential names).
func (s Settings) StackSequences() bool {
	return s.Index.Stacks && s.Stack.Name
}

// StackUUID tests if files should be stacked based on unique image or instance id.
func (s Settings) StackUUID() bool {
	return s.Index.Stacks && s.Stack.UUID
}

// StackMeta tests if files should be stacked based on their place and time metadata.
func (s Settings) StackMeta() bool {
	return s.Index.Stacks && s.Stack.Meta
}

// Load user settings from file.
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

// Save user settings to a file.
func (s *Settings) Save(fileName string) error {
	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	s.Propagate()

	if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// initSettings initializes user settings from a config file.
func (c *Config) initSettings() {
	c.settings = NewSettings()
	fileName := c.SettingsFile()

	if err := c.settings.Load(fileName); err == nil {
		log.Debugf("config: loaded settings from %s ", fileName)
	} else if err := c.settings.Save(fileName); err != nil {
		log.Errorf("failed creating %s: %s", fileName, err)
	} else {
		log.Debugf("config: created %s ", fileName)
	}

	i18n.SetDir(c.LocalesPath())

	c.settings.Propagate()
}

// Settings returns the current user settings.
func (c *Config) Settings() *Settings {
	if c.settings == nil {
		c.initSettings()
	}

	return c.settings
}
